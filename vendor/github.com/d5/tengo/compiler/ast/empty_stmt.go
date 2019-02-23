package ast

import "github.com/d5/tengo/compiler/source"

// EmptyStmt represents an empty statement.
type EmptyStmt struct {
	Semicolon source.Pos
	Implicit  bool
}

func (s *EmptyStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *EmptyStmt) Pos() source.Pos {
	return s.Semicolon
}

// End returns the position of first character immediately after the node.
func (s *EmptyStmt) End() source.Pos {
	if s.Implicit {
		return s.Semicolon
	}

	return s.Semicolon + 1
}

func (s *EmptyStmt) String() string {
	return ";"
}
