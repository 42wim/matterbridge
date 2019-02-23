package objects

import "github.com/d5/tengo/compiler/token"

// Continue represents a continue statement.
type Continue struct {
}

// TypeName returns the name of the type.
func (o *Continue) TypeName() string {
	return "continue"
}

func (o *Continue) String() string {
	return "<continue>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Continue) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Continue) Copy() Object {
	return &Continue{}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Continue) IsFalsy() bool {
	return false
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Continue) Equals(x Object) bool {
	return false
}
