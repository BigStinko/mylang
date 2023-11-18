package object

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"mylang/ast"
	"mylang/code"
)

type ObjectType string

type BuiltinFunction func(args ...Object) Object

const (
	NULL_OBJ = "NULL"
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ = "FLOAT"
	BOOLEAN_OBJ = "BOOLEAN"
	RUNE_OBJ = "RUNE"
	STRING_OBJ = "STRING"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ = "FUNCTION"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION"
	CLOSURE_OBJ = "CLOSURE"
	ERROR_OBJ = "ERROR"
	BUILTIN_OBJ = "BUILTIN"
	ARRAY_OBJ = "ARRAY"
	HASH_OBJ = "HASH"
	FILE_OBJ = "FILE"
)

// wrapper for values used by evaluator
type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}


type Null struct {}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string { return "null" }


type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }


type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string { return fmt.Sprintf("%f", f.Value) }


type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }


type Rune struct {
	Value rune
}

func (r *Rune) Type() ObjectType { return RUNE_OBJ }
func (r *Rune) Inspect() string { return string(r.Value) }


type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string { return s.Value }


type File struct {
	Path string
	Reader *bufio.Reader
	Writer *bufio.Writer
	Handle *os.File
}

func (f *File) Type() ObjectType { return FILE_OBJ }
func (f *File) Inspect() string { return fmt.Sprint(f.Path) }


type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, element := range a.Elements {
		elements = append(elements, element.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}


type HashKey struct {
	Type ObjectType
	Value uint64
}

type HashPair struct {
	Key Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}


func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// another layer of wrapper so when the object is encountered,
// the program knows when to return
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }


type Function struct {
	Parameters []*ast.Identifier
	Body *ast.BlockStatement
	Environment *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	parameters := []string{}
	for _, p := range f.Parameters {
		parameters = append(parameters, p.String())
	}

	out.WriteString("func")
	out.WriteString("(")
	out.WriteString(strings.Join(parameters, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}


type CompiledFunction struct {
	Instructions code.Instructions
	NumLocals int
	NumParameters int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}


type Closure struct {
	Function *CompiledFunction
	Free []Object
}

func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}


type Builtin struct {
	Function BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string { return "builtin function" }


type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string { return "ERROR: " + e.Message }
