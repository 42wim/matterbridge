package tengo

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/d5/tengo/v2/parser"
)

// Script can simplify compilation and execution of embedded scripts.
type Script struct {
	variables        map[string]*Variable
	modules          *ModuleMap
	input            []byte
	maxAllocs        int64
	maxConstObjects  int
	enableFileImport bool
	importDir        string
}

// NewScript creates a Script instance with an input script.
func NewScript(input []byte) *Script {
	return &Script{
		variables:       make(map[string]*Variable),
		input:           input,
		maxAllocs:       -1,
		maxConstObjects: -1,
	}
}

// Add adds a new variable or updates an existing variable to the script.
func (s *Script) Add(name string, value interface{}) error {
	obj, err := FromInterface(value)
	if err != nil {
		return err
	}
	s.variables[name] = &Variable{
		name:  name,
		value: obj,
	}
	return nil
}

// Remove removes (undefines) an existing variable for the script. It returns
// false if the variable name is not defined.
func (s *Script) Remove(name string) bool {
	if _, ok := s.variables[name]; !ok {
		return false
	}
	delete(s.variables, name)
	return true
}

// SetImports sets import modules.
func (s *Script) SetImports(modules *ModuleMap) {
	s.modules = modules
}

// SetImportDir sets the initial import directory for script files.
func (s *Script) SetImportDir(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	s.importDir = dir
	return nil
}

// SetMaxAllocs sets the maximum number of objects allocations during the run
// time. Compiled script will return ErrObjectAllocLimit error if it
// exceeds this limit.
func (s *Script) SetMaxAllocs(n int64) {
	s.maxAllocs = n
}

// SetMaxConstObjects sets the maximum number of objects in the compiled
// constants.
func (s *Script) SetMaxConstObjects(n int) {
	s.maxConstObjects = n
}

// EnableFileImport enables or disables module loading from local files. Local
// file modules are disabled by default.
func (s *Script) EnableFileImport(enable bool) {
	s.enableFileImport = enable
}

// Compile compiles the script with all the defined variables, and, returns
// Compiled object.
func (s *Script) Compile() (*Compiled, error) {
	symbolTable, globals, err := s.prepCompile()
	if err != nil {
		return nil, err
	}

	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("(main)", -1, len(s.input))
	p := parser.NewParser(srcFile, s.input, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	c := NewCompiler(srcFile, symbolTable, nil, s.modules, nil)
	c.EnableFileImport(s.enableFileImport)
	c.SetImportDir(s.importDir)
	if err := c.Compile(file); err != nil {
		return nil, err
	}

	// reduce globals size
	globals = globals[:symbolTable.MaxSymbols()+1]

	// global symbol names to indexes
	globalIndexes := make(map[string]int, len(globals))
	for _, name := range symbolTable.Names() {
		symbol, _, _ := symbolTable.Resolve(name, false)
		if symbol.Scope == ScopeGlobal {
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

// Run compiles and runs the scripts. Use returned compiled object to access
// global variables.
func (s *Script) Run() (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}
	err = compiled.Run()
	return
}

// RunContext is like Run but includes a context.
func (s *Script) RunContext(
	ctx context.Context,
) (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}
	err = compiled.RunContext(ctx)
	return
}

func (s *Script) prepCompile() (
	symbolTable *SymbolTable,
	globals []Object,
	err error,
) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = NewSymbolTable()
	for idx, fn := range builtinFuncs {
		symbolTable.DefineBuiltin(idx, fn.Name)
	}

	globals = make([]Object, GlobalsSize)

	for idx, name := range names {
		symbol := symbolTable.Define(name)
		if symbol.Index != idx {
			panic(fmt.Errorf("wrong symbol index: %d != %d",
				idx, symbol.Index))
		}
		globals[symbol.Index] = s.variables[name].value
	}
	return
}

// Compiled is a compiled instance of the user script. Use Script.Compile() to
// create Compiled object.
type Compiled struct {
	globalIndexes map[string]int // global symbol name to index
	bytecode      *Bytecode
	globals       []Object
	maxAllocs     int64
	lock          sync.RWMutex
}

// Run executes the compiled script in the virtual machine.
func (c *Compiled) Run() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	v := NewVM(c.bytecode, c.globals, c.maxAllocs)
	return v.Run()
}

// RunContext is like Run but includes a context.
func (c *Compiled) RunContext(ctx context.Context) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	v := NewVM(c.bytecode, c.globals, c.maxAllocs)
	ch := make(chan error, 1)
	go func() {
		ch <- v.Run()
	}()

	select {
	case <-ctx.Done():
		v.Abort()
		<-ch
		err = ctx.Err()
	case err = <-ch:
	}
	return
}

// Clone creates a new copy of Compiled. Cloned copies are safe for concurrent
// use by multiple goroutines.
func (c *Compiled) Clone() *Compiled {
	c.lock.Lock()
	defer c.lock.Unlock()

	clone := &Compiled{
		globalIndexes: c.globalIndexes,
		bytecode:      c.bytecode,
		globals:       make([]Object, len(c.globals)),
		maxAllocs:     c.maxAllocs,
	}
	// copy global objects
	for idx, g := range c.globals {
		if g != nil {
			clone.globals[idx] = g
		}
	}
	return clone
}

// IsDefined returns true if the variable name is defined (has value) before or
// after the execution.
func (c *Compiled) IsDefined(name string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	idx, ok := c.globalIndexes[name]
	if !ok {
		return false
	}
	v := c.globals[idx]
	if v == nil {
		return false
	}
	return v != UndefinedValue
}

// Get returns a variable identified by the name.
func (c *Compiled) Get(name string) *Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value := UndefinedValue
	if idx, ok := c.globalIndexes[name]; ok {
		value = c.globals[idx]
		if value == nil {
			value = UndefinedValue
		}
	}
	return &Variable{
		name:  name,
		value: value,
	}
}

// GetAll returns all the variables that are defined by the compiled script.
func (c *Compiled) GetAll() []*Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var vars []*Variable
	for name, idx := range c.globalIndexes {
		value := c.globals[idx]
		if value == nil {
			value = UndefinedValue
		}
		vars = append(vars, &Variable{
			name:  name,
			value: value,
		})
	}
	return vars
}

// Set replaces the value of a global variable identified by the name. An error
// will be returned if the name was not defined during compilation.
func (c *Compiled) Set(name string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	obj, err := FromInterface(value)
	if err != nil {
		return err
	}
	idx, ok := c.globalIndexes[name]
	if !ok {
		return fmt.Errorf("'%s' is not defined", name)
	}
	c.globals[idx] = obj
	return nil
}
