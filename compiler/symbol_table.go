package compiler

type SymbolScope string

type Symbol struct {
	Name string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable
	FreeSymbols []Symbol

	store map[string]Symbol
	definitions int
}

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
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
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}

		free := s.defineFree(obj)
		return free, true
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

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{
		Name: name,
		Index: 0,
		Scope: FunctionScope,
	}

	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{
		Name: original.Name,
		Index: len(s.FreeSymbols) - 1,
		Scope: FreeScope,
	}

	s.store[original.Name] = symbol
	return symbol
}
