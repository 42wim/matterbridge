package ast

import (
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
)

// ImportExpr represents an import expression
type ImportExpr struct {
	ModuleName string
	Token      token.Token
	TokenPos   source.Pos
}

func (e *ImportExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ImportExpr) Pos() source.Pos {
	return e.TokenPos
}

// End returns the position of first character immediately after the node.
func (e *ImportExpr) End() source.Pos {
	return source.Pos(int(e.TokenPos) + 10 + len(e.ModuleName)) // import("moduleName")
}

func (e *ImportExpr) String() string {
	return `import("` + e.ModuleName + `")"`
}
