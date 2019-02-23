package ast

import (
	"github.com/d5/tengo/compiler/source"
)

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	ReturnPos source.Pos
	Result    Expr
}

func (s *ReturnStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ReturnStmt) Pos() source.Pos {
	return s.ReturnPos
}

// End returns the position of first character immediately after the node.
func (s *ReturnStmt) End() source.Pos {
	if s.Result != nil {
		return s.Result.End()
	}

	return s.ReturnPos + 6
}

func (s *ReturnStmt) String() string {
	if s.Result != nil {
		return "return " + s.Result.String()
	}

	return "return"
}
