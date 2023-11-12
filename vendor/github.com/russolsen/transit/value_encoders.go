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
	"fmt"
	"github.com/pborman/uuid"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"time"
)

// ValueEncoder is the interface for objects that know how to
// transit encode a single value.
type ValueEncoder interface {
	IsStringable(reflect.Value) bool
	Encode(e Encoder, value reflect.Value, asString bool) error
}

type NilEncoder struct{}

func NewNilEncoder() *NilEncoder {
	return &NilEncoder{}
}

func (ie NilEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie NilEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	return e.emitter.EmitNil(asKey)
}

type RuneEncoder struct{}

func NewRuneEncoder() *RuneEncoder {
	return &RuneEncoder{}
}

func (ie RuneEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie RuneEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	r := v.Interface().(rune)
	return e.emitter.EmitString(fmt.Sprintf("~c%c", r), asKey)
}

type PointerEncoder struct{}

func NewPointerEncoder() *PointerEncoder {
	return &PointerEncoder{}
}

func (ie PointerEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie PointerEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	//log.Println("*** Defer pointer to:", v.Elem())
	return e.EncodeInterface(v.Elem().Interface(), asKey)
}

type UuidEncoder struct{}

func NewUuidEncoder() *UuidEncoder {
	return &UuidEncoder{}
}

func (ie UuidEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie UuidEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	u := v.Interface().(uuid.UUID)
	return e.emitter.EmitString(fmt.Sprintf("~u%v", u.String()), asKey)
}

type TimeEncoder struct{}

func NewTimeEncoder() *TimeEncoder {
	return &TimeEncoder{}
}

func (ie TimeEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie TimeEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	t := v.Interface().(time.Time)
	nanos := t.UnixNano()
	millis := nanos / int64(1000000)
	//millis := t.Unix() * 1000
	return e.emitter.EmitString(fmt.Sprintf("~m%d", millis), asKey)
}

type BoolEncoder struct{}

func NewBoolEncoder() *BoolEncoder {
	return &BoolEncoder{}
}

func (ie BoolEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie BoolEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	b := v.Bool()
	return e.emitter.EmitBool(b, asKey)
}

type FloatEncoder struct{}

func NewFloatEncoder() *FloatEncoder {
	return &FloatEncoder{}
}

func (ie FloatEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie FloatEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	f := v.Float()
	if math.IsNaN(f) {
		return e.emitter.EmitString("~zNaN", asKey)
	} else if math.IsInf(f, 1) {
		return e.emitter.EmitString("~zINF", asKey)
	} else if math.IsInf(f, -1) {
		return e.emitter.EmitString("~z-INF", asKey)
	} else {
		return e.emitter.EmitFloat(f, asKey)
	}
}

type DecimalEncoder struct{}

func NewDecimalEncoder() *DecimalEncoder {
	return &DecimalEncoder{}
}

func (ie DecimalEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie DecimalEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	f := v.Interface().(decimal.Decimal)
	return e.emitter.EmitString(fmt.Sprintf("~f%v", f.String()), asKey)
}

type BigFloatEncoder struct{}

func NewBigFloatEncoder() *BigFloatEncoder {
	return &BigFloatEncoder{}
}

func (ie BigFloatEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie BigFloatEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	f := v.Interface().(big.Float)
	return e.emitter.EmitString(fmt.Sprintf("~f%v", f.Text('f', 25)), asKey)
}

type BigIntEncoder struct{}

func NewBigIntEncoder() *BigIntEncoder {
	return &BigIntEncoder{}
}

func (ie BigIntEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie BigIntEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	i := v.Interface().(big.Int)
	return e.emitter.EmitString(fmt.Sprintf("~n%v", i.String()), asKey)
}

type BigRatEncoder struct{}

func NewBigRatEncoder() *BigRatEncoder {
	return &BigRatEncoder{}
}

func (ie BigRatEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie BigRatEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	r := v.Interface().(big.Rat)

	e.emitter.EmitStartArray()
	e.emitter.EmitTag("ratio")
	e.emitter.EmitArraySeparator()

	e.emitter.EmitStartArray()
	e.Encode(r.Num())
	e.emitter.EmitArraySeparator()
	e.Encode(r.Denom())
	e.emitter.EmitEndArray()

	return e.emitter.EmitEndArray()
}

type IntEncoder struct{}

func NewIntEncoder() *IntEncoder {
	return &IntEncoder{}
}

func (ie IntEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie IntEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	return e.emitter.EmitInt(v.Int(), asKey)
}

type UintEncoder struct{}

func NewUintEncoder() *UintEncoder {
	return &UintEncoder{}
}

func (ie UintEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie UintEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	return e.emitter.EmitInt(int64(v.Uint()), asKey)
}

type KeywordEncoder struct{}

func NewKeywordEncoder() *KeywordEncoder {
	return &KeywordEncoder{}
}

func (ie KeywordEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie KeywordEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	s := v.String()
	//log.Println("Encoding keyword:", s)
	return e.emitter.EmitString("~:"+s, true)
}

type SymbolEncoder struct{}

func NewSymbolEncoder() *SymbolEncoder {
	return &SymbolEncoder{}
}

func (ie SymbolEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie SymbolEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	s := v.String()
	//log.Println("Encoding symbol:", s)
	return e.emitter.EmitString("~$"+s, true)
}

type StringEncoder struct{}

func NewStringEncoder() *StringEncoder {
	return &StringEncoder{}
}

func (ie StringEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie StringEncoder) needsEscape(s string) bool {
	if len(s) == 0 {
		return false
	}

	firstCh := s[0:1]

	return firstCh == start || firstCh == reserved || firstCh == sub
}

func (ie StringEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	s := v.String()
	if ie.needsEscape(s) {
		s = "~" + s
	}
	return e.emitter.EmitString(s, asKey)
}

type UrlEncoder struct{}

func NewUrlEncoder() *UrlEncoder {
	return &UrlEncoder{}
}

func (ie UrlEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie UrlEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	u := v.Interface().(*url.URL)
	us := u.String()
	return e.emitter.EmitString(fmt.Sprintf("~r%s", us), asKey)
}

type TUriEncoder struct{}

func NewTUriEncoder() *TUriEncoder {
	return &TUriEncoder{}
}

func (ie TUriEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie TUriEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	u := v.Interface().(*TUri)
	return e.emitter.EmitString(fmt.Sprintf("~r%s", u.Value), asKey)
}

type ErrorEncoder struct{}

func NewErrorEncoder() *ErrorEncoder {
	return &ErrorEncoder{}
}

func (ie ErrorEncoder) IsStringable(v reflect.Value) bool {
	return true
}

func (ie ErrorEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	return NewTransitError("Dont know how to encode value", v)
}

type ArrayEncoder struct{}

func NewArrayEncoder() *ArrayEncoder {
	return &ArrayEncoder{}
}

func (ie ArrayEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie ArrayEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	e.emitter.EmitStartArray()

	l := v.Len()
	for i := 0; i < l; i++ {
		if i > 0 {
			e.emitter.EmitArraySeparator()
		}
		element := v.Index(i)
		err := e.EncodeInterface(element.Interface(), asKey)
		if err != nil {
			return err
		}
	}

	return e.emitter.EmitEndArray()
}

type MapEncoder struct {
	verbose bool
}

func NewMapEncoder(verbose bool) *MapEncoder {
	return &MapEncoder{verbose}
}

func (me MapEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (me MapEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	keys := KeyValues(v)

	if !me.allStringable(e, keys) {
		return me.encodeCompositeMap(e, v)
	} else if me.verbose {
		return me.encodeVerboseMap(e, v)
	} else {
		return me.encodeNormalMap(e, v)
	}
}

func (me MapEncoder) allStringable(e Encoder, keys []reflect.Value) bool {
	for _, key := range keys {
		valueEncoder := e.ValueEncoderFor(reflect.ValueOf(key.Interface()))
		if !valueEncoder.IsStringable(key) {
			return false
		}
	}
	return true
}

func (me MapEncoder) encodeCompositeMap(e Encoder, v reflect.Value) error {
	e.emitter.EmitStartArray()

	e.emitter.EmitTag("cmap")
	e.emitter.EmitArraySeparator()

	e.emitter.EmitStartArray()

	keys := KeyValues(v)

	for i, key := range keys {
		if i != 0 {
			e.emitter.EmitArraySeparator()
		}

		err := e.EncodeValue(key, false)

		if err != nil {
			return err
		}

		e.emitter.EmitArraySeparator()

		value := GetMapElement(v, key)
		err = e.EncodeValue(value, false)

		if err != nil {
			return err
		}
	}

	e.emitter.EmitEndArray()
	return e.emitter.EmitEndArray()
}

func (me MapEncoder) encodeNormalMap(e Encoder, v reflect.Value) error {
	//l := v.Len()
	e.emitter.EmitStartArray()

	e.emitter.EmitString("^ ", false)

	keys := KeyValues(v)

	for _, key := range keys {
		e.emitter.EmitArraySeparator()

		err := e.EncodeValue(key, true)

		if err != nil {
			return err
		}

		e.emitter.EmitArraySeparator()

		value := GetMapElement(v, key)

		err = e.EncodeValue(value, false)

		if err != nil {
			return err
		}
	}

	return e.emitter.EmitEndArray()
}

func (me MapEncoder) encodeVerboseMap(e Encoder, v reflect.Value) error {
	e.emitter.EmitStartMap()

	keys := KeyValues(v)

	for i, key := range keys {
		if i != 0 {
			e.emitter.EmitMapSeparator()
		}

		err := e.EncodeValue(key, true)

		if err != nil {
			return err
		}

		e.emitter.EmitKeySeparator()

		value := GetMapElement(v, key)

		err = e.EncodeValue(value, false)

		if err != nil {
			return err
		}
	}

	return e.emitter.EmitEndMap()
}

type TaggedValueEncoder struct{}

func NewTaggedValueEncoder() *TaggedValueEncoder {
	return &TaggedValueEncoder{}
}

func (ie TaggedValueEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie TaggedValueEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	t := v.Interface().(TaggedValue)

	e.emitter.EmitStartArray()
	e.emitter.EmitTag(string(t.Tag))
	e.emitter.EmitArraySeparator()
	e.EncodeInterface(t.Value, asKey)
	return e.emitter.EmitEndArray()
}

type SetEncoder struct{}

func NewSetEncoder() *SetEncoder {
	return &SetEncoder{}
}

func (ie SetEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie SetEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	s := v.Interface().(Set)

	//log.Println("*** Encode set:", v)

	//l := v.Len()
	e.emitter.EmitStartArray()
	e.emitter.EmitTag("set")
	e.emitter.EmitArraySeparator()

	e.emitter.EmitStartArray()

	for i, element := range s.Contents {
		if i != 0 {
			e.emitter.EmitArraySeparator()
		}
		err := e.EncodeInterface(element, asKey)
		if err != nil {
			return err
		}
	}

	e.emitter.EmitEndArray()

	return e.emitter.EmitEndArray()
}

type ListEncoder struct{}

func NewListEncoder() *ListEncoder {
	return &ListEncoder{}
}

func (ie ListEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie ListEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	lst := v.Interface().(*list.List)

	e.emitter.EmitStartArray()
	e.emitter.EmitTag("list")
	e.emitter.EmitArraySeparator()
	e.emitter.EmitStartArray()

	first := true
	for element := lst.Front(); element != nil; element = element.Next() {
		if first {
			first = false
		} else {
			e.emitter.EmitArraySeparator()
		}

		err := e.EncodeInterface(element.Value, asKey)
		if err != nil {
			log.Println("ERROR", err)
			return err
		}
	}

	e.emitter.EmitEndArray()
	return e.emitter.EmitEndArray()
}

type CMapEncoder struct{}

func NewCMapEncoder() *CMapEncoder {
	return &CMapEncoder{}
}

func (ie CMapEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie CMapEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	cmap := v.Interface().(*CMap)

	//l := v.Len()
	e.emitter.EmitStartArray()
	e.emitter.EmitTag("cmap")
	e.emitter.EmitArraySeparator()
	e.emitter.EmitStartArray()

	for i, entry := range cmap.Entries {
		if i != 0 {
			e.emitter.EmitArraySeparator()
		}

		err := e.EncodeInterface(entry.Key, false)
		if err != nil {
			return err
		}

		e.emitter.EmitArraySeparator()

		err = e.EncodeInterface(entry.Value, false)
		if err != nil {
			return err
		}
	}

	e.emitter.EmitEndArray()
	return e.emitter.EmitEndArray()
}

type LinkEncoder struct{}

func NewLinkEncoder() *LinkEncoder {
	return &LinkEncoder{}
}

func (ie LinkEncoder) IsStringable(v reflect.Value) bool {
	return false
}

func (ie LinkEncoder) Encode(e Encoder, v reflect.Value, asKey bool) error {
	link := v.Interface().(*Link)

	e.emitter.EmitStartArray()
	e.emitter.EmitTag("link")
	e.emitter.EmitArraySeparator()

	m := map[string]interface{}{
		"href":   link.Href,
		"rel":    link.Rel,
		"name":   link.Name,
		"prompt": link.Prompt,
		"render": link.Render,
	}

	e.Encode(m)
	return e.emitter.EmitEndArray()
}
