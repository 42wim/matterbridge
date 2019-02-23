package objects

import "github.com/d5/tengo/compiler/token"

// MapIterator represents an iterator for the map.
type MapIterator struct {
	v map[string]Object
	k []string
	i int
	l int
}

// TypeName returns the name of the type.
func (i *MapIterator) TypeName() string {
	return "map-iterator"
}

func (i *MapIterator) String() string {
	return "<map-iterator>"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (i *MapIterator) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// IsFalsy returns true if the value of the type is falsy.
func (i *MapIterator) IsFalsy() bool {
	return true
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (i *MapIterator) Equals(Object) bool {
	return false
}

// Copy returns a copy of the type.
func (i *MapIterator) Copy() Object {
	return &MapIterator{v: i.v, k: i.k, i: i.i, l: i.l}
}

// Next returns true if there are more elements to iterate.
func (i *MapIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

// Key returns the key or index value of the current element.
func (i *MapIterator) Key() Object {
	k := i.k[i.i-1]

	return &String{Value: k}
}

// Value returns the value of the current element.
func (i *MapIterator) Value() Object {
	k := i.k[i.i-1]

	return i.v[k]
}
