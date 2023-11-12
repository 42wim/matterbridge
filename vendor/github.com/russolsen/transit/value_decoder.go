// Copyright 2016 Russ Olsen. All Rights Reserved.
//
// This code is a Go port of the Java version created and maintained by Cognitect, therefore:
//
// Copyright 2014 Cognitect. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transit

import (
	"container/list"
	"encoding/base64"
	"github.com/pborman/uuid"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"strconv"
	"time"
)

// DecodeKeyword decodes ~: style keywords.
func DecodeKeyword(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	var result = Keyword(s)
	return result, nil
}

// DecodeKeyword decodes ~$ style symbols.
func DecodeSymbol(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	var result = Symbol(s)
	return result, nil
}

// DecodeIdentity returns the value unchanged.
func DecodeIdentity(d Decoder, x interface{}) (interface{}, error) {
	return x, nil
}

// DecodeCMap decodes maps with composite keys.
func DecodeCMap(d Decoder, x interface{}) (interface{}, error) {

	tagged := x.(TaggedValue)

	if !IsGenericArray(tagged.Value) {
		return nil, NewTransitError("Cmap contents are not an array.", tagged)
	}

	array := tagged.Value.([]interface{})

	if (len(array) % 2) != 0 {
		return nil, NewTransitError("Cmap contents must contain an even number of elements.", tagged)
	}

	var result = NewCMap()

	l := len(array)

	for i := 0; i < l; i += 2 {
		key := array[i]
		value := array[i+1]
		result.Append(key, value)
	}

	return result, nil
}

// DecodeSet decodes a transit set into a transit.Set instance.
func DecodeSet(d Decoder, x interface{}) (interface{}, error) {
	tagged := x.(TaggedValue)
	if !IsGenericArray(tagged.Value) {
		return nil, NewTransitError("Set contents are not an array.", tagged)
	}
	values := (tagged.Value).([]interface{})
	result := NewSet(values)
	return result, nil
}

// DecodeList decodes a transit list into a Go list.
func DecodeList(d Decoder, x interface{}) (interface{}, error) {
	tagged := x.(TaggedValue)
	if !IsGenericArray(tagged.Value) {
		return nil, NewTransitError("List contents are not an array.", tagged)
	}
	values := (tagged.Value).([]interface{})
	result := list.New()
	for _, item := range values {
		result.PushBack(item)
	}
	return result, nil
}

// DecodeQuote decodes a transit quoted value by simply returning the value.
func DecodeQuote(d Decoder, x interface{}) (interface{}, error) {
	tagged := x.(TaggedValue)
	return tagged.Value, nil
}

// DecodeRFC3339 decodes a time value into a Go time instance.
// TBD not 100% this covers all possible values.
func DecodeRFC3339(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	var result, err = time.Parse(time.RFC3339Nano, s)
	return result, err
}

// DecodeTime decodes a time value represended as millis since 1970.
func DecodeTime(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	var millis, _ = strconv.ParseInt(s, 10, 64)
	seconds := millis / 1000
	remainder_millis := millis - (seconds * 1000)
	nanos := remainder_millis * 1000000
	result := time.Unix(seconds, nanos).UTC()
	return result, nil
}

// DecodeBoolean decodes a transit boolean into a Go bool.
func DecodeBoolean(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	if s == "t" {
		return true, nil
	} else if s == "f" {
		return false, nil
	} else {
		return nil, &TransitError{Message: "Unknown boolean value."}
	}
}

// DecodeBigInteger decodes a transit big integer into a Go big.Int.
func DecodeBigInteger(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	result := new(big.Int)
	_, good := result.SetString(s, 10)
	if !good {
		return nil, &TransitError{Message: "Unable to part big integer: " + s}
	}
	return result, nil
}

// DecodeInteger decodes a transit integer into a plain Go int64
func DecodeInteger(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	result, err := strconv.ParseInt(s, 10, 64)
	return result, err
}

func newRational(a, b *big.Int) *big.Rat {
	var r = big.NewRat(1, 1)
	r.SetFrac(a, b)
	return r
}

func toBigInt(x interface{}) (*big.Int, error) {
	switch v := x.(type) {
	default:
		return nil, NewTransitError("Not a numeric value", v)
	case *big.Int:
		return v, nil
	case int64:
		return big.NewInt(v), nil
	}
}

// DecodeRatio decodes a transit ratio into a Go big.Rat.
func DecodeRatio(d Decoder, x interface{}) (interface{}, error) {
	tagged := x.(TaggedValue)
	if !IsGenericArray(tagged.Value) {
		return nil, NewTransitError("Ratio contents are not an array.", tagged)
	}

	values := (tagged.Value).([]interface{})

	if len(values) != 2 {
		return nil, NewTransitError("Ratio contents does not contain 2 elements.", tagged)
	}

	a, err := toBigInt(values[0])

	if err != nil {
		return nil, err
	}

	b, err := toBigInt(values[1])

	if err != nil {
		return nil, err
	}

	result := newRational(a, b)
	return *result, nil
}

// DecodeRune decodes a transit char.
func DecodeRune(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	return rune(s[0]), nil
}

// DecodeFloat decodes the value into a float.
func DecodeFloat(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	return strconv.ParseFloat(s, 64)
}

// DecodeDecimal decodes a transit big decimal into decimal.Decimal.
func DecodeDecimal(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	return decimal.NewFromString((s))
}

// DecodeRatio decodes a transit null/nil.
func DecodeNil(d Decoder, x interface{}) (interface{}, error) {
	return nil, nil
}

// DecodeRatio decodes a transit base64 encoded byte array into a
// Go byte array.
func DecodeByte(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	return base64.StdEncoding.DecodeString(s)
}

// DecodeLink decodes a transit link into an instance of Link.
func DecodeLink(d Decoder, x interface{}) (interface{}, error) {
	tv := x.(TaggedValue)
	v := tv.Value.(map[interface{}]interface{})
	l := NewLink()
	l.Href = v["href"].(*TUri)
	l.Name = v["name"].(string)
	l.Rel = v["rel"].(string)
	l.Prompt = v["prompt"].(string)
	l.Render = v["render"].(string)
	return l, nil
}

// DecodeURI decodes a transit URI into an instance of TUri.
func DecodeURI(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	return NewTUri(s), nil
}

// DecodeUUID decodes a transit UUID into an instance of net/UUID
func DecodeUUID(d Decoder, x interface{}) (interface{}, error) {
	s := x.(string)
	var u = uuid.Parse(s)
	if u == nil {
		return nil, &TransitError{Message: "Unable to parse uuid [" + s + "]"}
	}
	return u, nil
}

// DecodeSpecialNumber decodes NaN, INF and -INF into their Go equivalents.
func DecodeSpecialNumber(d Decoder, x interface{}) (interface{}, error) {
	tag := x.(string)
	if tag == "NaN" {
		return math.NaN(), nil
	} else if tag == "INF" {
		return math.Inf(1), nil
	} else if tag == "-INF" {
		return math.Inf(-1), nil
	} else {
		return nil, &TransitError{Message: "Bad special number:"}
	}
}
