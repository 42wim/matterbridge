package ast

import (
	"github.com/d5/tengo/compiler/source"
)

// ExportStmt represents an export statement.
type ExportStmt struct {
	ExportPos source.Pos
	Result    Expr
}

func (s *ExportStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ExportStmt) Pos() source.Pos {
	return s.ExportPos
}

// End returns the position of first character immediately after the node.
func (s *ExportStmt) End() source.Pos {
	return s.Result.End()
}

func (s *ExportStmt) String() string {
	return "export " + s.Result.String()
}
