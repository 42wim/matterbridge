package ast

import "github.com/d5/tengo/compiler/source"

// BadExpr represents a bad expression.
type BadExpr struct {
	From source.Pos
	To   source.Pos
}

func (e *BadExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BadExpr) Pos() source.Pos {
	return e.From
}

// End returns the position of first character immediately after the node.
func (e *BadExpr) End() source.Pos {
	return e.To
}

func (e *BadExpr) String() string {
	return "<bad expression>"
}
