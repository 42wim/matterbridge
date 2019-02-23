package ast

// Stmt represents a statement in the AST.
type Stmt interface {
	Node
	stmtNode()
}
