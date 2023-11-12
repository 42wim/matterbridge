package bencode

import (
	"io"
	"math/big"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"sync"

	"github.com/anacrolix/missinggo"
)

func isEmptyValue(v reflect.Value) bool {
	return missinggo.IsEmptyValue(v)
}

type Encoder struct {
	w       io.Writer
	scratch [64]byte
}

func (e *Encoder) Encode(v interface{}) (err error) {
	if v == nil {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(runtime.Error); ok {
				panic(e)
			}
			var ok bool
			err, ok = e.(error)
			if !ok {
				panic(e)
			}
		}
	}()
	e.reflectValue(reflect.ValueOf(v))
	return nil
}

type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

func (e *Encoder) write(s []byte) {
	_, err := e.w.Write(s)
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeString(s string) {
	for s != "" {
		n := copy(e.scratch[:], s)
		s = s[n:]
		e.write(e.scratch[:n])
	}
}

func (e *Encoder) reflectString(s string) {
	e.writeStringPrefix(int64(len(s)))
	e.writeString(s)
}

func (e *Encoder) writeStringPrefix(l int64) {
	b := strconv.AppendInt(e.scratch[:0], l, 10)
	e.write(b)
	e.writeString(":")
}

func (e *Encoder) reflectByteSlice(s []byte) {
	e.writeStringPrefix(int64(len(s)))
	e.write(s)
}

// Returns true if the value implements Marshaler interface and marshaling was
// done successfully.
func (e *Encoder) reflectMarshaler(v reflect.Value) bool {
	if !v.Type().Implements(marshalerType) {
		if v.Kind() != reflect.Ptr && v.CanAddr() && v.Addr().Type().Implements(marshalerType) {
			v = v.Addr()
		} else {
			return false
		}
	}
	m := v.Interface().(Marshaler)
	data, err := m.MarshalBencode()
	if err != nil {
		panic(&MarshalerError{v.Type(), err})
	}
	e.write(data)
	return true
}

var bigIntType = reflect.TypeOf((*big.Int)(nil)).Elem()

func (e *Encoder) reflectValue(v reflect.Value) {
	if e.reflectMarshaler(v) {
		return
	}

	if v.Type() == bigIntType {
		e.writeString("i")
		bi := v.Interface().(big.Int)
		e.writeString(bi.String())
		e.writeString("e")
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			e.writeString("i1e")
		} else {
			e.writeString("i0e")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.writeString("i")
		b := strconv.AppendInt(e.scratch[:0], v.Int(), 10)
		e.write(b)
		e.writeString("e")
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.writeString("i")
		b := strconv.AppendUint(e.scratch[:0], v.Uint(), 10)
		e.write(b)
		e.writeString("e")
	case reflect.String:
		e.reflectString(v.String())
	case reflect.Struct:
		e.writeString("d")
		for _, ef := range getEncodeFields(v.Type()) {
			fieldValue := ef.i(v)
			if !fieldValue.IsValid() {
				continue
			}
			if ef.omitEmpty && isEmptyValue(fieldValue) {
				continue
			}
			e.reflectString(ef.tag)
			e.reflectValue(fieldValue)
		}
		e.writeString("e")
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			panic(&MarshalTypeError{v.Type()})
		}
		if v.IsNil() {
			e.writeString("de")
			break
		}
		e.writeString("d")
		sv := stringValues(v.MapKeys())
		sort.Sort(sv)
		for _, key := range sv {
			e.reflectString(key.String())
			e.reflectValue(v.MapIndex(key))
		}
		e.writeString("e")
	case reflect.Slice, reflect.Array:
		e.reflectSequence(v)
	case reflect.Interface:
		e.reflectValue(v.Elem())
	case reflect.Ptr:
		if v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		} else {
			v = v.Elem()
		}
		e.reflectValue(v)
	default:
		panic(&MarshalTypeError{v.Type()})
	}
}

func (e *Encoder) reflectSequence(v reflect.Value) {
	// Use bencode string-type
	if v.Type().Elem().Kind() == reflect.Uint8 {
		if v.Kind() != reflect.Slice {
			// Can't use []byte optimization
			if !v.CanAddr() {
				e.writeStringPrefix(int64(v.Len()))
				for i := 0; i < v.Len(); i++ {
					var b [1]byte
					b[0] = byte(v.Index(i).Uint())
					e.write(b[:])
				}
				return
			}
			v = v.Slice(0, v.Len())
		}
		s := v.Bytes()
		e.reflectByteSlice(s)
		return
	}
	if v.IsNil() {
		e.writeString("le")
		return
	}
	e.writeString("l")
	for i, n := 0, v.Len(); i < n; i++ {
		e.reflectValue(v.Index(i))
	}
	e.writeString("e")
}

type encodeField struct {
	i         func(v reflect.Value) reflect.Value
	tag       string
	omitEmpty bool
}

type encodeFieldsSortType []encodeField

func (ef encodeFieldsSortType) Len() int           { return len(ef) }
func (ef encodeFieldsSortType) Swap(i, j int)      { ef[i], ef[j] = ef[j], ef[i] }
func (ef encodeFieldsSortType) Less(i, j int) bool { return ef[i].tag < ef[j].tag }

var (
	typeCacheLock     sync.RWMutex
	encodeFieldsCache = make(map[reflect.Type][]encodeField)
)

func getEncodeFields(t reflect.Type) []encodeField {
	typeCacheLock.RLock()
	fs, ok := encodeFieldsCache[t]
	typeCacheLock.RUnlock()
	if ok {
		return fs
	}
	fs = makeEncodeFields(t)
	typeCacheLock.Lock()
	defer typeCacheLock.Unlock()
	encodeFieldsCache[t] = fs
	return fs
}

func makeEncodeFields(t reflect.Type) (fs []encodeField) {
	for _i, n := 0, t.NumField(); _i < n; _i++ {
		i := _i
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if f.Anonymous {
			t := f.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			anonEFs := makeEncodeFields(t)
			for aefi := range anonEFs {
				anonEF := anonEFs[aefi]
				bottomField := anonEF
				bottomField.i = func(v reflect.Value) reflect.Value {
					v = v.Field(i)
					if v.Kind() == reflect.Ptr {
						if v.IsNil() {
							// This will skip serializing this value.
							return reflect.Value{}
						}
						v = v.Elem()
					}
					return anonEF.i(v)
				}
				fs = append(fs, bottomField)
			}
			continue
		}
		var ef encodeField
		ef.i = func(v reflect.Value) reflect.Value {
			return v.Field(i)
		}
		ef.tag = f.Name

		tv := getTag(f.Tag)
		if tv.Ignore() {
			continue
		}
		if tv.Key() != "" {
			ef.tag = tv.Key()
		}
		ef.omitEmpty = tv.OmitEmpty()
		fs = append(fs, ef)
	}
	fss := encodeFieldsSortType(fs)
	sort.Sort(fss)
	return fs
}
