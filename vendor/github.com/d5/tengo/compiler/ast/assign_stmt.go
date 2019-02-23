package ast

import (
	"strings"

	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/compiler/token"
)

// AssignStmt represents an assignment statement.
type AssignStmt struct {
	LHS      []Expr
	RHS      []Expr
	Token    token.Token
	TokenPos source.Pos
}

func (s *AssignStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *AssignStmt) Pos() source.Pos {
	return s.LHS[0].Pos()
}

// End returns the position of first character immediately after the node.
func (s *AssignStmt) End() source.Pos {
	return s.RHS[len(s.RHS)-1].End()
}

func (s *AssignStmt) String() string {
	var lhs, rhs []string
	for _, e := range s.LHS {
		lhs = append(lhs, e.String())
	}
	for _, e := range s.RHS {
		rhs = append(rhs, e.String())
	}

	return strings.Join(lhs, ", ") + " " + s.Token.String() + " " + strings.Join(rhs, ", ")
}
