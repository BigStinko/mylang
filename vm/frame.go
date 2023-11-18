package vm

import (
	"mylang/code"
	"mylang/object"
)

type Frame struct {
	closure *object.Closure
	ip int
	basePointer int
}

func NewFrame(cl *object.Closure, bp int) *Frame {
	return &Frame{
		closure: cl,
		ip: -1,
		basePointer: bp,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.closure.Function.Instructions
}
