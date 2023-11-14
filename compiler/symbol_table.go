package compiler

type SymbolScope string

type Symbol struct {
	Name string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store map[string]Symbol
	definitions int
}

const (
	GlobalScope SymbolScope = "GLOBAL"
)

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name: name,
		Index: s.definitions,
		Scope: GlobalScope,
	}

	s.store[name] = symbol
	s.definitions++
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	return obj, ok
}