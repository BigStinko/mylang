package vm

import (
	"encoding/binary"
	"fmt"
	"mylang/code"
	"mylang/compiler"
	"mylang/object"
)

const STACKSIZE = 2048

var (
	TRUE = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL = &object.Null{}
)

type VM struct {
	constants []object.Object
	instructions code.Instructions

	stack []object.Object
	sp int
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants: bytecode.Constants,

		stack: make([]object.Object, STACKSIZE),
		sp: 0,
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			index := binary.BigEndian.Uint16(vm.instructions[ip + 1:])
			ip += 2

			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeBinaryOperation(op) 
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		
		case code.OpNot:
			err := vm.executeNotOperator()
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(TRUE)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(FALSE)
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(binary.BigEndian.Uint16(vm.instructions[ip + 1:]))
			ip = pos - 1 // need to subtract one since ip gets incremented at the end of the loop
		case code.OpJumpFalse:
			pos := int(binary.BigEndian.Uint16(vm.instructions[ip + 1:]))
			ip += 2  // advance the pointer to after the instruction

			condition := vm.pop()
			if !isTruthy(condition) {
				// if false move instruction pointer to instruction before destination
				ip = pos - 1
			}
		case code.OpPop:
			vm.pop()
		case code.OpNull:
			err := vm.push(NULL)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= STACKSIZE {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case op == code.OpEqual:
		return vm.push(getBoolObject(left == right))
	case op == code.OpNotEqual:
		return vm.push(getBoolObject(left != right))
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		left.Type(), right.Type())
}

func (vm *VM) executeBinaryIntegerOperation(
	op code.Opcode,
	left, right object.Object,
)  error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		return vm.push(&object.Integer{Value:leftValue + rightValue})
	case code.OpSub:
		return vm.push(&object.Integer{Value:leftValue - rightValue})
	case code.OpMul:
		return vm.push(&object.Integer{Value:leftValue * rightValue})
	case code.OpDiv:
		return vm.push(&object.Integer{Value:leftValue / rightValue})
	case code.OpEqual:
		return vm.push(getBoolObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(getBoolObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(getBoolObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
}

func (vm *VM) executeNotOperator() error {
	operand := vm.pop()

	switch operand {
	case TRUE:
		return vm.push(FALSE)
	case FALSE:
		return vm.push(TRUE)
	case NULL:
		return vm.push(TRUE)
	default:
		return vm.push(FALSE)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	return vm.push(&object.Integer{Value: -operand.(*object.Integer).Value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func getBoolObject(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}
