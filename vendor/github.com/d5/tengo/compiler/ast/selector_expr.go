package ast

import "github.com/d5/tengo/compiler/source"

// SelectorExpr represents a selector expression.
type SelectorExpr struct {
	Expr Expr
	Sel  Expr
}

func (e *SelectorExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SelectorExpr) Pos() source.Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *SelectorExpr) End() source.Pos {
	return e.Sel.End()
}

func (e *SelectorExpr) String() string {
	return e.Expr.String() + "." + e.Sel.String()
}
