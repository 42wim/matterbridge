package ast

import "github.com/d5/tengo/compiler/source"

// Node represents a node in the AST.
type Node interface {
	// Pos returns the position of first character belonging to the node.
	Pos() source.Pos
	// End returns the position of first character immediately after the node.
	End() source.Pos
	// String returns a string representation of the node.
	String() string
}
