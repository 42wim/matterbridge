package parser

import (
	"strings"
)

// File represents a file unit.
type File struct {
	InputFile *SourceFile
	Stmts     []Stmt
}

// Pos returns the position of first character belonging to the node.
func (n *File) Pos() Pos {
	return Pos(n.InputFile.Base)
}

// End returns the position of first character immediately after the node.
func (n *File) End() Pos {
	return Pos(n.InputFile.Base + n.InputFile.Size)
}

func (n *File) String() string {
	var stmts []string
	for _, e := range n.Stmts {
		stmts = append(stmts, e.String())
	}
	return strings.Join(stmts, "; ")
}
