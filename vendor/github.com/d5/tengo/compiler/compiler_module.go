package compiler

import (
	"io/ioutil"
	"strings"

	"github.com/d5/tengo/compiler/ast"
	"github.com/d5/tengo/compiler/parser"
	"github.com/d5/tengo/objects"
)

func (c *Compiler) compileModule(expr *ast.ImportExpr) (*objects.CompiledFunction, error) {
	compiledModule, exists := c.loadCompiledModule(expr.ModuleName)
	if exists {
		return compiledModule, nil
	}

	moduleName := expr.ModuleName

	// read module source from loader
	var moduleSrc []byte
	if c.moduleLoader == nil {
		// default loader: read from local file
		if !strings.HasSuffix(moduleName, ".tengo") {
			moduleName += ".tengo"
		}

		if err := c.checkCyclicImports(expr, moduleName); err != nil {
			return nil, err
		}

		var err error
		moduleSrc, err = ioutil.ReadFile(moduleName)
		if err != nil {
			return nil, c.errorf(expr, "module file read error: %s", err.Error())
		}
	} else {
		if err := c.checkCyclicImports(expr, moduleName); err != nil {
			return nil, err
		}

		var err error
		moduleSrc, err = c.moduleLoader(moduleName)
		if err != nil {
			return nil, err
		}
	}

	compiledModule, err := c.doCompileModule(moduleName, moduleSrc)
	if err != nil {
		return nil, err
	}

	c.storeCompiledModule(moduleName, compiledModule)

	return compiledModule, nil
}

func (c *Compiler) checkCyclicImports(node ast.Node, moduleName string) error {
	if c.moduleName == moduleName {
		return c.errorf(node, "cyclic module import: %s", moduleName)
	} else if c.parent != nil {
		return c.parent.checkCyclicImports(node, moduleName)
	}

	return nil
}

func (c *Compiler) doCompileModule(moduleName string, src []byte) (*objects.CompiledFunction, error) {
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
	moduleCompiler := c.fork(modFile, moduleName, symbolTable)
	if err := moduleCompiler.Compile(file); err != nil {
		return nil, err
	}

	// add OpReturn (== export undefined) if export is missing
	if !moduleCompiler.lastInstructionIs(OpReturnValue) {
		moduleCompiler.emit(nil, OpReturn)
	}

	compiledFunc := moduleCompiler.Bytecode().MainFunction
	compiledFunc.NumLocals = symbolTable.MaxSymbols()

	return compiledFunc, nil
}

func (c *Compiler) loadCompiledModule(moduleName string) (mod *objects.CompiledFunction, ok bool) {
	if c.parent != nil {
		return c.parent.loadCompiledModule(moduleName)
	}

	mod, ok = c.compiledModules[moduleName]

	return
}

func (c *Compiler) storeCompiledModule(moduleName string, module *objects.CompiledFunction) {
	if c.parent != nil {
		c.parent.storeCompiledModule(moduleName, module)
	}

	c.compiledModules[moduleName] = module
}
