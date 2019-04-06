// A modified version of Go's JSON implementation.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"encoding/base64"
	"errors"
	"math"
	"strconv"

	"github.com/d5/tengo/objects"
)

// Encode returns the JSON encoding of the object.
func Encode(o objects.Object) ([]byte, error) {
	var b []byte

	switch o := o.(type) {
	case *objects.Array:
		b = append(b, '[')
		len1 := len(o.Value) - 1
		for idx, elem := range o.Value {
			eb, err := Encode(elem)
			if err != nil {
				return nil, err
			}
			b = append(b, eb...)
			if idx < len1 {
				b = append(b, ',')
			}
		}
		b = append(b, ']')
	case *objects.ImmutableArray:
		b = append(b, '[')
		len1 := len(o.Value) - 1
		for idx, elem := range o.Value {
			eb, err := Encode(elem)
			if err != nil {
				return nil, err
			}
			b = append(b, eb...)
			if idx < len1 {
				b = append(b, ',')
			}
		}
		b = append(b, ']')
	case *objects.Map:
		b = append(b, '{')
		len1 := len(o.Value) - 1
		idx := 0
		for key, value := range o.Value {
			b = strconv.AppendQuote(b, key)
			b = append(b, ':')
			eb, err := Encode(value)
			if err != nil {
				return nil, err
			}
			b = append(b, eb...)
			if idx < len1 {
				b = append(b, ',')
			}
			idx++
		}
		b = append(b, '}')
	case *objects.ImmutableMap:
		b = append(b, '{')
		len1 := len(o.Value) - 1
		idx := 0
		for key, value := range o.Value {
			b = strconv.AppendQuote(b, key)
			b = append(b, ':')
			eb, err := Encode(value)
			if err != nil {
				return nil, err
			}
			b = append(b, eb...)
			if idx < len1 {
				b = append(b, ',')
			}
			idx++
		}
		b = append(b, '}')
	case *objects.Bool:
		if o.IsFalsy() {
			b = strconv.AppendBool(b, false)
		} else {
			b = strconv.AppendBool(b, true)
		}
	case *objects.Bytes:
		b = append(b, '"')
		encodedLen := base64.StdEncoding.EncodedLen(len(o.Value))
		dst := make([]byte, encodedLen)
		base64.StdEncoding.Encode(dst, o.Value)
		b = append(b, dst...)
		b = append(b, '"')
	case *objects.Char:
		b = strconv.AppendInt(b, int64(o.Value), 10)
	case *objects.Float:
		var y []byte

		f := o.Value
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, errors.New("unsupported float value")
		}

		// Convert as if by ES6 number to string conversion.
		// This matches most other JSON generators.
		abs := math.Abs(f)
		fmt := byte('f')
		if abs != 0 {
			if abs < 1e-6 || abs >= 1e21 {
				fmt = 'e'
			}
		}
		y = strconv.AppendFloat(y, f, fmt, -1, 64)
		if fmt == 'e' {
			// clean up e-09 to e-9
			n := len(y)
			if n >= 4 && y[n-4] == 'e' && y[n-3] == '-' && y[n-2] == '0' {
				y[n-2] = y[n-1]
				y = y[:n-1]
			}
		}

		b = append(b, y...)
	case *objects.Int:
		b = strconv.AppendInt(b, o.Value, 10)
	case *objects.String:
		b = strconv.AppendQuote(b, o.Value)
	case *objects.Time:
		y, err := o.Value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b = append(b, y...)
	case *objects.Undefined:
		b = append(b, "null"...)
	default:
		// unknown type: ignore
	}

	return b, nil
}
