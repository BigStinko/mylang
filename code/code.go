package code

import (
	"fmt"
	"bytes"
	"encoding/binary"
)

type Instructions []byte

type Opcode byte

type Definition struct {
	Name string
	OperandWidths []int
}

const (
	OpConstant Opcode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpNot
	OpTrue
	OpFalse
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturn
	OpReturnValue
	OpJump
	OpJumpFalse
	OpSetGlobal
	OpGetGlobal
	OpSetLocal
	OpGetLocal
	OpGetFree
	OpClosure
	OpCurrentClosure
	OpGetBuiltin
	OpPop
	OpNull
)

var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant",       []int{2}},
	OpAdd:            {"OpAdd",            []int{}},
	OpSub:            {"OpSub",            []int{}},
	OpMul:            {"OpMul",            []int{}},
	OpDiv:            {"OpDiv",            []int{}},
	OpEqual:          {"OpEqual",          []int{}},
	OpNotEqual:       {"OpEqual",          []int{}},
	OpGreaterThan:    {"OpGreaterThan",    []int{}},
	OpMinus:          {"OpMinus",          []int{}},
	OpNot:            {"OpNot",            []int{}},
	OpTrue:           {"OpTrue",           []int{}},
	OpFalse:          {"OpFalse",          []int{}},
	OpArray:          {"OpArray",          []int{2}},
	OpHash:           {"OpHash",           []int{2}},
	OpIndex:          {"OpIndex",          []int{}},
	OpCall:           {"OpCall",           []int{1}},
	OpReturn:         {"OpReturn",         []int{}},
	OpReturnValue:    {"OpReturnValue",    []int{}},
	OpJump:           {"OpJump",           []int{2}},
	OpJumpFalse:      {"OpJumpFalse",      []int{2}},
	OpSetGlobal:      {"OpSetGlobal",      []int{2}},
	OpGetGlobal:      {"OpGetGlobal",      []int{2}},
	OpSetLocal:       {"OpSetLocal",       []int{1}},
	OpGetLocal:       {"OpGetLocal",       []int{1}},
	OpGetFree:        {"OpGetFree",        []int{1}},
	OpGetBuiltin:     {"OpGetBuiltin",     []int{1}},
	OpClosure:        {"OpClosure",        []int{2, 1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},
	OpPop:            {"OpPop",            []int{}},
	OpNull:           {"OpNull",           []int{}},
}

// returns a string representation of the list of instructions
func (ins Instructions) String() string {
	var out bytes.Buffer

	for i := 0; i < len(ins); {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

// returns a string representation of the instruction
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("Error: operand len%d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1]) 
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// Used to see if the given opCode is defined
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// takes an operation and operands and constructs an instruction.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}
	
	// determines the instruction length by adding the length of the operands
	// to the length of the operation
	instructionLength := 1
	for _, width := range def.OperandWidths {
		instructionLength += width
	}

	instruction := make([]byte, instructionLength)
	instruction[0] = byte(op)  // the first byte of an instruction is the opCode

	// starting from the second byte adds the operands to the instruction
	// with the most significant bit going to the lower location in the	array
	offset := 1 
	for i, operand := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		case 1:
			instruction[offset] = byte(operand)
		}
		offset += width
	}

	return instruction
}

// takes the definition of the given operation, and the given instruction.
// Using the definition of the operation, determines the amount of bytes
// necessary to pull from the instruction for the operands. If the operation 
// requires two operands each 2 bytes in width it will read the next two bytes 
// from the program, put it as the first operand, move the offset to after those 
// bytes and subsequently read the next two bytes for the second operand. 
// Also returns the amount of bytes read so the caller knows not to reread over
// those bytes.
// ReadOperands is the inverse of the Make function as it takes an instruction
// and an operation and returns the operands, whereas the make function takes
// an operation and operands and makes the instruction.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(binary.BigEndian.Uint16(ins[offset:]))
		case 1:
			operands[i] = int(uint8(ins[offset]))
		}

		offset += width
	}

	return operands, offset
}
