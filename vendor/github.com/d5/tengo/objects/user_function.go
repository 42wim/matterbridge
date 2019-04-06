package objects

import (
	"github.com/d5/tengo/compiler/token"
)

// UserFunction represents a user function.
type UserFunction struct {
	Name       string
	Value      CallableFunc
	EncodingID string
}

// TypeName returns the name of the type.
func (o *UserFunction) TypeName() string {
	return "user-function:" + o.Name
}

func (o *UserFunction) String() string {
	return "<user-function>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *UserFunction) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *UserFunction) Copy() Object {
	return &UserFunction{Value: o.Value}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *UserFunction) IsFalsy() bool {
	return false
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *UserFunction) Equals(x Object) bool {
	return false
}

// Call invokes a user function.
func (o *UserFunction) Call(args ...Object) (Object, error) {
	return o.Value(args...)
}
