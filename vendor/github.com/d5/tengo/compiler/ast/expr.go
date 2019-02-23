package ast

// Expr represents an expression node in the AST.
type Expr interface {
	Node
	exprNode()
}
