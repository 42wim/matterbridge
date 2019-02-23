package ast

import (
	"github.com/d5/tengo/compiler/source"
)

// ImmutableExpr represents an immutable expression
type ImmutableExpr struct {
	Expr     Expr
	ErrorPos source.Pos
	LParen   source.Pos
	RParen   source.Pos
}

func (e *ImmutableExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ImmutableExpr) Pos() source.Pos {
	return e.ErrorPos
}

// End returns the position of first character immediately after the node.
func (e *ImmutableExpr) End() source.Pos {
	return e.RParen
}

func (e *ImmutableExpr) String() string {
	return "immutable(" + e.Expr.String() + ")"
}
