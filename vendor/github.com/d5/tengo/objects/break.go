package objects

import "github.com/d5/tengo/compiler/token"

// Break represents a break statement.
type Break struct{}

// TypeName returns the name of the type.
func (o *Break) TypeName() string {
	return "break"
}

func (o *Break) String() string {
	return "<break>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Break) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Break) Copy() Object {
	return &Break{}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Break) IsFalsy() bool {
	return false
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Break) Equals(x Object) bool {
	return false
}
