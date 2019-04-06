package objects

import (
	"github.com/d5/tengo/compiler/token"
)

// ObjectPtr represents a free variable.
type ObjectPtr struct {
	Value *Object
}

func (o *ObjectPtr) String() string {
	return "free-var"
}

// TypeName returns the name of the type.
func (o *ObjectPtr) TypeName() string {
	return "<free-var>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *ObjectPtr) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *ObjectPtr) Copy() Object {
	return o
}

// IsFalsy returns true if the value of the type is falsy.
func (o *ObjectPtr) IsFalsy() bool {
	return o.Value == nil
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *ObjectPtr) Equals(x Object) bool {
	return o == x
}
