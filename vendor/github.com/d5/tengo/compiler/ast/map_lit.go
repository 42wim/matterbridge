package ast

import (
	"strings"

	"github.com/d5/tengo/compiler/source"
)

// MapLit represents a map literal.
type MapLit struct {
	LBrace   source.Pos
	Elements []*MapElementLit
	RBrace   source.Pos
}

func (e *MapLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *MapLit) Pos() source.Pos {
	return e.LBrace
}

// End returns the position of first character immediately after the node.
func (e *MapLit) End() source.Pos {
	return e.RBrace + 1
}

func (e *MapLit) String() string {
	var elements []string
	for _, m := range e.Elements {
		elements = append(elements, m.String())
	}

	return "{" + strings.Join(elements, ", ") + "}"
}
