package objects

import (
	"fmt"
	"strings"

	"github.com/d5/tengo/compiler/token"
)

// ImmutableMap represents an immutable map object.
type ImmutableMap struct {
	Value map[string]Object
}

// TypeName returns the name of the type.
func (o *ImmutableMap) TypeName() string {
	return "immutable-map"
}

func (o *ImmutableMap) String() string {
	var pairs []string
	for k, v := range o.Value {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v.String()))
	}

	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *ImmutableMap) BinaryOp(op token.Token, rhs Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *ImmutableMap) Copy() Object {
	c := make(map[string]Object)
	for k, v := range o.Value {
		c[k] = v.Copy()
	}

	return &Map{Value: c}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *ImmutableMap) IsFalsy() bool {
	return len(o.Value) == 0
}

// IndexGet returns the value for the given key.
func (o *ImmutableMap) IndexGet(index Object) (res Object, err error) {
	strIdx, ok := ToString(index)
	if !ok {
		err = ErrInvalidIndexType
		return
	}

	val, ok := o.Value[strIdx]
	if !ok {
		val = UndefinedValue
	}

	return val, nil
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *ImmutableMap) Equals(x Object) bool {
	var xVal map[string]Object
	switch x := x.(type) {
	case *Map:
		xVal = x.Value
	case *ImmutableMap:
		xVal = x.Value
	default:
		return false
	}

	if len(o.Value) != len(xVal) {
		return false
	}

	for k, v := range o.Value {
		tv := xVal[k]
		if !v.Equals(tv) {
			return false
		}
	}

	return true
}

// Iterate creates an immutable map iterator.
func (o *ImmutableMap) Iterate() Iterator {
	var keys []string
	for k := range o.Value {
		keys = append(keys, k)
	}

	return &MapIterator{
		v: o.Value,
		k: keys,
		l: len(keys),
	}
}
