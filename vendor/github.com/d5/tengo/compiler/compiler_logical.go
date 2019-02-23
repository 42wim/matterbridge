package compiler

import (
	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/token"
)

func (c *Compiler) compileLogical(node *ast.BinaryExpr) error {
	// left side term
	if err := c.Compile(node.LHS); err != nil {
		return err
	}

	// jump position
	var jumpPos int
	if node.Token == token.LAnd {
		jumpPos = c.emit(node, OpAndJump, 0)
	} else {
		jumpPos = c.emit(node, OpOrJump, 0)
	}

	// right side term
	if err := c.Compile(node.RHS); err != nil {
		return err
	}

	c.changeOperand(jumpPos, len(c.currentInstructions()))

	return nil
}
