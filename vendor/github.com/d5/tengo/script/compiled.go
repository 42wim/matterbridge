package script

import (
	"context"
	"fmt"
	"sync"

	"github.com/d5/tengo/compiler"
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/runtime"
)

// Compiled is a compiled instance of the user script.
// Use Script.Compile() to create Compiled object.
type Compiled struct {
	globalIndexes map[string]int // global symbol name to index
	bytecode      *compiler.Bytecode
	globals       []objects.Object
	maxAllocs     int64
	lock          sync.RWMutex
}

// Run executes the compiled script in the virtual machine.
func (c *Compiled) Run() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	v := runtime.NewVM(c.bytecode, c.globals, c.maxAllocs)

	return v.Run()
}

// RunContext is like Run but includes a context.
func (c *Compiled) RunContext(ctx context.Context) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	v := runtime.NewVM(c.bytecode, c.globals, c.maxAllocs)

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

// Clone creates a new copy of Compiled.
// Cloned copies are safe for concurrent use by multiple goroutines.
func (c *Compiled) Clone() *Compiled {
	c.lock.Lock()
	defer c.lock.Unlock()

	clone := &Compiled{
		globalIndexes: c.globalIndexes,
		bytecode:      c.bytecode,
		globals:       make([]objects.Object, len(c.globals)),
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

// IsDefined returns true if the variable name is defined (has value) before or after the execution.
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

	return v != objects.UndefinedValue
}

// Get returns a variable identified by the name.
func (c *Compiled) Get(name string) *Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value := objects.UndefinedValue

	if idx, ok := c.globalIndexes[name]; ok {
		value = c.globals[idx]
		if value == nil {
			value = objects.UndefinedValue
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
			value = objects.UndefinedValue
		}

		vars = append(vars, &Variable{
			name:  name,
			value: value,
		})
	}

	return vars
}

// Set replaces the value of a global variable identified by the name.
// An error will be returned if the name was not defined during compilation.
func (c *Compiled) Set(name string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	obj, err := objects.FromInterface(value)
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
