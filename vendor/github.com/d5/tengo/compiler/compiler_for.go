package compiler

import (
	"github.com/d5/tengo/compiler/ast"
)

func (c *Compiler) compileForStmt(stmt *ast.ForStmt) error {
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
	}()

	// init statement
	if stmt.Init != nil {
		if err := c.Compile(stmt.Init); err != nil {
			return err
		}
	}

	// pre-condition position
	preCondPos := len(c.currentInstructions())

	// condition expression
	postCondPos := -1
	if stmt.Cond != nil {
		if err := c.Compile(stmt.Cond); err != nil {
			return err
		}
		// condition jump position
		postCondPos = c.emit(stmt, OpJumpFalsy, 0)
	}

	// enter loop
	loop := c.enterLoop()

	// body statement
	if err := c.Compile(stmt.Body); err != nil {
		c.leaveLoop()
		return err
	}

	c.leaveLoop()

	// post-body position
	postBodyPos := len(c.currentInstructions())

	// post statement
	if stmt.Post != nil {
		if err := c.Compile(stmt.Post); err != nil {
			return err
		}
	}

	// back to condition
	c.emit(stmt, OpJump, preCondPos)

	// post-statement position
	postStmtPos := len(c.currentInstructions())
	if postCondPos >= 0 {
		c.changeOperand(postCondPos, postStmtPos)
	}

	// update all break/continue jump positions
	for _, pos := range loop.Breaks {
		c.changeOperand(pos, postStmtPos)
	}
	for _, pos := range loop.Continues {
		c.changeOperand(pos, postBodyPos)
	}

	return nil
}

func (c *Compiler) compileForInStmt(stmt *ast.ForInStmt) error {
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
	}()

	// for-in statement is compiled like following:
	//
	//   for :it := iterator(iterable); :it.next();  {
	//     k, v := :it.get()  // DEFINE operator
	//
	//     ... body ...
	//   }
	//
	// ":it" is a local variable but will be conflict with other user variables
	// because character ":" is not allowed.

	// init
	//   :it = iterator(iterable)
	itSymbol := c.symbolTable.Define(":it")
	if err := c.Compile(stmt.Iterable); err != nil {
		return err
	}
	c.emit(stmt, OpIteratorInit)
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, OpSetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, OpDefineLocal, itSymbol.Index)
	}

	// pre-condition position
	preCondPos := len(c.currentInstructions())

	// condition
	//  :it.HasMore()
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, OpGetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, OpGetLocal, itSymbol.Index)
	}
	c.emit(stmt, OpIteratorNext)

	// condition jump position
	postCondPos := c.emit(stmt, OpJumpFalsy, 0)

	// enter loop
	loop := c.enterLoop()

	// assign key variable
	if stmt.Key.Name != "_" {
		keySymbol := c.symbolTable.Define(stmt.Key.Name)
		if itSymbol.Scope == ScopeGlobal {
			c.emit(stmt, OpGetGlobal, itSymbol.Index)
		} else {
			c.emit(stmt, OpGetLocal, itSymbol.Index)
		}
		c.emit(stmt, OpIteratorKey)
		if keySymbol.Scope == ScopeGlobal {
			c.emit(stmt, OpSetGlobal, keySymbol.Index)
		} else {
			c.emit(stmt, OpDefineLocal, keySymbol.Index)
		}
	}

	// assign value variable
	if stmt.Value.Name != "_" {
		valueSymbol := c.symbolTable.Define(stmt.Value.Name)
		if itSymbol.Scope == ScopeGlobal {
			c.emit(stmt, OpGetGlobal, itSymbol.Index)
		} else {
			c.emit(stmt, OpGetLocal, itSymbol.Index)
		}
		c.emit(stmt, OpIteratorValue)
		if valueSymbol.Scope == ScopeGlobal {
			c.emit(stmt, OpSetGlobal, valueSymbol.Index)
		} else {
			c.emit(stmt, OpDefineLocal, valueSymbol.Index)
		}
	}

	// body statement
	if err := c.Compile(stmt.Body); err != nil {
		c.leaveLoop()
		return err
	}

	c.leaveLoop()

	// post-body position
	postBodyPos := len(c.currentInstructions())

	// back to condition
	c.emit(stmt, OpJump, preCondPos)

	// post-statement position
	postStmtPos := len(c.currentInstructions())
	c.changeOperand(postCondPos, postStmtPos)

	// update all break/continue jump positions
	for _, pos := range loop.Breaks {
		c.changeOperand(pos, postStmtPos)
	}
	for _, pos := range loop.Continues {
		c.changeOperand(pos, postBodyPos)
	}

	return nil
}
