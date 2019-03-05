package script

import (
	"context"
	"fmt"

	"github.com/d5/tengo/compiler"
	"github.com/d5/tengo/compiler/parser"
	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/runtime"
)

// Script can simplify compilation and execution of embedded scripts.
type Script struct {
	variables        map[string]*Variable
	builtinFuncs     []objects.Object
	builtinModules   map[string]*objects.Object
	userModuleLoader compiler.ModuleLoader
	input            []byte
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

// SetBuiltinFunctions allows to define builtin functions.
func (s *Script) SetBuiltinFunctions(funcs []*objects.BuiltinFunction) {
	if funcs != nil {
		s.builtinFuncs = make([]objects.Object, len(funcs))
		for idx, fn := range funcs {
			s.builtinFuncs[idx] = fn
		}
	} else {
		s.builtinFuncs = []objects.Object{}
	}
}

// SetBuiltinModules allows to define builtin modules.
func (s *Script) SetBuiltinModules(modules map[string]*objects.ImmutableMap) {
	if modules != nil {
		s.builtinModules = make(map[string]*objects.Object, len(modules))
		for k, mod := range modules {
			s.builtinModules[k] = objectPtr(mod)
		}
	} else {
		s.builtinModules = map[string]*objects.Object{}
	}
}

// SetUserModuleLoader sets the user module loader for the compiler.
func (s *Script) SetUserModuleLoader(loader compiler.ModuleLoader) {
	s.userModuleLoader = loader
}

// Compile compiles the script with all the defined variables, and, returns Compiled object.
func (s *Script) Compile() (*Compiled, error) {
	symbolTable, builtinModules, globals, err := s.prepCompile()
	if err != nil {
		return nil, err
	}

	fileSet := source.NewFileSet()
	srcFile := fileSet.AddFile("(main)", -1, len(s.input))

	p := parser.NewParser(srcFile, s.input, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	c := compiler.NewCompiler(srcFile, symbolTable, nil, builtinModules, nil)

	if s.userModuleLoader != nil {
		c.SetModuleLoader(s.userModuleLoader)
	}

	if err := c.Compile(file); err != nil {
		return nil, err
	}

	return &Compiled{
		symbolTable: symbolTable,
		machine:     runtime.NewVM(c.Bytecode(), globals, s.builtinFuncs, s.builtinModules),
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

func (s *Script) prepCompile() (symbolTable *compiler.SymbolTable, builtinModules map[string]bool, globals []*objects.Object, err error) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = compiler.NewSymbolTable()

	if s.builtinFuncs == nil {
		s.builtinFuncs = make([]objects.Object, len(objects.Builtins))
		for idx, fn := range objects.Builtins {
			s.builtinFuncs[idx] = &objects.BuiltinFunction{
				Name:  fn.Name,
				Value: fn.Value,
			}
		}
	}

	if s.builtinModules == nil {
		s.builtinModules = make(map[string]*objects.Object)
	}

	for idx, fn := range s.builtinFuncs {
		f := fn.(*objects.BuiltinFunction)
		symbolTable.DefineBuiltin(idx, f.Name)
	}

	builtinModules = make(map[string]bool)
	for name := range s.builtinModules {
		builtinModules[name] = true
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

func objectPtr(o objects.Object) *objects.Object {
	return &o
}
