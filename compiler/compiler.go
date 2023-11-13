package compiler

import (
	"fmt"
	"mylang/ast"
	"mylang/code"
	"mylang/object"
	"mylang/token"
)

type Compiler struct {
	instructions code.Instructions
	constants []object.Object

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
	return &Compiler{
		instructions: code.Instructions{},
		constants: []object.Object{},
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

		if c.lastInstruction.Opcode == code.OpPop {
			c.removePop()
		}
		
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpFalsePos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstruction.Opcode == code.OpPop {
				c.removePop()
			}
		}

		afterAlternativePos := len(c.instructions)
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
	
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	
	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	}

	return nil
}

func (c *Compiler) MakeBytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants: c.constants,
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	var ins code.Instructions = code.Make(op, operands...)
	var pos int = c.addInstruction(ins)

	c.beforeLastInstruction = c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	return pos

}

func (c *Compiler) addInstruction(ins []byte) int {
	newInstructionPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return newInstructionPos
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos + i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(pos int, operand int) {
	op := code.Opcode(c.instructions[pos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(pos, newInstruction)
}

func (c *Compiler) removePop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.beforeLastInstruction
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}
