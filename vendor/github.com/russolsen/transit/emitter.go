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
	"fmt"
	"io"
	"strings"
)

type DataEmitter interface {
	Emit(s string) error
	EmitString(s string, cacheable bool) error
	EmitTag(s string) error
	EmitInt(i int64, asKey bool) error
	EmitFloat(f float64, asKey bool) error
	EmitNil(asKey bool) error
	EmitBool(bool, asKey bool) error
	EmitStartArray() error
	EmitArraySeparator() error
	EmitEndArray() error

	EmitStartMap() error
	EmitMapSeparator() error
	EmitKeySeparator() error
	EmitEndMap() error
}

type JsonEmitter struct {
	writer io.Writer
	cache  Cache
}

func NewJsonEmitter(w io.Writer, cache Cache) *JsonEmitter {
	return &JsonEmitter{writer: w, cache: cache}
}

// Emit the string unaltered and without quotes. This is the lowest level emitter.

func (je JsonEmitter) Emit(s string) error {
	_, err := je.writer.Write([]byte(s))
	return err
}

// EmitBase emits the basic value supplied, encoding it as JSON.

func (je JsonEmitter) EmitBase(x interface{}) error {
	bytes, err := json.Marshal(x)
	if err == nil {
		_, err = je.writer.Write(bytes)
	}
	return err
}

// EmitsTag emits a transit #tag. The string supplied should not include the '#'.

func (je JsonEmitter) EmitTag(s string) error {
	return je.EmitString("~#"+s, true)
}

func (je JsonEmitter) EmitString(s string, cacheable bool) error {
	if je.cache.IsCacheable(s, cacheable) {
		s = je.cache.Write(s)
	}
	return je.EmitBase(s)
}

const MaxJsonInt = 1<<53 - 1

func (je JsonEmitter) EmitInt(i int64, asKey bool) error {
	if asKey || (i > MaxJsonInt) {
		return je.EmitString(fmt.Sprintf("~i%d", i), asKey)
	}
	return je.EmitBase(i)
}

func (je JsonEmitter) EmitNil(asKey bool) error {
	if asKey {
		return je.EmitString("~_", false)
	} else {
		return je.EmitBase(nil)
	}
}

func (je JsonEmitter) EmitFloat(f float64, asKey bool) error {
	if asKey {
		return je.EmitString(fmt.Sprintf("~d%g", f), asKey)
	} else {
		s := fmt.Sprintf("%g", f)
		if !strings.ContainsAny(s, ".eE") {
			s = s + ".0" // Horible hack!
		}
		return je.Emit(s)
	}
}

func (je JsonEmitter) EmitStartArray() error {
	return je.Emit("[")
}

func (je JsonEmitter) EmitEndArray() error {
	return je.Emit("]")
}

func (je JsonEmitter) EmitArraySeparator() error {
	return je.Emit(",")
}

func (je JsonEmitter) EmitStartMap() error {
	return je.Emit("{")
}

func (je JsonEmitter) EmitEndMap() error {
	return je.Emit("}")
}

func (je JsonEmitter) EmitMapSeparator() error {
	return je.Emit(",")
}

func (je JsonEmitter) EmitKeySeparator() error {
	return je.Emit(":")
}

func (je JsonEmitter) EmitBool(x bool, asKey bool) error {
	if asKey {
		if x {
			return je.EmitString("~?t", false)
		} else {
			return je.EmitString("~?f", false)
		}
	} else {
		return je.EmitBase(x)
	}
}
