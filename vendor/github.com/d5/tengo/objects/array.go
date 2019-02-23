package objects

import (
	"fmt"
	"strings"

	"github.com/d5/tengo/compiler/token"
)

// Array represents an array of objects.
type Array struct {
	Value []Object
}

// TypeName returns the name of the type.
func (o *Array) TypeName() string {
	return "array"
}

func (o *Array) String() string {
	var elements []string
	for _, e := range o.Value {
		elements = append(elements, e.String())
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Array) BinaryOp(op token.Token, rhs Object) (Object, error) {
	if rhs, ok := rhs.(*Array); ok {
		switch op {
		case token.Add:
			if len(rhs.Value) == 0 {
				return o, nil
			}
			return &Array{Value: append(o.Value, rhs.Value...)}, nil
		}
	}

	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Array) Copy() Object {
	var c []Object
	for _, elem := range o.Value {
		c = append(c, elem.Copy())
	}

	return &Array{Value: c}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Array) IsFalsy() bool {
	return len(o.Value) == 0
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Array) Equals(x Object) bool {
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
func (o *Array) IndexGet(index Object) (res Object, err error) {
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

// IndexSet sets an element at a given index.
func (o *Array) IndexSet(index, value Object) (err error) {
	intIdx, ok := ToInt(index)
	if !ok {
		err = ErrInvalidIndexType
		return
	}

	if intIdx < 0 || intIdx >= len(o.Value) {
		err = ErrIndexOutOfBounds
		return
	}

	o.Value[intIdx] = value

	return nil
}

// Iterate creates an array iterator.
func (o *Array) Iterate() Iterator {
	return &ArrayIterator{
		v: o.Value,
		l: len(o.Value),
	}
}
