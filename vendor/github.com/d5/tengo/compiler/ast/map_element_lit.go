package ast

import "github.com/d5/tengo/compiler/source"

// MapElementLit represents a map element.
type MapElementLit struct {
	Key      string
	KeyPos   source.Pos
	ColonPos source.Pos
	Value    Expr
}

func (e *MapElementLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *MapElementLit) Pos() source.Pos {
	return e.KeyPos
}

// End returns the position of first character immediately after the node.
func (e *MapElementLit) End() source.Pos {
	return e.Value.End()
}

func (e *MapElementLit) String() string {
	return e.Key + ": " + e.Value.String()
}
