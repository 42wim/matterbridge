package script

import (
	"context"
	"fmt"

	"github.com/d5/tengo/compiler"
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/runtime"
)

// Compiled is a compiled instance of the user script.
// Use Script.Compile() to create Compiled object.
type Compiled struct {
	symbolTable *compiler.SymbolTable
	machine     *runtime.VM
}

// Run executes the compiled script in the virtual machine.
func (c *Compiled) Run() error {
	return c.machine.Run()
}

// RunContext is like Run but includes a context.
func (c *Compiled) RunContext(ctx context.Context) (err error) {
	ch := make(chan error, 1)

	go func() {
		ch <- c.machine.Run()
	}()

	select {
	case <-ctx.Done():
		c.machine.Abort()
		<-ch
		err = ctx.Err()
	case err = <-ch:
	}

	return
}

// IsDefined returns true if the variable name is defined (has value) before or after the execution.
func (c *Compiled) IsDefined(name string) bool {
	symbol, _, ok := c.symbolTable.Resolve(name)
	if !ok {
		return false
	}

	v := c.machine.Globals()[symbol.Index]
	if v == nil {
		return false
	}

	return *v != objects.UndefinedValue
}

// Get returns a variable identified by the name.
func (c *Compiled) Get(name string) *Variable {
	value := &objects.UndefinedValue

	symbol, _, ok := c.symbolTable.Resolve(name)
	if ok && symbol.Scope == compiler.ScopeGlobal {
		value = c.machine.Globals()[symbol.Index]
		if value == nil {
			value = &objects.UndefinedValue
		}
	}

	return &Variable{
		name:  name,
		value: value,
	}
}

// GetAll returns all the variables that are defined by the compiled script.
func (c *Compiled) GetAll() []*Variable {
	var vars []*Variable
	for _, name := range c.symbolTable.Names() {
		symbol, _, ok := c.symbolTable.Resolve(name)
		if ok && symbol.Scope == compiler.ScopeGlobal {
			value := c.machine.Globals()[symbol.Index]
			if value == nil {
				value = &objects.UndefinedValue
			}

			vars = append(vars, &Variable{
				name:  name,
				value: value,
			})
		}
	}

	return vars
}

// Set replaces the value of a global variable identified by the name.
// An error will be returned if the name was not defined during compilation.
func (c *Compiled) Set(name string, value interface{}) error {
	obj, err := objects.FromInterface(value)
	if err != nil {
		return err
	}

	symbol, _, ok := c.symbolTable.Resolve(name)
	if !ok || symbol.Scope != compiler.ScopeGlobal {
		return fmt.Errorf("'%s' is not defined", name)
	}

	c.machine.Globals()[symbol.Index] = &obj

	return nil
}
