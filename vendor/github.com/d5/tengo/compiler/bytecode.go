package compiler

import (
	"encoding/gob"
	"fmt"
	"io"
	"reflect"

	"github.com/d5/tengo/compiler/source"
	"github.com/d5/tengo/objects"
)

// Bytecode is a compiled instructions and constants.
type Bytecode struct {
	FileSet      *source.FileSet
	MainFunction *objects.CompiledFunction
	Constants    []objects.Object
}

// Decode reads Bytecode data from the reader.
func (b *Bytecode) Decode(r io.Reader) error {
	dec := gob.NewDecoder(r)

	if err := dec.Decode(&b.FileSet); err != nil {
		return err
	}
	// TODO: files in b.FileSet.File does not have their 'set' field properly set to b.FileSet
	// as it's private field and not serialized by gob encoder/decoder.

	if err := dec.Decode(&b.MainFunction); err != nil {
		return err
	}

	if err := dec.Decode(&b.Constants); err != nil {
		return err
	}

	// replace Bool and Undefined with known value
	for i, v := range b.Constants {
		b.Constants[i] = cleanupObjects(v)
	}

	return nil
}

// Encode writes Bytecode data to the writer.
func (b *Bytecode) Encode(w io.Writer) error {
	enc := gob.NewEncoder(w)

	if err := enc.Encode(b.FileSet); err != nil {
		return err
	}

	if err := enc.Encode(b.MainFunction); err != nil {
		return err
	}

	// constants
	return enc.Encode(b.Constants)
}

// FormatInstructions returns human readable string representations of
// compiled instructions.
func (b *Bytecode) FormatInstructions() []string {
	return FormatInstructions(b.MainFunction.Instructions, 0)
}

// FormatConstants returns human readable string representations of
// compiled constants.
func (b *Bytecode) FormatConstants() (output []string) {
	for cidx, cn := range b.Constants {
		switch cn := cn.(type) {
		case *objects.CompiledFunction:
			output = append(output, fmt.Sprintf("[% 3d] (Compiled Function|%p)", cidx, &cn))
			for _, l := range FormatInstructions(cn.Instructions, 0) {
				output = append(output, fmt.Sprintf("     %s", l))
			}
		default:
			output = append(output, fmt.Sprintf("[% 3d] %s (%s|%p)", cidx, cn, reflect.TypeOf(cn).Elem().Name(), &cn))
		}
	}

	return
}

func cleanupObjects(o objects.Object) objects.Object {
	switch o := o.(type) {
	case *objects.Bool:
		if o.IsFalsy() {
			return objects.FalseValue
		}
		return objects.TrueValue
	case *objects.Undefined:
		return objects.UndefinedValue
	case *objects.Array:
		for i, v := range o.Value {
			o.Value[i] = cleanupObjects(v)
		}
	case *objects.Map:
		for k, v := range o.Value {
			o.Value[k] = cleanupObjects(v)
		}
	}

	return o
}

func init() {
	gob.Register(&source.FileSet{})
	gob.Register(&source.File{})
	gob.Register(&objects.Array{})
	gob.Register(&objects.ArrayIterator{})
	gob.Register(&objects.Bool{})
	gob.Register(&objects.Break{})
	gob.Register(&objects.BuiltinFunction{})
	gob.Register(&objects.Bytes{})
	gob.Register(&objects.Char{})
	gob.Register(&objects.Closure{})
	gob.Register(&objects.CompiledFunction{})
	gob.Register(&objects.Continue{})
	gob.Register(&objects.Error{})
	gob.Register(&objects.Float{})
	gob.Register(&objects.ImmutableArray{})
	gob.Register(&objects.ImmutableMap{})
	gob.Register(&objects.Int{})
	gob.Register(&objects.Map{})
	gob.Register(&objects.MapIterator{})
	gob.Register(&objects.ReturnValue{})
	gob.Register(&objects.String{})
	gob.Register(&objects.StringIterator{})
	gob.Register(&objects.Time{})
	gob.Register(&objects.Undefined{})
	gob.Register(&objects.UserFunction{})
}
