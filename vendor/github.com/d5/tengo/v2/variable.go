package tengo

import (
	"errors"
)

// Variable is a user-defined variable for the script.
type Variable struct {
	name  string
	value Object
}

// NewVariable creates a Variable.
func NewVariable(name string, value interface{}) (*Variable, error) {
	obj, err := FromInterface(value)
	if err != nil {
		return nil, err
	}
	return &Variable{
		name:  name,
		value: obj,
	}, nil
}

// Name returns the name of the variable.
func (v *Variable) Name() string {
	return v.name
}

// Value returns an empty interface of the variable value.
func (v *Variable) Value() interface{} {
	return ToInterface(v.value)
}

// ValueType returns the name of the value type.
func (v *Variable) ValueType() string {
	return v.value.TypeName()
}

// Int returns int value of the variable value.
// It returns 0 if the value is not convertible to int.
func (v *Variable) Int() int {
	c, _ := ToInt(v.value)
	return c
}

// Int64 returns int64 value of the variable value. It returns 0 if the value
// is not convertible to int64.
func (v *Variable) Int64() int64 {
	c, _ := ToInt64(v.value)
	return c
}

// Float returns float64 value of the variable value. It returns 0.0 if the
// value is not convertible to float64.
func (v *Variable) Float() float64 {
	c, _ := ToFloat64(v.value)
	return c
}

// Char returns rune value of the variable value. It returns 0 if the value is
// not convertible to rune.
func (v *Variable) Char() rune {
	c, _ := ToRune(v.value)
	return c
}

// Bool returns bool value of the variable value. It returns 0 if the value is
// not convertible to bool.
func (v *Variable) Bool() bool {
	c, _ := ToBool(v.value)
	return c
}

// Array returns []interface value of the variable value. It returns 0 if the
// value is not convertible to []interface.
func (v *Variable) Array() []interface{} {
	switch val := v.value.(type) {
	case *Array:
		var arr []interface{}
		for _, e := range val.Value {
			arr = append(arr, ToInterface(e))
		}
		return arr
	}
	return nil
}

// Map returns map[string]interface{} value of the variable value. It returns
// 0 if the value is not convertible to map[string]interface{}.
func (v *Variable) Map() map[string]interface{} {
	switch val := v.value.(type) {
	case *Map:
		kv := make(map[string]interface{})
		for mk, mv := range val.Value {
			kv[mk] = ToInterface(mv)
		}
		return kv
	}
	return nil
}

// String returns string value of the variable value. It returns 0 if the value
// is not convertible to string.
func (v *Variable) String() string {
	c, _ := ToString(v.value)
	return c
}

// Bytes returns a byte slice of the variable value. It returns nil if the
// value is not convertible to byte slice.
func (v *Variable) Bytes() []byte {
	c, _ := ToByteSlice(v.value)
	return c
}

// Error returns an error if the underlying value is error object. If not,
// this returns nil.
func (v *Variable) Error() error {
	err, ok := v.value.(*Error)
	if ok {
		return errors.New(err.String())
	}
	return nil
}

// Object returns an underlying Object of the variable value. Note that
// returned Object is a copy of an actual Object used in the script.
func (v *Variable) Object() Object {
	return v.value
}

// IsUndefined returns true if the underlying value is undefined.
func (v *Variable) IsUndefined() bool {
	return v.value == UndefinedValue
}
