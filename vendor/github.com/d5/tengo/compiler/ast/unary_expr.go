package ast

import (
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
)

// UnaryExpr represents an unary operator expression.
type UnaryExpr struct {
	Expr     Expr
	Token    token.Token
	TokenPos source.Pos
}

func (e *UnaryExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *UnaryExpr) Pos() source.Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *UnaryExpr) End() source.Pos {
	return e.Expr.End()
}

func (e *UnaryExpr) String() string {
	return "(" + e.Token.String() + e.Expr.String() + ")"
}
