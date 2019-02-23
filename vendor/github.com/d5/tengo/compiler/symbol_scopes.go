package compiler

// SymbolScope represents a symbol scope.
type SymbolScope string

// List of symbol scopes
const (
	ScopeGlobal  SymbolScope = "GLOBAL"
	ScopeLocal               = "LOCAL"
	ScopeBuiltin             = "BUILTIN"
	ScopeFree                = "FREE"
)
