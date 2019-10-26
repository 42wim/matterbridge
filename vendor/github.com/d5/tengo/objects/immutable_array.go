package objects

import (
	"fmt"
	"strings"

	"github.com/d5/tengo/compiler/token"
)

// ImmutableArray represents an immutable array of objects.
type ImmutableArray struct {
	Value []Object
}

// TypeName returns the name of the type.
func (o *ImmutableArray) TypeName() string {
	return "immutable-array"
}

func (o *ImmutableArray) String() string {
	var elements []string
	for _, e := range o.Value {
		elements = append(elements, e.String())
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *ImmutableArray) BinaryOp(op token.Token, rhs Object) (Object, error) {
	if rhs, ok := rhs.(*ImmutableArray); ok {
		switch op {
		case token.Add:
			return &Array{Value: append(o.Value, rhs.Value...)}, nil
		}
	}

	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *ImmutableArray) Copy() Object {
	var c []Object
	for _, elem := range o.Value {
		c = append(c, elem.Copy())
	}

	return &Array{Value: c}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *ImmutableArray) IsFalsy() bool {
	return len(o.Value) == 0
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *ImmutableArray) Equals(x Object) bool {
	var xVal []Object
	switch x := x.(type) {
	case *Array:
		xVal = x.Value
	case *ImmutableArray:
		xVal = x.Value
	default:
		return false
	}

	if len(o.Value) != len(xVal) {
		return false
	}

	for i, e := range o.Value {
		if !e.Equals(xVal[i]) {
			return false
		}
	}

	return true
}

// IndexGet returns an element at a given index.
func (o *ImmutableArray) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(*Int)
	if !ok {
		err = ErrInvalidIndexType
		return
	}

	idxVal := int(intIdx.Value)

	if idxVal < 0 || idxVal >= len(o.Value) {
		res = UndefinedValue
		return
	}

	res = o.Value[idxVal]

	return
}

// Iterate creates an array iterator.
func (o *ImmutableArray) Iterate() Iterator {
	return &ArrayIterator{
		v: o.Value,
		l: len(o.Value),
	}
}
