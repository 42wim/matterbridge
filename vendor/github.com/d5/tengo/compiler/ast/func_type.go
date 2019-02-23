package ast

import "github.com/d5/tengo/compiler/source"

// FuncType represents a function type definition.
type FuncType struct {
	FuncPos source.Pos
	Params  *IdentList
}

func (e *FuncType) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FuncType) Pos() source.Pos {
	return e.FuncPos
}

// End returns the position of first character immediately after the node.
func (e *FuncType) End() source.Pos {
	return e.Params.End()
}

func (e *FuncType) String() string {
	return "func" + e.Params.String()
}
