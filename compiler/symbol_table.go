package compiler

type SymbolScope string

type Symbol struct {
	Name string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store map[string]Symbol
	definitions int
}

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// defines a symbol for the current symbol table
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name: name,
		Index: s.definitions,
	}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.definitions++
	return symbol
}

// checks all scopes for the given symbol
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		return obj, ok
	}
	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Name: name,
		Index: index,
		Scope: BuiltinScope,
	}

	s.store[name] = symbol
	return symbol
}
