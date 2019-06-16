package objects

import (
	"bytes"

	"github.com/d5/tengo"
	"github.com/d5/tengo/compiler/token"
)

// Bytes represents a byte array.
type Bytes struct {
	Value []byte
}

func (o *Bytes) String() string {
	return string(o.Value)
}

// TypeName returns the name of the type.
func (o *Bytes) TypeName() string {
	return "bytes"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Bytes) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch op {
	case token.Add:
		switch rhs := rhs.(type) {
		case *Bytes:
			if len(o.Value)+len(rhs.Value) > tengo.MaxBytesLen {
				return nil, ErrBytesLimit
			}

			return &Bytes{Value: append(o.Value, rhs.Value...)}, nil
		}
	}

	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Bytes) Copy() Object {
	return &Bytes{Value: append([]byte{}, o.Value...)}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Bytes) IsFalsy() bool {
	return len(o.Value) == 0
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Bytes) Equals(x Object) bool {
	t, ok := x.(*Bytes)
	if !ok {
		return false
	}

	return bytes.Equal(o.Value, t.Value)
}

// IndexGet returns an element (as Int) at a given index.
func (o *Bytes) IndexGet(index Object) (res Object, err error) {
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

	res = &Int{Value: int64(o.Value[idxVal])}

	return
}

// Iterate creates a bytes iterator.
func (o *Bytes) Iterate() Iterator {
	return &BytesIterator{
		v: o.Value,
		l: len(o.Value),
	}
}
