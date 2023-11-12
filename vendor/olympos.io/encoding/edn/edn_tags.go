// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package edn implements encoding and decoding of EDN values as defined in
// https://github.com/edn-format/edn. For a full introduction on how to use
// go-edn, see https://github.com/go-edn/edn/blob/v1/docs/introduction.md. Fully
// self-contained examples of go-edn can be found at
// https://github.com/go-edn/edn/tree/v1/examples.
//
// Note that the small examples in this package is not checking errors as
// persively as you should do when you use this package. This is done because
// I'd like the examples to be easily readable and understandable. The bigger
// examples provide proper error handling.
package edn

import (
	"encoding/base64"
	"errors"
	"math/big"
	"reflect"
	"sync"
	"time"
)

var (
	ErrNotFunc         = errors.New("Value is not a function")
	ErrMismatchArities = errors.New("Function does not have single argument in, two argument out")
	ErrNotConcrete     = errors.New("Value is not a concrete non-function type")
	ErrTagOverwritten  = errors.New("Previous tag implementation was overwritten")
)

var globalTags TagMap

// A TagMap contains mappings from tag literals to functions and structs that is
// used when decoding.
type TagMap struct {
	sync.RWMutex
	m map[string]reflect.Value
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// AddTagFn adds fn as a converter function for tagname tags to this TagMap. fn
// must have the signature func(T) (U, error), where T is the expected input
// type and U is the output type. See Decoder.AddTagFn for examples.
func (tm *TagMap) AddTagFn(tagname string, fn interface{}) error {
	// TODO: check name
	rfn := reflect.ValueOf(fn)
	rtyp := rfn.Type()
	if rtyp.Kind() != reflect.Func {
		return ErrNotFunc
	}
	if rtyp.NumIn() != 1 || rtyp.NumOut() != 2 || !rtyp.Out(1).Implements(errorType) {
		// ok to have variadic arity?
		return ErrMismatchArities
	}
	return tm.addVal(tagname, rfn)
}

// MustAddTagFn adds fn as a converter function for tagname tags to this TagMap
// like AddTagFn, except this function panics if the tag could not be added.
func (tm *TagMap) MustAddTagFn(tagname string, fn interface{}) {
	if err := tm.AddTagFn(tagname, fn); err != nil {
		panic(err)
	}
}

func (tm *TagMap) addVal(name string, val reflect.Value) error {
	tm.Lock()
	if tm.m == nil {
		tm.m = map[string]reflect.Value{}
	}
	_, ok := tm.m[name]
	tm.m[name] = val
	tm.Unlock()
	if ok {
		return ErrTagOverwritten
	} else {
		return nil
	}
}

// AddTagFn adds fn as a converter function for tagname tags to the global
// TagMap. fn must have the signature func(T) (U, error), where T is the
// expected input type and U is the output type. See Decoder.AddTagFn for
// examples.
func AddTagFn(tagname string, fn interface{}) error {
	return globalTags.AddTagFn(tagname, fn)
}

// MustAddTagFn adds fn as a converter function for tagname tags to the global
// TagMap like AddTagFn, except this function panics if the tag could not be added.
func MustAddTagFn(tagname string, fn interface{}) {
	globalTags.MustAddTagFn(tagname, fn)
}

// AddTagStructs adds the struct as a matching struct for tagname tags to this
// TagMap. val can not be a channel, function, interface or an unsafe pointer.
// See Decoder.AddTagStruct for examples.
func (tm *TagMap) AddTagStruct(tagname string, val interface{}) error {
	rstruct := reflect.ValueOf(val)
	switch rstruct.Type().Kind() {
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return ErrNotConcrete
	}
	return tm.addVal(tagname, rstruct)
}

// AddTagStructs adds the struct as a matching struct for tagname tags to the
// global TagMap. val can not be a channel, function, interface or an unsafe
// pointer. See Decoder.AddTagStruct for examples.
func AddTagStruct(tagname string, val interface{}) error {
	return globalTags.AddTagStruct(tagname, val)
}

func init() {
	err := AddTagFn("inst", func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339Nano, s)
	})
	if err != nil {
		panic(err)
	}
	err = AddTagFn("base64", base64.StdEncoding.DecodeString)
	if err != nil {
		panic(err)
	}
}

// A MathContext specifies the precision and rounding mode for
// `math/big.Float`s when decoding.
type MathContext struct {
	Precision uint
	Mode      big.RoundingMode
}

// The GlobalMathContext is the global MathContext. It is used if no other
// context is provided. See MathContext for example usage.
var GlobalMathContext = MathContext{
	Mode:      big.ToNearestEven,
	Precision: 192,
}
