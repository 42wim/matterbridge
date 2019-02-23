package objects

import (
	"math"
	"strconv"

	"github.com/d5/tengo/compiler/token"
)

// Float represents a floating point number value.
type Float struct {
	Value float64
}

func (o *Float) String() string {
	return strconv.FormatFloat(o.Value, 'f', -1, 64)
}

// TypeName returns the name of the type.
func (o *Float) TypeName() string {
	return "float"
}

// BinaryOp returns another object that is the result of
// a given binary operator and a right-hand side object.
func (o *Float) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch rhs := rhs.(type) {
	case *Float:
		switch op {
		case token.Add:
			r := o.Value + rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Sub:
			r := o.Value - rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Mul:
			r := o.Value * rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Quo:
			r := o.Value / rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Less:
			if o.Value < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case *Int:
		switch op {
		case token.Add:
			r := o.Value + float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Sub:
			r := o.Value - float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Mul:
			r := o.Value * float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Quo:
			r := o.Value / float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Less:
			if o.Value < float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}

	return nil, ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *Float) Copy() Object {
	return &Float{Value: o.Value}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Float) IsFalsy() bool {
	return math.IsNaN(o.Value)
}

// Equals returns true if the value of the type
// is equal to the value of another object.
func (o *Float) Equals(x Object) bool {
	t, ok := x.(*Float)
	if !ok {
		return false
	}

	return o.Value == t.Value
}
