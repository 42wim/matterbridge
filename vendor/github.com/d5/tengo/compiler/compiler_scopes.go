package compiler

import "github.com/d5/tengo/compiler/source"

func (c *Compiler) currentInstructions() []byte {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) currentSourceMap() map[int]source.Pos {
	return c.scopes[c.scopeIndex].sourceMap
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		symbolInit: make(map[string]bool),
		sourceMap:  make(map[int]source.Pos),
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = c.symbolTable.Fork(false)

	if c.trace != nil {
		c.printTrace("SCOPE", c.scopeIndex)
	}
}

func (c *Compiler) leaveScope() (instructions []byte, sourceMap map[int]source.Pos) {
	instructions = c.currentInstructions()
	sourceMap = c.currentSourceMap()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Parent(true)

	if c.trace != nil {
		c.printTrace("SCOPL", c.scopeIndex)
	}

	return
}
