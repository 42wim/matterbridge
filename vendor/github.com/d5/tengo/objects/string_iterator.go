package objects

import "github.com/d5/tengo/compiler/token"

// StringIterator represents an iterator for a string.
type StringIterator struct {
	v []rune
	i int
	l int
}

// TypeName returns the name of the type.
func (i *StringIterator) TypeName() string {
	return "string-iterator"
}

func (i *StringIterator) String() string {
	return "<string-iterator>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (i *StringIterator) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// IsFalsy returns true if the value of the type is falsy.
func (i *StringIterator) IsFalsy() bool {
	return true
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (i *StringIterator) Equals(Object) bool {
	return false
}

// Copy returns a copy of the type.
func (i *StringIterator) Copy() Object {
	return &StringIterator{v: i.v, i: i.i, l: i.l}
}

// Next returns true if there are more elements to iterate.
func (i *StringIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

// Key returns the key or index value of the current element.
func (i *StringIterator) Key() Object {
	return &Int{Value: int64(i.i - 1)}
}

// Value returns the value of the current element.
func (i *StringIterator) Value() Object {
	return &Char{Value: i.v[i.i-1]}
}
