package ast

import (
	"strings"

	"github.com/d5/tengo/compiler/source"
)

// IdentList represents a list of identifiers.
type IdentList struct {
	LParen  source.Pos
	VarArgs bool
	List    []*Ident
	RParen  source.Pos
}

// Pos returns the position of first character belonging to the node.
func (n *IdentList) Pos() source.Pos {
	if n.LParen.IsValid() {
		return n.LParen
	}

	if len(n.List) > 0 {
		return n.List[0].Pos()
	}

	return source.NoPos
}

// End returns the position of first character immediately after the node.
func (n *IdentList) End() source.Pos {
	if n.RParen.IsValid() {
		return n.RParen + 1
	}

	if l := len(n.List); l > 0 {
		return n.List[l-1].End()
	}

	return source.NoPos
}

// NumFields returns the number of fields.
func (n *IdentList) NumFields() int {
	if n == nil {
		return 0
	}

	return len(n.List)
}

func (n *IdentList) String() string {
	var list []string
	for i, e := range n.List {
		if n.VarArgs && i == len(n.List)-1 {
			list = append(list, "..."+e.String())
		} else {
			list = append(list, e.String())
		}
	}

	return "(" + strings.Join(list, ", ") + ")"
}
