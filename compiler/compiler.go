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

	return &Compiler{
		constants: []object.Object{},
		symbolTable: NewSymbolTable(),
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
	
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

		c.emit(code.OpPop)

	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	
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

		if c.lastInstructionIs(code.OpPop) {
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
			if c.lastInstructionIs(code.OpPop) {
				c.removePop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)
	
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
		case token.GT:
			c.emit(code.OpGreaterThan)
		case token.EQ:
			c.emit(code.OpEqual)
		case token.NOT_EQ:
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unkown operator %s", node.Operator)
		}
	
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
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)
	
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}
		c.emit(code.OpCall)
	
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	
	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	
	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
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
	
	case *ast.FunctionLiteral:
		c.enterScope()

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
 
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{Instructions: instructions}
		c.emit(code.OpConstant, c.addConstant(compiledFn))
	
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.emit(code.OpGetGlobal, symbol.Index)
	}

	return nil
}

func (c *Compiler) MakeBytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants: c.constants,
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	var ins code.Instructions = code.Make(op, operands...)
	var pos int = c.addInstruction(ins)

	c.scopes[c.scopeIndex].beforeLastInstruction = c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	return pos

}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		beforeLastInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes) - 1]
	c.scopeIndex--

	return instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	newInstructionPos := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return newInstructionPos
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos + i] = newInstruction[i]
	}
}

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
