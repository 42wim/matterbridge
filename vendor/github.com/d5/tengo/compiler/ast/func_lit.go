package ast

import "github.com/d5/tengo/compiler/source"

// FuncLit represents a function literal.
type FuncLit struct {
	Type *FuncType
	Body *BlockStmt
}

func (e *FuncLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FuncLit) Pos() source.Pos {
	return e.Type.Pos()
}

// End returns the position of first character immediately after the node.
func (e *FuncLit) End() source.Pos {
	return e.Body.End()
}

func (e *FuncLit) String() string {
	return "func" + e.Type.Params.String() + " " + e.Body.String()
}
