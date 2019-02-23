package ast

import (
	"github.com/d5/tengo/compiler/source"
)

// ErrorExpr represents an error expression
type ErrorExpr struct {
	Expr     Expr
	ErrorPos source.Pos
	LParen   source.Pos
	RParen   source.Pos
}

func (e *ErrorExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ErrorExpr) Pos() source.Pos {
	return e.ErrorPos
}

// End returns the position of first character immediately after the node.
func (e *ErrorExpr) End() source.Pos {
	return e.RParen
}

func (e *ErrorExpr) String() string {
	return "error(" + e.Expr.String() + ")"
}
