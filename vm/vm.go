package vm

import (
	"fmt"
	"encoding/binary"
	"mylang/code"
	"mylang/compiler"
	"mylang/object"
)

const (
	STACKSIZE = 2048
	GLOBALSIZE = 65536
	MAXFRAMES = 1024
)

var (
	TRUE = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL = &object.Null{}
)

type VM struct {
	constants []object.Object

	stack []object.Object
	sp int
	
	globals []object.Object

	frames []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MAXFRAMES)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,
		
		stack: make([]object.Object, STACKSIZE),
		sp: 0,

		globals: make([]object.Object, GLOBALSIZE),

		frames: frames,
		framesIndex: 1,
	}
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions()) - 1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			index := binary.BigEndian.Uint16(ins[ip + 1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}
		case code.OpAdd,
			code.OpSub,
			code.OpMul,
			code.OpDiv,
			code.OpEqual,
			code.OpNotEqual,
			code.OpGreaterThan:
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
		case code.OpArray:
			numElements := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp - numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp - numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return nil
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := int(uint8(ins[ip + 1]))
			vm.currentFrame().ip += 1

			err := vm.callFunction(numArgs)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(NULL)
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()  // get return value from top of the stack

			frame := vm.popFrame()  // return to the outer frame
			vm.sp = frame.basePointer - 1  // take the function call off the stack

			err := vm.push(returnValue)  // replace function call with return value
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			// need to subtract one since ip gets incremented at the end of the loop
			vm.currentFrame().ip = pos - 1 
		case code.OpJumpFalse:
			pos := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			vm.currentFrame().ip += 2  // advance the pointer to after the instruction

			condition := vm.pop()
			if !isTruthy(condition) {
				// if false move instruction pointer to instruction before destination
				vm.currentFrame().ip = pos - 1
			}
		case code.OpSetGlobal:
			globalIndex := binary.BigEndian.Uint16(ins[ip + 1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := binary.BigEndian.Uint16(ins[ip + 1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := int(uint8(ins[ip + 1]))
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer + localIndex] = vm.pop()
		case code.OpGetLocal:
			localIndex := int(uint8(ins[ip + 1]))
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer + localIndex])
			if err != nil {
				return err
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

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex - 1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) callFunction(numArgs int) error {
	fn, ok := vm.stack[vm.sp - 1 - numArgs].(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("calling non-function")
	}

	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParameters, numArgs)
	}

	frame := NewFrame(fn, vm.sp - numArgs)
	vm.pushFrame(frame)

	vm.sp = frame.basePointer + fn.NumLocals

	return nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
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

func (vm *VM) executeBinaryStringOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	switch op {
	case code.OpAdd:
		return vm.push(&object.String{Value: leftValue + rightValue})
	case code.OpEqual:
		return vm.push(getBoolObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(getBoolObject(leftValue != rightValue))
	default:
		return fmt.Errorf("unknown string operator: %d", op)
	}
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	maximum := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > maximum {
		return vm.push(NULL)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(NULL)
	}

	return vm.push(pair.Value)
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex - startIndex)

	for i := 0; i < endIndex - startIndex; i++ {
		elements[i] = vm.stack[startIndex + i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := 0; i < endIndex - startIndex; i += 2 {
		key := vm.stack[startIndex + i]
		value := vm.stack[startIndex + i + 1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
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
