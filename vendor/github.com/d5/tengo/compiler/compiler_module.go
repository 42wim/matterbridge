package compiler

import (
	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/parser"
	"github.com/d5/tengo/objects"
)

func (c *Compiler) checkCyclicImports(node ast.Node, modulePath string) error {
	if c.modulePath == modulePath {
		return c.errorf(node, "cyclic module import: %s", modulePath)
	} else if c.parent != nil {
		return c.parent.checkCyclicImports(node, modulePath)
	}

	return nil
}

func (c *Compiler) compileModule(node ast.Node, moduleName, modulePath string, src []byte) (*objects.CompiledFunction, error) {
	if err := c.checkCyclicImports(node, modulePath); err != nil {
		return nil, err
	}

	compiledModule, exists := c.loadCompiledModule(modulePath)
	if exists {
		return compiledModule, nil
	}

	modFile := c.file.Set().AddFile(moduleName, -1, len(src))
	p := parser.NewParser(modFile, src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	symbolTable := NewSymbolTable()

	// inherit builtin functions
	for _, sym := range c.symbolTable.BuiltinSymbols() {
		symbolTable.DefineBuiltin(sym.Index, sym.Name)
	}

	// no global scope for the module
	symbolTable = symbolTable.Fork(false)

	// compile module
	moduleCompiler := c.fork(modFile, modulePath, symbolTable)
	if err := moduleCompiler.Compile(file); err != nil {
		return nil, err
	}

	// code optimization
	moduleCompiler.optimizeFunc(node)

	compiledFunc := moduleCompiler.Bytecode().MainFunction
	compiledFunc.NumLocals = symbolTable.MaxSymbols()

	c.storeCompiledModule(modulePath, compiledFunc)

	return compiledFunc, nil
}

func (c *Compiler) loadCompiledModule(modulePath string) (mod *objects.CompiledFunction, ok bool) {
	if c.parent != nil {
		return c.parent.loadCompiledModule(modulePath)
	}

	mod, ok = c.compiledModules[modulePath]

	return
}

func (c *Compiler) storeCompiledModule(modulePath string, module *objects.CompiledFunction) {
	if c.parent != nil {
		c.parent.storeCompiledModule(modulePath, module)
	}

	c.compiledModules[modulePath] = module
}
