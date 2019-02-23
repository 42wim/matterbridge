package script

import (
	"context"
	"fmt"

	"github.com/d5/tengo/compiler"
	"github.com/d5/tengo/compiler/parser"
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/runtime"
	"github.com/d5/tengo/stdlib"
)

// Script can simplify compilation and execution of embedded scripts.
type Script struct {
	variables         map[string]*Variable
	removedBuiltins   map[string]bool
	removedStdModules map[string]bool
	userModuleLoader  compiler.ModuleLoader
	input             []byte
}

// New creates a Script instance with an input script.
func New(input []byte) *Script {
	return &Script{
		variables: make(map[string]*Variable),
		input:     input,
	}
}

// Add adds a new variable or updates an existing variable to the script.
func (s *Script) Add(name string, value interface{}) error {
	obj, err := objects.FromInterface(value)
	if err != nil {
		return err
	}

	s.variables[name] = &Variable{
		name:  name,
		value: &obj,
	}

	return nil
}

// Remove removes (undefines) an existing variable for the script.
// It returns false if the variable name is not defined.
func (s *Script) Remove(name string) bool {
	if _, ok := s.variables[name]; !ok {
		return false
	}

	delete(s.variables, name)

	return true
}

// DisableBuiltinFunction disables a builtin function.
func (s *Script) DisableBuiltinFunction(name string) {
	if s.removedBuiltins == nil {
		s.removedBuiltins = make(map[string]bool)
	}

	s.removedBuiltins[name] = true
}

// DisableStdModule disables a standard library module.
func (s *Script) DisableStdModule(name string) {
	if s.removedStdModules == nil {
		s.removedStdModules = make(map[string]bool)
	}

	s.removedStdModules[name] = true
}

// SetUserModuleLoader sets the user module loader for the compiler.
func (s *Script) SetUserModuleLoader(loader compiler.ModuleLoader) {
	s.userModuleLoader = loader
}

// Compile compiles the script with all the defined variables, and, returns Compiled object.
func (s *Script) Compile() (*Compiled, error) {
	symbolTable, stdModules, globals, err := s.prepCompile()
	if err != nil {
		return nil, err
	}

	fileSet := source.NewFileSet()
	srcFile := fileSet.AddFile("(main)", -1, len(s.input))

	p := parser.NewParser(srcFile, s.input, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, fmt.Errorf("parse error: %s", err.Error())
	}

	c := compiler.NewCompiler(srcFile, symbolTable, nil, stdModules, nil)

	if s.userModuleLoader != nil {
		c.SetModuleLoader(s.userModuleLoader)
	}

	if err := c.Compile(file); err != nil {
		return nil, err
	}

	return &Compiled{
		symbolTable: symbolTable,
		machine:     runtime.NewVM(c.Bytecode(), globals, nil),
	}, nil
}

// Run compiles and runs the scripts.
// Use returned compiled object to access global variables.
func (s *Script) Run() (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}

	err = compiled.Run()

	return
}

// RunContext is like Run but includes a context.
func (s *Script) RunContext(ctx context.Context) (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}

	err = compiled.RunContext(ctx)

	return
}

func (s *Script) prepCompile() (symbolTable *compiler.SymbolTable, stdModules map[string]bool, globals []*objects.Object, err error) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = compiler.NewSymbolTable()
	for idx, fn := range objects.Builtins {
		if !s.removedBuiltins[fn.Name] {
			symbolTable.DefineBuiltin(idx, fn.Name)
		}
	}

	stdModules = make(map[string]bool)
	for name := range stdlib.Modules {
		if !s.removedStdModules[name] {
			stdModules[name] = true
		}
	}

	globals = make([]*objects.Object, runtime.GlobalsSize, runtime.GlobalsSize)

	for idx, name := range names {
		symbol := symbolTable.Define(name)
		if symbol.Index != idx {
			panic(fmt.Errorf("wrong symbol index: %d != %d", idx, symbol.Index))
		}

		globals[symbol.Index] = s.variables[name].value
	}

	return
}

func (s *Script) copyVariables() map[string]*Variable {
	vars := make(map[string]*Variable)
	for n, v := range s.variables {
		vars[n] = v
	}

	return vars
}
