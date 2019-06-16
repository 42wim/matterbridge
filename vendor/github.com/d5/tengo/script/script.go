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
	modules          *objects.ModuleMap
	input            []byte
	maxAllocs        int64
	maxConstObjects  int
	enableFileImport bool
}

// New creates a Script instance with an input script.
func New(input []byte) *Script {
	return &Script{
		variables:       make(map[string]*Variable),
		input:           input,
		maxAllocs:       -1,
		maxConstObjects: -1,
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
		value: obj,
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

// SetImports sets import modules.
func (s *Script) SetImports(modules *objects.ModuleMap) {
	s.modules = modules
}

// SetMaxAllocs sets the maximum number of objects allocations during the run time.
// Compiled script will return runtime.ErrObjectAllocLimit error if it exceeds this limit.
func (s *Script) SetMaxAllocs(n int64) {
	s.maxAllocs = n
}

// SetMaxConstObjects sets the maximum number of objects in the compiled constants.
func (s *Script) SetMaxConstObjects(n int) {
	s.maxConstObjects = n
}

// EnableFileImport enables or disables module loading from local files.
// Local file modules are disabled by default.
func (s *Script) EnableFileImport(enable bool) {
	s.enableFileImport = enable
}

// Compile compiles the script with all the defined variables, and, returns Compiled object.
func (s *Script) Compile() (*Compiled, error) {
	symbolTable, globals, err := s.prepCompile()
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

	c := compiler.NewCompiler(srcFile, symbolTable, nil, s.modules, nil)
	c.EnableFileImport(s.enableFileImport)
	if err := c.Compile(file); err != nil {
		return nil, err
	}

	// reduce globals size
	globals = globals[:symbolTable.MaxSymbols()+1]

	// global symbol names to indexes
	globalIndexes := make(map[string]int, len(globals))
	for _, name := range symbolTable.Names() {
		symbol, _, _ := symbolTable.Resolve(name)
		if symbol.Scope == compiler.ScopeGlobal {
			globalIndexes[name] = symbol.Index
		}
	}

	// remove duplicates from constants
	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()

	// check the constant objects limit
	if s.maxConstObjects >= 0 {
		cnt := bytecode.CountObjects()
		if cnt > s.maxConstObjects {
			return nil, fmt.Errorf("exceeding constant objects limit: %d", cnt)
		}
	}

	return &Compiled{
		globalIndexes: globalIndexes,
		bytecode:      bytecode,
		globals:       globals,
		maxAllocs:     s.maxAllocs,
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

func (s *Script) prepCompile() (symbolTable *compiler.SymbolTable, globals []objects.Object, err error) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = compiler.NewSymbolTable()
	for idx, fn := range objects.Builtins {
		symbolTable.DefineBuiltin(idx, fn.Name)
	}

	globals = make([]objects.Object, runtime.GlobalsSize)

	for idx, name := range names {
		symbol := symbolTable.Define(name)
		if symbol.Index != idx {
			panic(fmt.Errorf("wrong symbol index: %d != %d", idx, symbol.Index))
		}

		globals[symbol.Index] = s.variables[name].value
	}

	return
}
