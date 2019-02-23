package ast

import "github.com/d5/tengo/compiler/source"

// ParenExpr represents a parenthesis wrapped expression.
type ParenExpr struct {
	Expr   Expr
	LParen source.Pos
	RParen source.Pos
}

func (e *ParenExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ParenExpr) Pos() source.Pos {
	return e.LParen
}

// End returns the position of first character immediately after the node.
func (e *ParenExpr) End() source.Pos {
	return e.RParen + 1
}

func (e *ParenExpr) String() string {
	return "(" + e.Expr.String() + ")"
}
