package vm

import (
	"mylang/code"
	"mylang/object"
)

type Frame struct {
	function *object.CompiledFunction
	ip int
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, bp int) *Frame {
	return &Frame{
		function: fn,
		ip: -1,
		basePointer: bp,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.function.Instructions
}
