package objects

import (
	"strconv"

	"github.com/d5/tengo"
	"github.com/d5/tengo/compiler/token"
)

// String represents a string value.
type String struct {
	Value   string
	runeStr []rune
}

// TypeName returns the name of the type.
func (o *String) TypeName() string {
	return "string"
}

func (o *String) String() string {
	return strconv.Quote(o.Value)
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *String) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch op {
	case token.Add:
		switch rhs := rhs.(type) {
		case *String:
			if len(o.Value)+len(rhs.Value) > tengo.MaxStringLen {
				return nil, ErrStringLimit
			}
			return &String{Value: o.Value + rhs.Value}, nil
		default:
			rhsStr := rhs.String()
			if len(o.Value)+len(rhsStr) > tengo.MaxStringLen {
				return nil, ErrStringLimit
			}
			return &String{Value: o.Value + rhsStr}, nil
		}
	}

	return nil, ErrInvalidOperator
}

// IsFalsy returns true if the value of the type is falsy.
func (o *String) IsFalsy() bool {
	return len(o.Value) == 0
}

// Copy returns a copy of the type.
func (o *String) Copy() Object {
	return &String{Value: o.Value}
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *String) Equals(x Object) bool {
	t, ok := x.(*String)
	if !ok {
		return false
	}

	return o.Value == t.Value
}

// IndexGet returns a character at a given index.
func (o *String) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(*Int)
	if !ok {
		err = ErrInvalidIndexType
		return
	}

	idxVal := int(intIdx.Value)

	if o.runeStr == nil {
		o.runeStr = []rune(o.Value)
	}

	if idxVal < 0 || idxVal >= len(o.runeStr) {
		res = UndefinedValue
		return
	}

	res = &Char{Value: o.runeStr[idxVal]}

	return
}

// Iterate creates a string iterator.
func (o *String) Iterate() Iterator {
	if o.runeStr == nil {
		o.runeStr = []rune(o.Value)
	}

	return &StringIterator{
		v: o.runeStr,
		l: len(o.runeStr),
	}
}
