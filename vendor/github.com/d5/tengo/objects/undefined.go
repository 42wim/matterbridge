package objects

import "github.com/d5/tengo/compiler/token"

// Undefined represents an undefined value.
type Undefined struct{}

// TypeName returns the name of the type.
func (o *Undefined) TypeName() string {
	return "undefined"
}

func (o *Undefined) String() string {
	return "<undefined>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Undefined) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Undefined) Copy() Object {
	return o
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Undefined) IsFalsy() bool {
	return true
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Undefined) Equals(x Object) bool {
	return o == x
}

// IndexGet returns an element at a given index.
func (o *Undefined) IndexGet(index Object) (Object, error) {
	return UndefinedValue, nil
}

// Iterate creates a map iterator.
func (o *Undefined) Iterate() Iterator {
	return o
}

// Next returns true if there are more elements to iterate.
func (o *Undefined) Next() bool {
	return false
}

// Key returns the key or index value of the current element.
func (o *Undefined) Key() Object {
	return o
}

// Value returns the value of the current element.
func (o *Undefined) Value() Object {
	return o
}
