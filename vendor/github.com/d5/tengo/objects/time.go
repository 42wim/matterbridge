package objects

import (
	"time"

	"github.com/d5/tengo/compiler/token"
)

// Time represents a time value.
type Time struct {
	Value time.Time
}

func (o *Time) String() string {
	return o.Value.String()
}

// TypeName returns the name of the type.
func (o *Time) TypeName() string {
	return "time"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Time) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch rhs := rhs.(type) {
	case *Int:
		switch op {
		case token.Add: // time + int => time
			if rhs.Value == 0 {
				return o, nil
			}
			return &Time{Value: o.Value.Add(time.Duration(rhs.Value))}, nil
		case token.Sub: // time - int => time
			if rhs.Value == 0 {
				return o, nil
			}
			return &Time{Value: o.Value.Add(time.Duration(-rhs.Value))}, nil
		}
	case *Time:
		switch op {
		case token.Sub: // time - time => int (duration)
			return &Int{Value: int64(o.Value.Sub(rhs.Value))}, nil
		case token.Less: // time < time => bool
			if o.Value.Before(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value.After(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value.Equal(rhs.Value) || o.Value.Before(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value.Equal(rhs.Value) || o.Value.After(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}

	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Time) Copy() Object {
	return &Time{Value: o.Value}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Time) IsFalsy() bool {
	return o.Value.IsZero()
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Time) Equals(x Object) bool {
	t, ok := x.(*Time)
	if !ok {
		return false
	}

	return o.Value.Equal(t.Value)
}
