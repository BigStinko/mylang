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

// creates the vm with the given bytecode added as the main function.
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
		// when the compiler encounters a literal it replaces it with an
		// OpConstant instruction that tells the vm to push the constant 
		// from the constant pool to the stack
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
		
		// builds the array from the elements on top of the stack.
		// The first operand gives the number of elements in the array
		case code.OpArray:
			numElements := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp - numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		
		// builds the hash from the elements on top of the stack.
		// the first operand gives the number of keys/values in the hash
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

		// takes the index and left values from the top of the stack
		// for executing the index expression
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		
		// gets the number of arguments passed to the function from
		// the top of the stack from the first operand of OpCall
		case code.OpCall:
			numArgs := int(uint8(ins[ip + 1]))
			vm.currentFrame().ip += 1

			err := vm.callFunction(numArgs)
			if err != nil {
				return err
			}

		// ( returns from the current function by taking the frame off the
		// frame stack and setting the stack pointer to the value
		// it was before the function call. It then puts null on top of
		// the stack because there were no return values
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(NULL)
			if err != nil {
				return err
			}

		// returns from the current function by saving the return value
		// from the top of the stack, and going to the previous frame
		// sets the stack pointer to the location before the function
		// call
		case code.OpReturnValue:
			returnValue := vm.pop()  // get return value from top of the stack

			frame := vm.popFrame()  // return to the outer frame
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)  // replace function call with return value
			if err != nil {
				return err
			}

		// sets the instruction pointer to the location given by the operand
		// of the jump instruction
		case code.OpJump:
			pos := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			// need to subtract one since ip gets incremented at the end of the loop
			vm.currentFrame().ip = pos - 1 
		
		// gets the location of where to jump, and jumps if the value on top
		// of the stack is false
		case code.OpJumpFalse:
			pos := int(binary.BigEndian.Uint16(ins[ip + 1:]))
			vm.currentFrame().ip += 2  // advance the pointer to after the instruction

			condition := vm.pop()
			if !isTruthy(condition) {
				// if false move instruction pointer to instruction be(fore destination
				vm.currentFrame().ip = pos - 1
			}

		// takes the index to be associated with the global object from the 
		// operand of the instruction and assigns the object on top of the
		// stack to that index in the globals pool
		case code.OpSetGlobal:
			globalIndex := binary.BigEndian.Uint16(ins[ip + 1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()

		// puts the object assosiated with the index provided by the operand
		// on top of the stack
		case code.OpGetGlobal:
			globalIndex := binary.BigEndian.Uint16(ins[ip + 1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		// takes the index to be associated with the local object from the
		// operand of the instruction, and takes the value on top of the
		// stack and puts it at the location determined by the offset of 
		// the index from the frame's base pointer on the stack
		case code.OpSetLocal:
			localIndex := int(uint8(ins[ip + 1]))
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer + localIndex] = vm.pop()
		
		// puts the object from the stack in the location given by the
		// operand of the instruction on top of the stack
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
		
		// puts null on top of the stack
		case code.OpNull:
			err := vm.push(NULL)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// returns the value on top of the stack
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// returns the value that was most recently taken off the stack
func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

// puts the object on top of the stack and increments the stack pointer
// first checks if the stack has space
func (vm *VM) push(obj object.Object) error {
	if vm.sp >= STACKSIZE {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

// returns the value on top of the stack and decrements the stack pointer
func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

// returns the current frame of the vm
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex - 1]
}

// adds the given frame to the frame stack
func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

// removes the current frame from the frame stack
func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

// gets the function object from the stack and first checks that the
// number of arguments given matches the number of parameters to the
// function. Then creates a new frame for the function with the base
// pointer being the stack pointer minus the number of arguments to the
// function. The arguments now sit in the area of the stack given to the
// function on top of the base pointer. OpGetLocal instructions can now
// reference these arguments with their offset from the base pointer.(
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

// executes the binary operation based on the types of the values on top of the
// stack and the given operation
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
