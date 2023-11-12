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
	"encoding/json"
	"io"
	"strings"
)

type Handler func(Decoder, interface{}) (interface{}, error)

type Decoder struct {
	jsd      *json.Decoder
	decoders map[string]Handler
	cache    *RollingCache
}

// NewDecoder returns a new Decoder, ready to read from r.
func NewDecoder(r io.Reader) *Decoder {
	jsd := json.NewDecoder(r)
	return NewJsonDecoder(jsd)
}

// NewDecoder returns a new Decoder, ready to read from jsr.
func NewJsonDecoder(jsd *json.Decoder) *Decoder {
	jsd.UseNumber()

	decoders := make(map[string]Handler)

	d := Decoder{jsd: jsd, decoders: decoders, cache: NewRollingCache()}
	initHandlers(&d)

	return &d
}

func initHandlers(d *Decoder) {
	d.AddHandler("_", DecodeNil)
	d.AddHandler(":", DecodeKeyword)
	d.AddHandler("?", DecodeBoolean)
	d.AddHandler("b", DecodeByte)
	d.AddHandler("d", DecodeFloat)
	d.AddHandler("i", DecodeInteger)
	d.AddHandler("n", DecodeBigInteger)
	d.AddHandler("f", DecodeDecimal)
	d.AddHandler("c", DecodeRune)
	d.AddHandler("$", DecodeSymbol)
	d.AddHandler("t", DecodeRFC3339)
	d.AddHandler("m", DecodeTime)
	d.AddHandler("u", DecodeUUID)
	d.AddHandler("r", DecodeURI)
	d.AddHandler("'", DecodeQuote)
	d.AddHandler("z", DecodeSpecialNumber)

	d.AddHandler("set", DecodeSet)
	d.AddHandler("link", DecodeLink)
	d.AddHandler("list", DecodeList)
	d.AddHandler("cmap", DecodeCMap)
	d.AddHandler("ratio", DecodeRatio)
	d.AddHandler("unknown", DecodeIdentity)

}

// AddHandler adds a new handler to the decoder, allowing you to extend the types it can handle.
func (d Decoder) AddHandler(tag string, valueDecoder Handler) {
	d.decoders[tag] = valueDecoder
}

func (d Decoder) parseString(s string) (interface{}, error) {

	if d.cache.HasKey(s) {
		return d.Parse(d.cache.Read(s), false)

	} else if !strings.HasPrefix(s, start) {
		return s, nil

	} else if strings.HasPrefix(s, startTag) {
		return TagId(s[2:]), nil

	} else if vd := d.decoders[s[1:2]]; vd != nil {
		return vd(d, s[2:])

	} else if strings.HasPrefix(s, escapeTag) ||
		strings.HasPrefix(s, escapeSub) ||
		strings.HasPrefix(s, escapeRes) {
		return s[1:], nil

	} else {
		tv := TaggedValue{TagId(s[1:2]), s[2:]}
		return d.decoders["unknown"](d, tv)
	}
}

func (d Decoder) parseSingleEntryMap(m map[string]interface{}) (interface{}, error) {
	// The loop here is just a convenient way to get at the only
	// entry in the map.
	for k, v := range m {
		key, err := d.Parse(k, true)
		if err != nil {
			return nil, err
		}

		value, err := d.Parse(v, true)
		if err != nil {
			return nil, err
		}

		if tag, isTag := key.(TagId); isTag {
			tv := TaggedValue{Tag: tag, Value: value}
			valueDecoder := d.DecoderFor(tag)
			return valueDecoder(d, tv)
		} else {
			return map[interface{}]interface{}{key: value}, nil
		}
	}

	return nil, nil // Should never get here
}

func (d Decoder) parseMultiEntryMap(m map[string]interface{}) (interface{}, error) {
	var result = make(map[interface{}]interface{})

	for k, v := range m {
		key, err := d.Parse(k, true)
		if err != nil {
			return nil, err
		}

		value, err := d.Parse(v, false)
		if err != nil {
			return nil, err
		}

		result[key] = value
	}

	return result, nil
}

func (d Decoder) parseMap(m map[string]interface{}) (interface{}, error) {
	if len(m) != 1 {
		return d.parseMultiEntryMap(m)
	} else {
		return d.parseSingleEntryMap(m)
	}
}

func (d Decoder) parseNormalArray(x []interface{}) (interface{}, error) {
	var result = make([]interface{}, len(x))

	for i, v := range x {
		var err error
		result[i], err = d.Parse(v, false)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (d Decoder) parseCMap(x []interface{}) (interface{}, error) {
	var result = NewCMap()

	l := len(x)

	for i := 1; i < l; i += 2 {
		key, err := d.Parse(x[i], true)
		if err != nil {
			return nil, err
		}

		value, err := d.Parse(x[i+1], false)
		if err != nil {
			return nil, err
		}
		result.Append(key, value)
	}

	return result, nil
}

func (d Decoder) parseArrayMap(x []interface{}) (interface{}, error) {
	result := make(map[interface{}]interface{})

	l := len(x)

	for i := 1; i < l; i += 2 {
		key, err := d.Parse(x[i], true)
		if err != nil {
			return nil, err
		}

		value, err := d.Parse(x[i+1], false)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, nil
}

func (d Decoder) DecoderFor(tagid TagId) Handler {
	key := string(tagid)

	handler := d.decoders[key]
	if handler == nil {
		handler = d.decoders["unknown"]
	}
	return handler
}

func (d Decoder) parseArray(x []interface{}) (interface{}, error) {
	if len(x) == 0 {
		return x, nil
	}

	e0, err := d.Parse(x[0], false)

	if err != nil {
		return nil, err
	}

	if e0 == mapAsArray {
		return d.parseArrayMap(x)
	}

	if tagId, isTag := e0.(TagId); isTag {
		var value interface{}

		if value, err = d.Parse(x[1], false); err != nil {
			return nil, err
		}

		tv := TaggedValue{Tag: tagId, Value: value}
		valueDecoder := d.DecoderFor(tagId)
		return valueDecoder(d, tv)
	}

	return d.parseNormalArray(x)
}

func (d Decoder) parseNumber(x json.Number) (interface{}, error) {
	var s = x.String()
	var err error

	var result interface{}

	if strings.ContainsAny(s, ".Ee") {
		result, err = x.Float64()
	} else {
		result, err = x.Int64()
	}

	return result, err
}

func (d Decoder) Parse(x interface{}, asKey bool) (interface{}, error) {

	switch v := x.(type) {
	default:
		return nil, &TransitError{Message: "Unexpected type"}
	case nil:
		return v, nil

	case bool:
		return v, nil

	case json.Number:
		return d.parseNumber(v)

	case string:
		result, err := d.parseString(v)

		if err == nil && d.cache.IsCacheable(v, asKey) {
			d.cache.Write(v)
		}
		return result, err

	case map[string]interface{}:
		return d.parseMap(v)

	case []interface{}:
		return d.parseArray(v)
	}
}

// Decode decodes the next Transit value from the stream.
func (d Decoder) Decode() (interface{}, error) {
	var jsonObject interface{}
	var err = d.jsd.Decode(&jsonObject)

	if err != nil {
		return nil, err
	} else {
		return d.Parse(jsonObject, false)
	}
}

// DecodeFromString is a handly function that decodes Transit data held in a string.
func DecodeFromString(s string) (interface{}, error) {
	reader := strings.NewReader(s)
	decoder := NewDecoder(reader)
	return decoder.Decode()
}
