package ast

import "github.com/d5/tengo/compiler/source"

// Ident represents an identifier.
type Ident struct {
	Name    string
	NamePos source.Pos
}

func (e *Ident) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *Ident) Pos() source.Pos {
	return e.NamePos
}

// End returns the position of first character immediately after the node.
func (e *Ident) End() source.Pos {
	return source.Pos(int(e.NamePos) + len(e.Name))
}

func (e *Ident) String() string {
	if e != nil {
		return e.Name
	}

	return nullRep
}
