package objects

import (
	"github.com/d5/tengo/compiler/token"
)

// Closure represents a function closure.
type Closure struct {
	Fn   *CompiledFunction
	Free []*ObjectPtr
}

// TypeName returns the name of the type.
func (o *Closure) TypeName() string {
	return "closure"
}

func (o *Closure) String() string {
	return "<closure>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Closure) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Closure) Copy() Object {
	return &Closure{
		Fn:   o.Fn.Copy().(*CompiledFunction),
		Free: append([]*ObjectPtr{}, o.Free...), // DO NOT Copy() of elements; these are variable pointers
	}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Closure) IsFalsy() bool {
	return false
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Closure) Equals(x Object) bool {
	return false
}
