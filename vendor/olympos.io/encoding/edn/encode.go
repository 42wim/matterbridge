// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"encoding/base64"
	"io"
	"math"
	"math/big"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"
)

// Marshal returns the EDN encoding of v.
//
// Marshal traverses the value v recursively.
// If an encountered value implements the Marshaler interface
// and is not a nil pointer, Marshal calls its MarshalEDN method
// to produce EDN.  The nil pointer exception is not strictly necessary
// but mimics a similar, necessary exception in the behavior of
// UnmarshalEDN.
//
// Otherwise, Marshal uses the following type-dependent default encodings:
//
// Boolean values encode as EDN booleans.
//
// Integers encode as EDN integers.
//
// Floating point values encode as EDN floats.
//
// String values encode as EDN strings coerced to valid UTF-8,
// replacing invalid bytes with the Unicode replacement rune.
// The angle brackets "<" and ">" are escaped to "\u003c" and "\u003e"
// to keep some browsers from misinterpreting EDN output as HTML.
// Ampersand "&" is also escaped to "\u0026" for the same reason.
//
// Array and slice values encode as EDN arrays, except that
// []byte encodes as a base64-encoded string, and a nil slice
// encodes as the nil EDN value.
//
// Struct values encode as EDN maps. Each exported struct field
// becomes a member of the map unless
//   - the field's tag is "-", or
//   - the field is empty and its tag specifies the "omitempty" option.
// The empty values are false, 0, any
// nil pointer or interface value, and any array, slice, map, or string of
// length zero. The map's default key is the struct field name as a keyword,
// but can be specified in the struct field's tag value. The "edn" key in
// the struct field's tag value is the key name, followed by an optional comma
// and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `edn:"-"`
//
//   // Field appears in EDN as key :my-name.
//   Field int `edn:"myName"`
//
//   // Field appears in EDN as key :my-name and
//   // the field is omitted from the object if its value is empty,
//   // as defined above.
//   Field int `edn:"my-name,omitempty"`
//
//   // Field appears in EDN as key :field (the default), but
//   // the field is skipped if empty.
//   // Note the leading comma.
//   Field int `edn:",omitempty"`
//
// The "str", "key" and "sym" options signals that a field name should be
// written as a string, keyword or symbol, respectively. If none are specified,
// then the default behaviour is to emit them as keywords. Examples:
//
//    // Default behaviour: field name will be encoded as :foo
//    Foo int
//
//    // Encode Foo as string with name "string-foo"
//    Foo int `edn:"string-foo,str"`
//
//    // Encode Foo as symbol with name sym-foo
//    Foo int `edn:"sym-foo,sym"`
//
// Anonymous struct fields are usually marshaled as if their inner exported fields
// were fields in the outer struct, subject to the usual Go visibility rules amended
// as described in the next paragraph.
// An anonymous struct field with a name given in its EDN tag is treated as
// having that name, rather than being anonymous.
// An anonymous struct field of interface type is treated the same as having
// that type as its name, rather than being anonymous.
//
// The Go visibility rules for struct fields are amended for EDN when
// deciding which field to marshal or unmarshal. If there are
// multiple fields at the same level, and that level is the least
// nested (and would therefore be the nesting level selected by the
// usual Go rules), the following extra rules apply:
//
// 1) Of those fields, if any are EDN-tagged, only tagged fields are considered,
// even if there are multiple untagged fields that would otherwise conflict.
// 2) If there is exactly one field (tagged or not according to the first rule), that is selected.
// 3) Otherwise there are multiple fields, and all are ignored; no error occurs.
//
// To force ignoring of an anonymous struct field in both current and earlier
// versions, give the field a EDN tag of "-".
//
// Map values usually encode as EDN maps. There are no limitations on the keys
// or values -- as long as they can be encoded to EDN, anything goes. Map values
// will be encoded as sets if their value type is either a bool or a struct with
// no fields.
//
// If you want to ensure that a value is encoded as a map, you can specify that
// as follows:
//
//    // Encode Foo as a map, instead of the default set
//    Foo map[int]bool `edn:",map"`
//
// Arrays and slices are encoded as vectors by default. As with maps and sets,
// you can specify that a field should be encoded as a list instead, by using
// the option "list":
//
//    // Encode Foo as a list, instead of the default vector
//    Foo []int `edn:",list"`
//
// Pointer values encode as the value pointed to.
// A nil pointer encodes as the nil EDN object.
//
// Interface values encode as the value contained in the interface.
// A nil interface value encodes as the nil EDN value.
//
// Channel, complex, and function values cannot be encoded in EDN.
// Attempting to encode such a value causes Marshal to return
// an UnsupportedTypeError.
//
// EDN cannot represent cyclic data structures and Marshal does not
// handle them. Passing cyclic structures to Marshal will result in
// an infinite recursion.
//
func Marshal(v interface{}) ([]byte, error) {
	e := &encodeState{}
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalPPrint is like Marshal but applies PPrint to format the output.
func MarshalPPrint(v interface{}, opts *PPrintOpts) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = PPrint(&buf, b, opts)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// An Encoder writes EDN values to an output stream.
type Encoder struct {
	writer io.Writer
	ec     encodeState
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
		ec:     encodeState{},
	}
}

// Encode writes the EDN encoding of v to the stream, followed by a newline
// character.
//
// See the documentation for Marshal for details about the conversion of Go
// values to EDN.
func (e *Encoder) Encode(v interface{}) error {
	e.ec.needsDelim = false
	err := e.ec.marshal(v)
	if err != nil {
		e.ec.Reset()
		return err
	}
	b := e.ec.Bytes()
	e.ec.Reset()
	_, err = e.writer.Write(b)
	if err != nil {
		return err
	}
	_, err = e.writer.Write([]byte{'\n'})
	return err
}

// EncodeIndent writes the indented EDN encoding of v to the stream, followed by
// a newline character.
//
// See the documentation for MarshalIndent for details about the conversion of
// Go values to EDN.
func (e *Encoder) EncodeIndent(v interface{}, prefix, indent string) error {
	e.ec.needsDelim = false
	err := e.ec.marshal(v)
	if err != nil {
		e.ec.Reset()
		return err
	}
	b := e.ec.Bytes()
	var buf bytes.Buffer
	err = Indent(&buf, b, prefix, indent)
	e.ec.Reset()
	if err != nil {
		return err
	}
	_, err = e.writer.Write(buf.Bytes())
	if err != nil {
		return err
	}
	_, err = e.writer.Write([]byte{'\n'})
	return err
}

// EncodePPrint writes the pretty-printed EDN encoding of v to the stream,
// followed by a newline character.
//
// See the documentation for MarshalPPrint for details about the conversion of
// Go values to EDN.
func (e *Encoder) EncodePPrint(v interface{}, opts *PPrintOpts) error {
	e.ec.needsDelim = false
	err := e.ec.marshal(v)
	if err != nil {
		e.ec.Reset()
		return err
	}
	b := e.ec.Bytes()
	var buf bytes.Buffer
	err = PPrint(&buf, b, opts)
	e.ec.Reset()
	if err != nil {
		return err
	}
	_, err = e.writer.Write(buf.Bytes())
	if err != nil {
		return err
	}
	_, err = e.writer.Write([]byte{'\n'})
	return err
}

// Marshaler is the interface implemented by objects that
// can marshal themselves into valid EDN.
type Marshaler interface {
	MarshalEDN() ([]byte, error)
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "edn: unsupported type: " + e.Type.String()
}

// An UnsupportedValueError is returned by Marshal when attempting to encode an
// unsupported value. Examples include the float values NaN and Infinity.
type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "edn: unsupported value: " + e.Str
}

// A MarshalerError is returned by Marshal when encoding a type with a
// MarshalEDN function fails.
type MarshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *MarshalerError) Error() string {
	return "edn: error calling MarshalEDN for type " + e.Type.String() + ": " + e.Err.Error()
}

var hex = "0123456789abcdef"

// An encodeState encodes EDN into a bytes.Buffer.
type encodeState struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte
	needsDelim   bool
	mc           *MathContext
}

// mathContext returns the math context to use. If not set in the encodeState,
// the global math context is used.
func (e *encodeState) mathContext() *MathContext {
	if e.mc != nil {
		return e.mc
	}
	return &GlobalMathContext
}

var encodeStatePool sync.Pool

func newEncodeState() *encodeState {
	if v := encodeStatePool.Get(); v != nil {
		e := v.(*encodeState)
		e.Reset()
		return e
	}
	return new(encodeState)
}

func (e *encodeState) marshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(v))
	return nil
}

func (e *encodeState) error(err error) {
	panic(err)
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (e *encodeState) reflectValue(v reflect.Value) {
	valueEncoder(v)(e, v)
}

type encoderFunc func(e *encodeState, v reflect.Value)

type typeAndTag struct {
	t     reflect.Type
	ctype tagType
}

var encoderCache struct {
	sync.RWMutex
	m map[typeAndTag]encoderFunc
}

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type(), tagUndefined)
}

func typeEncoder(t reflect.Type, tagType tagType) encoderFunc {
	tac := typeAndTag{t, tagType}
	encoderCache.RLock()
	f := encoderCache.m[tac]
	encoderCache.RUnlock()
	if f != nil {
		return f
	}
	couldUseJSON := readCanUseJSONTag()

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it.  This indirect
	// func is only used for recursive types.
	encoderCache.Lock()
	if encoderCache.m == nil {
		encoderCache.m = make(map[typeAndTag]encoderFunc)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	encoderCache.m[tac] = func(e *encodeState, v reflect.Value) {
		wg.Wait()
		f(e, v)
	}
	encoderCache.Unlock()

	// Compute fields without lock.
	// Might duplicate effort but won't hold other computations back.
	f = newTypeEncoder(t, tagType, true)
	wg.Done()
	encoderCache.Lock()
	if couldUseJSON != readCanUseJSONTag() {
		// cache has been invalidated, unlock and retry recursively.
		encoderCache.Unlock()
		return typeEncoder(t, tagType)
	}
	encoderCache.m[tac] = f
	encoderCache.Unlock()
	return f
}

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
	instType      = reflect.TypeOf((*time.Time)(nil)).Elem()
)

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(t reflect.Type, tagType tagType, allowAddr bool) encoderFunc {
	if t.Implements(marshalerType) {
		return marshalerEncoder
	}
	if t.Kind() != reflect.Ptr && allowAddr {
		if reflect.PtrTo(t).Implements(marshalerType) {
			return newCondAddrEncoder(addrMarshalerEncoder, newTypeEncoder(t, tagType, false))
		}
	}

	// Handle specific types first
	switch t {
	case bigIntType:
		return bigIntEncoder
	case bigFloatType:
		return bigFloatEncoder
	case instType:
		return instEncoder
	}

	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int32:
		if tagType == tagRune {
			return runeEncoder
		}
		return intEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(t, tagType)
	case reflect.Map:
		return newMapEncoder(t, tagType)
	case reflect.Slice:
		return newSliceEncoder(t, tagType)
	case reflect.Array:
		return newArrayEncoder(t, tagType)
	case reflect.Ptr:
		return newPtrEncoder(t, tagType)
	default:
		return unsupportedTypeEncoder
	}
}

func invalidValueEncoder(e *encodeState, v reflect.Value) {
	e.writeNil()
}

func marshalerEncoder(e *encodeState, v reflect.Value) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		e.writeNil()
		return
	}
	m := v.Interface().(Marshaler)
	b, err := m.MarshalEDN()
	if err == nil {
		// copy EDN into buffer, checking (token) validity.
		e.ensureDelim()
		err = Compact(&e.Buffer, b)
		e.needsDelim = true
	}
	if err != nil {
		e.error(&MarshalerError{v.Type(), err})
	}
}

func addrMarshalerEncoder(e *encodeState, v reflect.Value) {
	va := v.Addr()
	if va.IsNil() {
		e.writeNil()
		return
	}
	m := va.Interface().(Marshaler)
	b, err := m.MarshalEDN()
	if err == nil {
		// copy EDN into buffer, checking (token) validity.
		e.ensureDelim()
		err = Compact(&e.Buffer, b)
		e.needsDelim = true
	}
	if err != nil {
		e.error(&MarshalerError{v.Type(), err})
	}
}

func boolEncoder(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	if v.Bool() {
		e.WriteString("true")
	} else {
		e.WriteString("false")
	}
	e.needsDelim = true
}

func runeEncoder(e *encodeState, v reflect.Value) {
	encodeRune(&e.Buffer, rune(v.Int()))
	e.needsDelim = true
}

func intEncoder(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	b := strconv.AppendInt(e.scratch[:0], v.Int(), 10)
	e.Write(b)
	e.needsDelim = true
}

func uintEncoder(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	b := strconv.AppendUint(e.scratch[:0], v.Uint(), 10)
	e.Write(b)
	e.needsDelim = true
}

func bigIntEncoder(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	bi := v.Interface().(big.Int)
	b := []byte(bi.String())
	e.Write(b)
	e.WriteByte('N')
	e.needsDelim = true
}

func bigFloatEncoder(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	bf := new(big.Float)
	mc := e.mathContext()
	val := v.Interface().(big.Float)
	bf.Set(&val).SetMode(mc.Mode)
	b := []byte(bf.Text('g', int(mc.Precision)))
	e.Write(b)
	e.WriteByte('M')
	e.needsDelim = true
}

func instEncoder(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	t := v.Interface().(time.Time)
	e.Write([]byte(t.Format(`#inst"` + time.RFC3339Nano + `"`)))
}

type floatEncoder int // number of bits

func (bits floatEncoder) encode(e *encodeState, v reflect.Value) {
	f := v.Float()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		e.error(&UnsupportedValueError{v, strconv.FormatFloat(f, 'g', -1, int(bits))})
	}
	e.ensureDelim()
	b := strconv.AppendFloat(e.scratch[:0], f, 'g', -1, int(bits))
	if ix := bytes.IndexAny(b, ".eE"); ix < 0 {
		b = append(b, '.', '0')
	}
	e.Write(b)
	e.needsDelim = true
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

func stringEncoder(e *encodeState, v reflect.Value) {
	e.string(v.String())
}

func interfaceEncoder(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.writeNil()
		return
	}
	e.reflectValue(v.Elem())
}

func unsupportedTypeEncoder(e *encodeState, v reflect.Value) {
	e.error(&UnsupportedTypeError{v.Type()})
}

type structEncoder struct {
	fields    []field
	fieldEncs []encoderFunc
}

func (se *structEncoder) encode(e *encodeState, v reflect.Value) {
	e.WriteByte('{')
	e.needsDelim = false
	for i, f := range se.fields {
		fv := fieldByIndex(v, f.index)
		if !fv.IsValid() || f.omitEmpty && isEmptyValue(fv) {
			continue
		}
		switch f.fnameType {
		case emitKey:
			e.ensureDelim()
			e.WriteByte(':')
			e.WriteString(f.name)
			e.needsDelim = true
		case emitString:
			e.string(f.name)
			e.needsDelim = false
		case emitSym:
			e.ensureDelim()
			e.WriteString(f.name)
			e.needsDelim = true
		}
		se.fieldEncs[i](e, fv)
	}
	e.WriteByte('}')
	e.needsDelim = false
}

func newStructEncoder(t reflect.Type, tagType tagType) encoderFunc {
	fields := cachedTypeFields(t)
	se := &structEncoder{
		fields:    fields,
		fieldEncs: make([]encoderFunc, len(fields)),
	}
	for i, f := range fields {
		se.fieldEncs[i] = typeEncoder(typeByIndex(t, f.index), f.tagType)
	}
	return se.encode
}

type mapEncoder struct {
	keyEnc  encoderFunc
	elemEnc encoderFunc
}

func (me *mapEncoder) encode(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.writeNil()
		return
	}
	e.WriteByte('{')
	e.needsDelim = false
	mk := v.MapKeys()
	// NB: We don't get deterministic results here, because we don't iterate in a
	// determinstic way.
	for _, k := range mk {
		if e.needsDelim { // bypass conventional whitespace to use commas instead
			e.WriteByte(',')
			e.needsDelim = false
		}
		me.keyEnc(e, k)
		me.elemEnc(e, v.MapIndex(k))
	}
	e.WriteByte('}')
	e.needsDelim = false
}

type mapSetEncoder struct {
	keyEnc encoderFunc
}

func (me *mapSetEncoder) encode(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.writeNil()
		return
	}
	e.ensureDelim()
	e.WriteByte('#')
	e.WriteByte('{')
	e.needsDelim = false
	mk := v.MapKeys()
	// not deterministic this one either.
	for _, k := range mk {
		mval := v.MapIndex(k)
		if mval.Kind() != reflect.Bool || mval.Bool() {
			me.keyEnc(e, k)
		}
	}
	e.WriteByte('}')
	e.needsDelim = false
}

func newMapEncoder(t reflect.Type, tagType tagType) encoderFunc {
	canBeSet := false
	switch t.Elem().Kind() {
	case reflect.Struct:
		if t.Elem().NumField() == 0 {
			canBeSet = true
		}
	case reflect.Bool:
		canBeSet = true
	}
	if (tagType == tagUndefined || tagType == tagSet) && canBeSet {
		me := &mapSetEncoder{typeEncoder(t.Key(), tagUndefined)}
		return me.encode
	}
	if tagType != tagUndefined && tagType != tagMap {
		return unsupportedTypeEncoder
	}
	me := &mapEncoder{
		typeEncoder(t.Key(), tagUndefined),
		typeEncoder(t.Elem(), tagUndefined),
	}
	return me.encode
}

func encodeByteSlice(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.writeNil()
		return
	}
	s := v.Bytes()
	e.ensureDelim()
	e.WriteString(`#base64"`)
	if len(s) < 1024 {
		// for small buffers, using Encode directly is much faster.
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
		base64.StdEncoding.Encode(dst, s)
		e.Write(dst)
	} else {
		// for large buffers, avoid unnecessary extra temporary
		// buffer space.
		enc := base64.NewEncoder(base64.StdEncoding, e)
		enc.Write(s)
		enc.Close()
	}
	e.WriteByte('"')
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (e *encodeState) ensureDelim() {
	if e.needsDelim {
		e.WriteByte(' ')
	}
}

func (e *encodeState) writeNil() {
	e.ensureDelim()
	e.WriteString("nil")
	e.needsDelim = true
}

func (se *sliceEncoder) encode(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.writeNil()
		return
	}
	se.arrayEnc(e, v)
}

func newSliceEncoder(t reflect.Type, tagType tagType) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		return encodeByteSlice
	}
	enc := &sliceEncoder{newArrayEncoder(t, tagType)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae *arrayEncoder) encode(e *encodeState, v reflect.Value) {
	e.WriteByte('[')
	e.needsDelim = false
	n := v.Len()
	for i := 0; i < n; i++ {
		ae.elemEnc(e, v.Index(i))
	}
	e.WriteByte(']')
	e.needsDelim = false
}

type listArrayEncoder struct {
	elemEnc encoderFunc
}

func (ae *listArrayEncoder) encode(e *encodeState, v reflect.Value) {
	e.WriteByte('(')
	e.needsDelim = false
	n := v.Len()
	for i := 0; i < n; i++ {
		ae.elemEnc(e, v.Index(i))
	}
	e.WriteByte(')')
	e.needsDelim = false
}

type setArrayEncoder struct {
	elemEnc encoderFunc
}

func (ae *setArrayEncoder) encode(e *encodeState, v reflect.Value) {
	e.ensureDelim()
	e.WriteByte('#')
	e.WriteByte('{')
	e.needsDelim = false
	n := v.Len()
	for i := 0; i < n; i++ {
		ae.elemEnc(e, v.Index(i))
	}
	e.WriteByte('}')
	e.needsDelim = false
}

func newArrayEncoder(t reflect.Type, tagType tagType) encoderFunc {
	switch tagType {
	case tagList:
		enc := &listArrayEncoder{typeEncoder(t.Elem(), tagUndefined)}
		return enc.encode
	case tagSet:
		enc := &setArrayEncoder{typeEncoder(t.Elem(), tagUndefined)}
		return enc.encode
	default:
		enc := &arrayEncoder{typeEncoder(t.Elem(), tagUndefined)}
		return enc.encode
	}
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe *ptrEncoder) encode(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.writeNil()
		return
	}
	pe.elemEnc(e, v.Elem())
}

func newPtrEncoder(t reflect.Type, tagType tagType) encoderFunc {
	enc := &ptrEncoder{typeEncoder(t.Elem(), tagType)}
	return enc.encode
}

type condAddrEncoder struct {
	canAddrEnc, elseEnc encoderFunc
}

func (ce *condAddrEncoder) encode(e *encodeState, v reflect.Value) {
	if v.CanAddr() {
		ce.canAddrEnc(e, v)
	} else {
		ce.elseEnc(e, v)
	}
}

// newCondAddrEncoder returns an encoder that checks whether its value
// CanAddr and delegates to canAddrEnc if so, else to elseEnc.
func newCondAddrEncoder(canAddrEnc, elseEnc encoderFunc) encoderFunc {
	enc := &condAddrEncoder{canAddrEnc: canAddrEnc, elseEnc: elseEnc}
	return enc.encode
}

// NOTE: keep in sync with stringBytes below.
func (e *encodeState) string(s string) (int, error) {
	len0 := e.Len()
	e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' && b != '<' && b != '>' && b != '&' {
				i++
				continue
			}
			if start < i {
				e.WriteString(s[start:i])
			}
			switch b {
			case '\\', '"':
				e.WriteByte('\\')
				e.WriteByte(b)
			case '\n':
				e.WriteByte('\\')
				e.WriteByte('n')
			case '\r':
				e.WriteByte('\\')
				e.WriteByte('r')
			case '\t':
				e.WriteByte('\\')
				e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \n and \r,
				// as well as <, > and &. The latter are escaped because they
				// can lead to security holes when user-controlled strings
				// are rendered into EDN and served to some browsers.
				e.WriteString(`\u00`)
				e.WriteByte(hex[b>>4])
				e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.WriteString(s[start:i])
			}
			e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.WriteString(s[start:])
	}
	e.WriteByte('"')
	e.needsDelim = false
	return e.Len() - len0, nil
}

// NOTE: keep in sync with string above.
func (e *encodeState) stringBytes(s []byte) (int, error) {
	len0 := e.Len()
	e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' && b != '<' && b != '>' && b != '&' {
				i++
				continue
			}
			if start < i {
				e.Write(s[start:i])
			}
			switch b {
			case '\\', '"':
				e.WriteByte('\\')
				e.WriteByte(b)
			case '\n':
				e.WriteByte('\\')
				e.WriteByte('n')
			case '\r':
				e.WriteByte('\\')
				e.WriteByte('r')
			case '\t':
				e.WriteByte('\\')
				e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \n and \r,
				// as well as <, >, and &. The latter are escaped because they
				// can lead to security holes when user-controlled strings
				// are rendered into EDN and served to some browsers.
				e.WriteString(`\u00`)
				e.WriteByte(hex[b>>4])
				e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.Write(s[start:i])
			}
			e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.Write(s[start:])
	}
	e.WriteByte('"')
	e.needsDelim = false
	return e.Len() - len0, nil
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

// A field represents a single field found in a struct.
type field struct {
	name      string
	nameBytes []byte                 // []byte(name)
	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent

	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	fnameType emitType
	tagType   tagType
}

type emitType int

const (
	emitSym emitType = iota
	emitKey
	emitString
)

type tagType int

const (
	tagUndefined tagType = iota
	tagSet
	tagMap
	tagVec
	tagList
	tagRune
)

func fillField(f field) field {
	f.nameBytes = []byte(f.name)
	f.equalFold = foldFunc(f.nameBytes)
	return f
}

// byName sorts field by name, breaking ties with depth,
// then breaking ties with "name came from edn tag", then
// breaking ties with index sequence.
type byName []field

func (x byName) Len() int { return len(x) }

func (x byName) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byName) Less(i, j int) bool {
	if x[i].name != x[j].name {
		return x[i].name < x[j].name
	}
	if len(x[i].index) != len(x[j].index) {
		return len(x[i].index) < len(x[j].index)
	}
	if x[i].tag != x[j].tag {
		return x[i].tag
	}
	return byIndex(x).Less(i, j)
}

// byIndex sorts field by index sequence.
type byIndex []field

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// typeFields returns a list of fields that edn should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func typeFields(t reflect.Type) []field {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.PkgPath != "" && !sf.Anonymous { // unexported
					continue
				}
				tag := sf.Tag.Get("edn")
				if tag == "" && readCanUseJSONTag() {
					tag = sf.Tag.Get("json")
				}
				if tag == "-" {
					continue
				}
				name, opts := parseTag(tag)
				if !isValidTag(name) {
					name = ""
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Add tagging rules:
				var emit emitType
				switch {
				case opts.Contains("sym"):
					emit = emitSym
				case opts.Contains("str"):
					emit = emitString
				case opts.Contains("key"):
					fallthrough
				default:
					emit = emitKey
				}
				// key, sym, str

				var tagType tagType // add tag rules
				switch {
				case opts.Contains("set"):
					tagType = tagSet
				case opts.Contains("map"):
					tagType = tagMap
				case opts.Contains("vector"):
					tagType = tagVec
				case opts.Contains("list"):
					tagType = tagList
				case opts.Contains("rune"):
					tagType = tagRune
				default:
					tagType = tagUndefined
				}

				// Record found field and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						r := []rune(sf.Name)
						r[0] = unicode.ToLower(r[0])
						name = string(r)
					}
					fields = append(fields, fillField(field{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						fnameType: emit,
						tagType:   tagType,
					}))
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, fillField(field{name: ft.Name(), index: index, typ: ft}))
				}
			}
		}
	}

	sort.Sort(byName(fields))

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with EDN tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byIndex(fields))

	return fields
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// EDN tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order. The winner
	// must therefore be one with the shortest index length. Drop all
	// longer entries, which is easy: just truncate the slice.
	length := len(fields[0].index)
	tagged := -1 // Index of first tagged field.
	for i, f := range fields {
		if len(f.index) > length {
			fields = fields[:i]
			break
		}
		if f.tag {
			if tagged >= 0 {
				// Multiple tagged fields at the same level: conflict.
				// Return no field.
				return field{}, false
			}
			tagged = i
		}
	}
	if tagged >= 0 {
		return fields[tagged], true
	}
	// All remaining fields have the same length. If there's more than one,
	// we have a conflict (two fields named "X" at the same level) and we
	// return no field.
	if len(fields) > 1 {
		return field{}, false
	}
	return fields[0], true
}

var canUseJSONTag int32

func readCanUseJSONTag() bool {
	return atomic.LoadInt32(&canUseJSONTag) == 1
}

// UseJSONAsFallback can be set to true to let go-edn parse structs with
// information from the `json` tag for encoding and decoding type fields if not
// the `edn` tag field is set. This is not threadsafe: Encoding and decoding
// happening while this is called may return results that mix json and non-json
// tag reading. Preferably you call this in an init() function to ensure it is
// either set or unset.
func UseJSONAsFallback(val bool) {
	set := int32(0)
	if val {
		set = 1
	}

	// Here comes the funny stuff: Cache invalidation. Right now we lock and
	// unlock these independently of eachother, so it's fine to lock them in this
	// order. However, if we decide to change this later on, the only reasonable
	// change would be that you may grab the encoderCache lock before the
	// fieldCache lock. Therefore we do it in this order, although it should not
	// matter strictly speaking.
	encoderCache.Lock()
	fieldCache.Lock()
	atomic.StoreInt32(&canUseJSONTag, set)
	fieldCache.m = nil
	encoderCache.m = nil
	fieldCache.Unlock()
	encoderCache.Unlock()
}

var fieldCache struct {
	sync.RWMutex
	m map[reflect.Type][]field
}

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) []field {
	fieldCache.RLock()
	f := fieldCache.m[t]
	fieldCache.RUnlock()
	if f != nil {
		return f
	}
	couldUseJSON := readCanUseJSONTag()

	// Compute fields without lock.
	// Might duplicate effort but won't hold other computations back.
	f = typeFields(t)
	if f == nil {
		f = []field{}
	}

	fieldCache.Lock()
	if couldUseJSON != readCanUseJSONTag() {
		// cache has been invalidated, unlock and retry recursively.
		fieldCache.Unlock()
		return cachedTypeFields(t)
	}
	if fieldCache.m == nil {
		fieldCache.m = map[reflect.Type][]field{}
	}
	fieldCache.m[t] = f
	fieldCache.Unlock()
	return f
}
