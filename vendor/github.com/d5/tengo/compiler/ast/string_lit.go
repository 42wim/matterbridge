package ast

import "github.com/d5/tengo/compiler/source"

// StringLit represents a string literal.
type StringLit struct {
	Value    string
	ValuePos source.Pos
	Literal  string
}

func (e *StringLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *StringLit) Pos() source.Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *StringLit) End() source.Pos {
	return source.Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *StringLit) String() string {
	return e.Literal
}
