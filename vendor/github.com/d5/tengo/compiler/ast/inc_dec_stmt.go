package ast

import (
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
)

// IncDecStmt represents increment or decrement statement.
type IncDecStmt struct {
	Expr     Expr
	Token    token.Token
	TokenPos source.Pos
}

func (s *IncDecStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *IncDecStmt) Pos() source.Pos {
	return s.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (s *IncDecStmt) End() source.Pos {
	return source.Pos(int(s.TokenPos) + 2)
}

func (s *IncDecStmt) String() string {
	return s.Expr.String() + s.Token.String()
}
