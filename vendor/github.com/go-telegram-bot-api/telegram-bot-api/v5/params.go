package tgbotapi

import (
	"encoding/json"
	"reflect"
	"strconv"
)

// Params represents a set of parameters that gets passed to a request.
type Params map[string]string

// AddNonEmpty adds a value if it not an empty string.
func (p Params) AddNonEmpty(key, value string) {
	if value != "" {
		p[key] = value
	}
}

// AddNonZero adds a value if it is not zero.
func (p Params) AddNonZero(key string, value int) {
	if value != 0 {
		p[key] = strconv.Itoa(value)
	}
}

// AddNonZero64 is the same as AddNonZero except uses an int64.
func (p Params) AddNonZero64(key string, value int64) {
	if value != 0 {
		p[key] = strconv.FormatInt(value, 10)
	}
}

// AddBool adds a value of a bool if it is true.
func (p Params) AddBool(key string, value bool) {
	if value {
		p[key] = strconv.FormatBool(value)
	}
}

// AddNonZeroFloat adds a floating point value that is not zero.
func (p Params) AddNonZeroFloat(key string, value float64) {
	if value != 0 {
		p[key] = strconv.FormatFloat(value, 'f', 6, 64)
	}
}

// AddInterface adds an interface if it is not nil and can be JSON marshalled.
func (p Params) AddInterface(key string, value interface{}) error {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return nil
	}

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	p[key] = string(b)

	return nil
}

// AddFirstValid attempts to add the first item that is not a default value.
//
// For example, AddFirstValid(0, "", "test") would add "test".
func (p Params) AddFirstValid(key string, args ...interface{}) error {
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			if v != 0 {
				p[key] = strconv.Itoa(v)
				return nil
			}
		case int64:
			if v != 0 {
				p[key] = strconv.FormatInt(v, 10)
				return nil
			}
		case string:
			if v != "" {
				p[key] = v
				return nil
			}
		case nil:
		default:
			b, err := json.Marshal(arg)
			if err != nil {
				return err
			}

			p[key] = string(b)
			return nil
		}
	}

	return nil
}
