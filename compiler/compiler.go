package compiler

import (
	"fmt"
	"sort"
	"mylang/ast"
	"mylang/code"
	"mylang/object"
	"mylang/token"
)

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes []CompilationScope
	scopeIndex int
}

type CompilationScope struct {
	instructions code.Instructions
	lastInstruction EmittedInstruction
	beforeLastInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants []object.Object
}

type EmittedInstruction struct {
	Opcode code.Opcode
	Position int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		beforeLastInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants: []object.Object{},
		symbolTable: symbolTable,
		scopes: []CompilationScope{mainScope},
		scopeIndex: 0,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	
	// defines a name in the symbolTable and returns a number
	// to refer to that name in the operand of the set operation.
	// the set operation tells the vm that the value on top of the
	// stack is to be associated with the given symbol index
	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	
	case *ast.AssignmentStatement:
		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Name.Value)
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		switch symbol.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobal, symbol.Index)
		case LocalScope:
			c.emit(code.OpSetLocal, symbol.Index)
		case FreeScope:
			c.emit(code.OpSetFree, symbol.Index)
		default:
			return fmt.Errorf("cannot redefine functions")
		}
	
	// the return value operation tells the vm that the return value
	// to return from the function is on top of the stack
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	// compiles the expression statement and takes the result off
	// the top of the stack with OpPop as the value is not being used
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

		c.emit(code.OpPop)

	// calls compile on every statement in the block
	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	
	// constructs an if statement using jumps by first putting the
	// result of the condition on the stack. Then a jump if false
	// opperation is added that jumps to a location if the top
	// of the stack is false then the instructions for the consequence
	// are compiled. After the consequence, an unconditional jump is 
	// added that jumps to the end of the if expression. Then the
	// alternative is compiled if it exists. The locations of the
	// jumps are added after the consequence and alternatives are
	// compiled to determine the length of the set of their instructions
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpFalsePos := c.emit(code.OpJumpFalse, 9999)

		err = c.Compile(node.Consequence)

		if err != nil {
			return err
		}

		if !c.lastInstructionIs(code.OpPop) {
			c.emit(code.OpNull)
		} else {
			c.removePop()
		}
		
		jumpPos := c.emit(code.OpJump, 9999)
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpFalsePos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if !c.lastInstructionIs(code.OpPop) {
				c.emit(code.OpNull)
			} else {
				c.removePop()
			}
		}


		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.WhileExpression:
		jumpPos := len(c.currentInstructions())
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		
		jumpFalsePos := c.emit(code.OpJumpFalse, 9999)

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}
		
		c.emit(code.OpJump, jumpPos)
		c.changeOperand(jumpFalsePos, len(c.currentInstructions()))
		c.emit(code.OpNull)
	
	case *ast.SwitchExpression:
		jumpPositions := []int{}

		for _, choice := range node.Cases {
			if choice.Default {
				continue
			}

			err := c.Compile(node.Value)
			if err != nil {
				return err
			}

			err = c.Compile(choice.Value)
			if err != nil {
				return err
			}

			c.emit(code.OpEqual)

			jumpFalsePos := c.emit(code.OpJumpFalse, 9999)

			err = c.Compile(choice.Body)
			if err != nil {
				return err
			}

			if !c.lastInstructionIs(code.OpPop) {
				c.emit(code.OpNull)
			} else {
				c.removePop()
			}

			jumpPositions = append(jumpPositions, c.emit(code.OpJump, 9999))
			c.changeOperand(jumpFalsePos, len(c.currentInstructions()))
		}

		for _, choice := range node.Cases {
			if choice.Default {
				err := c.Compile(choice.Body)
				if err != nil {
					return err
				}
				
				if !c.lastInstructionIs(code.OpPop) {
					c.emit(code.OpNull)
				} else {
					c.removePop()
				}
			}
		}

		var endPos int = len(c.currentInstructions())

		for _, pos := range jumpPositions {
			c.changeOperand(pos, endPos)
		}

	// compiles both sides of the infix expression. The infix operations
	// take the top two values of the stack for their operation, and puts
	// the result on top of the stack
	case *ast.InfixExpression:
		if node.Token.Type == token.LT {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Token.Type {
		case token.PLUS:
			c.emit(code.OpAdd)
		case token.MINUS:
			c.emit(code.OpSub)
		case token.ASTERISK:
			c.emit(code.OpMul)
		case token.SLASH:
			c.emit(code.OpDiv)
		case token.MODULO:
			c.emit(code.OpMod)
		case token.GT:
			c.emit(code.OpGreaterThan)
		case token.EQ:
			c.emit(code.OpEqual)
		case token.NOT_EQ:
			c.emit(code.OpNotEqual)
		case token.AND:
			c.emit(code.OpAnd)
		case token.OR:
			c.emit(code.OpOr)
		default:
			return fmt.Errorf("unkown operator %s", node.Operator)
		}
	
	// compiles the expression to the right of the operator.
	// prefix operations take the top of the stack for the operation
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Token.Type {
		case token.MINUS:
			c.emit(code.OpMinus)
		case token.BANG:
			c.emit(code.OpNot)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	
	case *ast.IndexExpression:
		err := c.Compile(node.Left)  // left object goes on the stack
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)  // index goes on the stack
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)
	
	// compiles the function literal and the arguments. The opcall
	// operation tells the vm that the function literal and the arguments
	// are on the stack. The operand for the call operation has 
	// the number of arguments on the stack
	case *ast.CallExpression:
		err := c.Compile(node.Function)  // function object goes on the stack
		if err != nil {
			return err
		}

		for _, a := range node.Arguments { // arguments go on the stack
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
	
	// puts the integer literal on the stack
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	
	// puts the float literal on to the stack
	case *ast.FloatLiteral:
		flt := &object.Float{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(flt))
	
	// puts the boolean object on the stack
	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	
	// puts the string literal on the stack
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	
	// compiles the elements of the literal. The array operation has
	// the number of elements that the vm needs to take of the stack
	// to build the array
	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	
	// just like OpArray, OpHash has the number values the vm has to
	// take off the stack to construct the hash
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {  // pairs go on the stack
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs) * 2)
	
	// starts a new scope for the instructions in the function
	// defines the parameters in the function scope, and compiles
	// the instructions of the body. To account for implicit returns,
	// looks at the last instruction to see if is an OpPop (meaning
	// the end of an expression statement) and replaces it with a
	// return value operation. If there is no implicit or explicit
	// return value then the OpReturn signifies no return value.
	// the compiled function object takes the instructions from the
	// scope created for the function, the number of symbols defined,
	// and the number of parameters to the function
	case *ast.FunctionLiteral:
		c.enterScope()

		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
			c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

			c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.definitions
		instructions := c.leaveScope()

		for _, symbol := range freeSymbols {
			c.loadSymbol(symbol)
		}

		compiledFn := &object.CompiledFunction{
			Instructions: instructions,
			NumLocals: numLocals,
			NumParameters: len(node.Parameters),
			Name: node.Name,
		}
		//fmt.Print(node.Name + "\n")
		//fmt.Print(instructions.String())

		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))
	

	// gets the symbol mapped to the identifier and emits the
	// get operation with the number representing the symbol
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.loadSymbol(symbol)
	}

	return nil
}

// returns the instructions and constants generated by the compiler
func (c *Compiler) MakeBytecode() *Bytecode {
	//fmt.Print(c.currentInstructions().String())
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants: c.constants,
	}
}

// makes the instruction for the given opcode and operands, and adds it to
// the current instructions. Then advances the last and before last
// instructions in the current scope, and returns the position of the
// location of the add instruction in the instructions list
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	var ins code.Instructions = code.Make(op, operands...)
	var pos int = c.addInstruction(ins)

	next := EmittedInstruction{Opcode: op, Position: pos}
	last := c.scopes[c.scopeIndex].lastInstruction

	c.scopes[c.scopeIndex].beforeLastInstruction = last
	c.scopes[c.scopeIndex].lastInstruction = next
	
	return pos
}

// creates a new scope and symbolTable and steps into it
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		beforeLastInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// saves the instructions from the current scope, and then decrements
// the scope and symbolTable. Returns the decremented instructions
func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes) - 1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	return instructions
}

// adds the instructions to the current scope and returns the position
// in the scope.
func (c *Compiler) addInstruction(ins []byte) int {
	newInstructionPos := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	return newInstructionPos
}

// replaces the instruction at the given position in the current instructions
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos + i] = newInstruction[i]
	}
}

// changes the operand at the given position by making a new instruction
// replacing the instruction
func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) removePop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	beforeLast := c.scopes[c.scopeIndex].beforeLastInstruction
	
	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = beforeLast
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}

func (c *Compiler) emptyStack() bool {
	switch c.scopes[c.scopeIndex].lastInstruction.Opcode {
	case code.OpSetGlobal,
		code.OpSetLocal,
		code.OpSetFree:
		return true
	default:
		return false
	}
}
