package ast

import "github.com/d5/tengo/compiler/source"

// ForInStmt represents a for-in statement.
type ForInStmt struct {
	ForPos   source.Pos
	Key      *Ident
	Value    *Ident
	Iterable Expr
	Body     *BlockStmt
}

func (s *ForInStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ForInStmt) Pos() source.Pos {
	return s.ForPos
}

// End returns the position of first character immediately after the node.
func (s *ForInStmt) End() source.Pos {
	return s.Body.End()
}

func (s *ForInStmt) String() string {
	if s.Value != nil {
		return "for " + s.Key.String() + ", " + s.Value.String() + " in " + s.Iterable.String() + " " + s.Body.String()
	}

	return "for " + s.Key.String() + " in " + s.Iterable.String() + " " + s.Body.String()
}
