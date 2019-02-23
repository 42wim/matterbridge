package objects

import (
	"fmt"

	"github.com/d5/tengo/compiler/token"
)

// Error represents a string value.
type Error struct {
	Value Object
}

// TypeName returns the name of the type.
func (o *Error) TypeName() string {
	return "error"
}

func (o *Error) String() string {
	if o.Value != nil {
		return fmt.Sprintf("error: %s", o.Value.String())
	}

	return "error"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Error) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Error) IsFalsy() bool {
	return true // error is always false.
}

// Copy returns a copy of the type.
func (o *Error) Copy() Object {
	return &Error{Value: o.Value.Copy()}
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Error) Equals(x Object) bool {
	return o == x // pointer equality
}
