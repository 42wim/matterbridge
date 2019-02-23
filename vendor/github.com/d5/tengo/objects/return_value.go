package objects

import "github.com/d5/tengo/compiler/token"

// ReturnValue represents a value that is being returned.
type ReturnValue struct {
	Value Object
}

// TypeName returns the name of the type.
func (o *ReturnValue) TypeName() string {
	return "return-value"
}

func (o *ReturnValue) String() string {
	return "<return-value>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *ReturnValue) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *ReturnValue) Copy() Object {
	return &ReturnValue{Value: o.Copy()}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *ReturnValue) IsFalsy() bool {
	return false
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *ReturnValue) Equals(x Object) bool {
	return false
}
