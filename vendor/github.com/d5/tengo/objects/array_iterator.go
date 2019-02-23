package objects

import "github.com/d5/tengo/compiler/token"

// ArrayIterator is an iterator for an array.
type ArrayIterator struct {
	v []Object
	i int
	l int
}

// TypeName returns the name of the type.
func (i *ArrayIterator) TypeName() string {
	return "array-iterator"
}

func (i *ArrayIterator) String() string {
	return "<array-iterator>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (i *ArrayIterator) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// IsFalsy returns true if the value of the type is falsy.
func (i *ArrayIterator) IsFalsy() bool {
	return true
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (i *ArrayIterator) Equals(Object) bool {
	return false
}

// Copy returns a copy of the type.
func (i *ArrayIterator) Copy() Object {
	return &ArrayIterator{v: i.v, i: i.i, l: i.l}
}

// Next returns true if there are more elements to iterate.
func (i *ArrayIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

// Key returns the key or index value of the current element.
func (i *ArrayIterator) Key() Object {
	return &Int{Value: int64(i.i - 1)}
}

// Value returns the value of the current element.
func (i *ArrayIterator) Value() Object {
	return i.v[i.i-1]
}
