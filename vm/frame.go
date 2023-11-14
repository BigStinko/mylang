package vm

import (
	"mylang/code"
	"mylang/object"
)

type Frame struct {
	function *object.CompiledFunction
	ip int
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{function: fn, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.function.Instructions
}
