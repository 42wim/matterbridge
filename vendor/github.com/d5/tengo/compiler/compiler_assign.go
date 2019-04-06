package compiler

import (
	"fmt"

	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/token"
)

func (c *Compiler) compileAssign(node ast.Node, lhs, rhs []ast.Expr, op token.Token) error {
	numLHS, numRHS := len(lhs), len(rhs)
	if numLHS > 1 || numRHS > 1 {
		return c.errorf(node, "tuple assignment not allowed")
	}

	// resolve and compile left-hand side
	ident, selectors := resolveAssignLHS(lhs[0])
	numSel := len(selectors)

	if op == token.Define && numSel > 0 {
		// using selector on new variable does not make sense
		return c.errorf(node, "operator ':=' not allowed with selector")
	}

	symbol, depth, exists := c.symbolTable.Resolve(ident)
	if op == token.Define {
		if depth == 0 && exists {
			return c.errorf(node, "'%s' redeclared in this block", ident)
		}

		symbol = c.symbolTable.Define(ident)
	} else {
		if !exists {
			return c.errorf(node, "unresolved reference '%s'", ident)
		}
	}

	// +=, -=, *=, /=
	if op != token.Assign && op != token.Define {
		if err := c.Compile(lhs[0]); err != nil {
			return err
		}
	}

	// compile RHSs
	for _, expr := range rhs {
		if err := c.Compile(expr); err != nil {
			return err
		}
	}

	switch op {
	case token.AddAssign:
		c.emit(node, OpBinaryOp, int(token.Add))
	case token.SubAssign:
		c.emit(node, OpBinaryOp, int(token.Sub))
	case token.MulAssign:
		c.emit(node, OpBinaryOp, int(token.Mul))
	case token.QuoAssign:
		c.emit(node, OpBinaryOp, int(token.Quo))
	case token.RemAssign:
		c.emit(node, OpBinaryOp, int(token.Rem))
	case token.AndAssign:
		c.emit(node, OpBinaryOp, int(token.And))
	case token.OrAssign:
		c.emit(node, OpBinaryOp, int(token.Or))
	case token.AndNotAssign:
		c.emit(node, OpBinaryOp, int(token.AndNot))
	case token.XorAssign:
		c.emit(node, OpBinaryOp, int(token.Xor))
	case token.ShlAssign:
		c.emit(node, OpBinaryOp, int(token.Shl))
	case token.ShrAssign:
		c.emit(node, OpBinaryOp, int(token.Shr))
	}

	// compile selector expressions (right to left)
	for i := numSel - 1; i >= 0; i-- {
		if err := c.Compile(selectors[i]); err != nil {
			return err
		}
	}

	switch symbol.Scope {
	case ScopeGlobal:
		if numSel > 0 {
			c.emit(node, OpSetSelGlobal, symbol.Index, numSel)
		} else {
			c.emit(node, OpSetGlobal, symbol.Index)
		}
	case ScopeLocal:
		if numSel > 0 {
			c.emit(node, OpSetSelLocal, symbol.Index, numSel)
		} else {
			if op == token.Define && !symbol.LocalAssigned {
				c.emit(node, OpDefineLocal, symbol.Index)
			} else {
				c.emit(node, OpSetLocal, symbol.Index)
			}
		}

		// mark the symbol as local-assigned
		symbol.LocalAssigned = true
	case ScopeFree:
		if numSel > 0 {
			c.emit(node, OpSetSelFree, symbol.Index, numSel)
		} else {
			c.emit(node, OpSetFree, symbol.Index)
		}
	default:
		panic(fmt.Errorf("invalid assignment variable scope: %s", symbol.Scope))
	}

	return nil
}

func resolveAssignLHS(expr ast.Expr) (name string, selectors []ast.Expr) {
	switch term := expr.(type) {
	case *ast.SelectorExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Sel)
		return

	case *ast.IndexExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Index)

	case *ast.Ident:
		name = term.Name
	}

	return
}
