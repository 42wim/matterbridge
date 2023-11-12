// Copyright (c) 2018 David Crawshaw <david@zentus.com>
// Copyright (c) 2021 Ross Light <ross@zombiezen.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
//
// SPDX-License-Identifier: ISC

package sqlite

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"modernc.org/libc"
	"modernc.org/libc/sys/types"
	lib "modernc.org/sqlite/lib"
)

var auxdata struct {
	mu  sync.RWMutex
	m   map[uintptr]interface{}
	ids idGen
}

// Context is a SQL function execution context.
// It is in no way related to a Go context.Context.
// https://sqlite.org/c3ref/context.html
type Context struct {
	tls *libc.TLS
	ptr uintptr
}

// Conn returns the database connection that is calling the SQL function.
func (ctx Context) Conn() *Conn {
	connPtr := lib.Xsqlite3_context_db_handle(ctx.tls, ctx.ptr)
	allConns.mu.RLock()
	defer allConns.mu.RUnlock()
	return allConns.table[connPtr]
}

// AuxData returns the auxiliary data associated with the given argument, with
// zero being the leftmost argument, or nil if no such data is present.
//
// Auxiliary data may be used by (non-aggregate) SQL functions to associate
// metadata with argument values. If the same value is passed to multiple
// invocations of the same SQL function during query execution, under some
// circumstances the associated metadata may be preserved. An example of where
// this might be useful is in a regular-expression matching function. The
// compiled version of the regular expression can be stored as metadata
// associated with the pattern string. Then as long as the pattern string
// remains the same, the compiled regular expression can be reused on multiple
// invocations of the same function.
//
// For more details, see https://www.sqlite.org/c3ref/get_auxdata.html
func (ctx Context) AuxData(arg int) interface{} {
	id := lib.Xsqlite3_get_auxdata(ctx.tls, ctx.ptr, int32(arg))
	if id == 0 {
		return nil
	}
	auxdata.mu.RLock()
	defer auxdata.mu.RUnlock()
	return auxdata.m[id]
}

// SetAuxData sets the auxiliary data associated with the given argument, with
// zero being the leftmost argument. SQLite is free to discard the metadata at
// any time, including during the call to SetAuxData.
//
// Auxiliary data may be used by (non-aggregate) SQL functions to associate
// metadata with argument values. If the same value is passed to multiple
// invocations of the same SQL function during query execution, under some
// circumstances the associated metadata may be preserved. An example of where
// this might be useful is in a regular-expression matching function. The
// compiled version of the regular expression can be stored as metadata
// associated with the pattern string. Then as long as the pattern string
// remains the same, the compiled regular expression can be reused on multiple
// invocations of the same function.
//
// For more details, see https://www.sqlite.org/c3ref/get_auxdata.html
func (ctx Context) SetAuxData(arg int, data interface{}) {
	auxdata.mu.Lock()
	id := auxdata.ids.next()
	if auxdata.m == nil {
		auxdata.m = make(map[uintptr]interface{})
	}
	auxdata.m[id] = data
	auxdata.mu.Unlock()

	// The following is a conversion from function value to uintptr. It assumes
	// the memory representation described in https://golang.org/s/go11func.
	//
	// It does this by doing the following in order:
	// 1) Create a Go struct containing a pointer to a pointer to
	//    freeAuxData. It is assumed that the pointer to freeAuxData will be
	//    stored in the read-only data section and thus will not move.
	// 2) Convert the pointer to the Go struct to a pointer to uintptr through
	//    unsafe.Pointer. This is permitted via Rule #1 of unsafe.Pointer.
	// 3) Dereference the pointer to uintptr to obtain the function value as a
	//    uintptr. This is safe as long as function values are passed as pointers.
	deleteFn := *(*uintptr)(unsafe.Pointer(&struct {
		f func(*libc.TLS, uintptr)
	}{freeAuxData}))

	lib.Xsqlite3_set_auxdata(ctx.tls, ctx.ptr, int32(arg), id, deleteFn)
}

func freeAuxData(tls *libc.TLS, id uintptr) {
	auxdata.mu.Lock()
	defer auxdata.mu.Unlock()
	delete(auxdata.m, id)
	auxdata.ids.reclaim(id)
}

func (ctx Context) result(v Value, err error) {
	if err != nil {
		ctx.resultError(err)
		return
	}
	if v.tls != nil {
		if ctx.tls != v.tls {
			ctx.resultError(fmt.Errorf("function result Value from different connection"))
			return
		}
		lib.Xsqlite3_result_value(ctx.tls, ctx.ptr, v.ptrOrType)
		return
	}
	switch ColumnType(v.ptrOrType) {
	case 0, TypeNull:
		lib.Xsqlite3_result_null(ctx.tls, ctx.ptr)
	case TypeInteger:
		lib.Xsqlite3_result_int64(ctx.tls, ctx.ptr, v.n)
	case TypeFloat:
		lib.Xsqlite3_result_double(ctx.tls, ctx.ptr, v.float())
	case TypeText:
		if len(v.s) == 0 {
			lib.Xsqlite3_result_text(ctx.tls, ctx.ptr, emptyCString, 0, sqliteStatic)
		} else {
			cv, err := libc.CString(v.s)
			if err != nil {
				ctx.resultError(fmt.Errorf("alloc function result: %w", err))
				return
			}
			lib.Xsqlite3_result_text(ctx.tls, ctx.ptr, cv, int32(len(v.s)), freeFuncPtr)
		}
	case TypeBlob:
		if len(v.s) == 0 {
			lib.Xsqlite3_result_blob(ctx.tls, ctx.ptr, emptyCString, 0, sqliteStatic)
		} else {
			cv, err := malloc(ctx.tls, types.Size_t(len(v.s)))
			if err != nil {
				ctx.resultError(fmt.Errorf("alloc function result: %w", err))
				return
			}
			copy(libc.GoBytes(cv, len(v.s)), v.s)
			lib.Xsqlite3_result_blob(ctx.tls, ctx.ptr, cv, int32(len(v.s)), freeFuncPtr)
		}
	default:
		panic("unknown result Value type")
	}
}

func (ctx Context) resultError(err error) {
	errstr := err.Error()
	cerrstr, err := libc.CString(errstr)
	if err != nil {
		panic(err)
	}
	defer libc.Xfree(ctx.tls, cerrstr)
	lib.Xsqlite3_result_error(ctx.tls, ctx.ptr, cerrstr, int32(len(errstr)))
	lib.Xsqlite3_result_error_code(ctx.tls, ctx.ptr, int32(ErrCode(err)))
}

// Value represents a value that can be stored in a database table. The zero
// value is NULL. The accessor methods on Value may perform automatic
// conversions and thus methods on Value must not be called concurrently.
type Value struct {
	tls       *libc.TLS
	ptrOrType uintptr // pointer to sqlite_value if tls != nil, ColumnType otherwise

	s string
	n int64
}

// IntegerValue returns a new Value representing the given integer.
func IntegerValue(i int64) Value {
	return Value{ptrOrType: uintptr(TypeInteger), n: i}
}

// FloatValue returns a new Value representing the given floating-point number.
func FloatValue(f float64) Value {
	return Value{ptrOrType: uintptr(TypeFloat), n: int64(math.Float64bits(f))}
}

// TextValue returns a new Value representing the given string.
func TextValue(s string) Value {
	return Value{ptrOrType: uintptr(TypeText), s: s}
}

// BlobValue returns a new blob Value, copying the bytes from the given
// byte slice.
func BlobValue(b []byte) Value {
	return Value{ptrOrType: uintptr(TypeBlob), s: string(b)}
}

// Type returns the data type of the value. The result of Type is undefined if
// an automatic type conversion has occurred due to calling one of the other
// accessor methods.
func (v Value) Type() ColumnType {
	if v.ptrOrType == 0 {
		return TypeNull
	}
	if v.tls == nil {
		return ColumnType(v.ptrOrType)
	}
	return ColumnType(lib.Xsqlite3_value_type(v.tls, v.ptrOrType))
}

// Conversions follow the table in https://sqlite.org/c3ref/column_blob.html

// Int returns the value as an integer.
func (v Value) Int() int {
	return int(v.Int64())
}

// Int64 returns the value as a 64-bit integer.
func (v Value) Int64() int64 {
	if v.ptrOrType == 0 {
		return 0
	}
	if v.tls == nil {
		switch ColumnType(v.ptrOrType) {
		case TypeNull:
			return 0
		case TypeInteger:
			return v.n
		case TypeFloat:
			return int64(v.float())
		case TypeBlob, TypeText:
			return castTextToInteger(v.s)
		default:
			panic("unknown value type")
		}
	}
	return int64(lib.Xsqlite3_value_int64(v.tls, v.ptrOrType))
}

// castTextToInteger emulates the SQLite CAST operator for a TEXT value to
// INTEGER, as documented in https://sqlite.org/lang_expr.html#castexpr
func castTextToInteger(s string) int64 {
	const digits = "0123456789"
	s = strings.TrimSpace(s)
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		s = s[:1+len(longestPrefix(s[1:], digits))]
	} else {
		s = longestPrefix(s, digits)
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func longestPrefix(s string, allowSet string) string {
sloop:
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(allowSet); j++ {
			if s[i] == allowSet[j] {
				continue sloop
			}
		}
		return s[:i]
	}
	return s
}

// Float returns the value as floating-point number
func (v Value) Float() float64 {
	if v.ptrOrType == 0 {
		return 0
	}
	if v.tls == nil {
		switch ColumnType(v.ptrOrType) {
		case TypeNull:
			return 0
		case TypeInteger:
			return float64(v.n)
		case TypeFloat:
			return v.float()
		case TypeBlob, TypeText:
			return castTextToReal(v.s)
		default:
			panic("unknown value type")
		}
	}
	return float64(lib.Xsqlite3_value_double(v.tls, v.ptrOrType))
}

func (v Value) float() float64 { return math.Float64frombits(uint64(v.n)) }

// castTextToReal emulates the SQLite CAST operator for a TEXT value to
// REAL, as documented in https://sqlite.org/lang_expr.html#castexpr
func castTextToReal(s string) float64 {
	s = strings.TrimSpace(s)
	for ; len(s) > 0; s = s[:len(s)-1] {
		n, err := strconv.ParseFloat(s, 64)
		if !errors.Is(err, strconv.ErrSyntax) {
			return n
		}
	}
	return 0
}

// Text returns the value as a string.
func (v Value) Text() string {
	if v.ptrOrType == 0 {
		return ""
	}
	if v.tls == nil {
		switch ColumnType(v.ptrOrType) {
		case TypeNull:
			return ""
		case TypeInteger:
			return strconv.FormatInt(v.n, 10)
		case TypeFloat:
			return strconv.FormatFloat(v.float(), 'g', -1, 64)
		case TypeText, TypeBlob:
			return v.s
		default:
			panic("unknown value type")
		}
	}
	ptr := lib.Xsqlite3_value_text(v.tls, v.ptrOrType)
	return goStringN(ptr, int(lib.Xsqlite3_value_bytes(v.tls, v.ptrOrType)))
}

// Blob returns a copy of the value as a blob.
func (v Value) Blob() []byte {
	if v.ptrOrType == 0 {
		return nil
	}
	if v.tls == nil {
		switch ColumnType(v.ptrOrType) {
		case TypeNull:
			return nil
		case TypeInteger:
			return strconv.AppendInt(nil, v.n, 10)
		case TypeFloat:
			return strconv.AppendFloat(nil, v.float(), 'g', -1, 64)
		case TypeBlob, TypeText:
			return []byte(v.s)
		default:
			panic("unknown value type")
		}
	}
	ptr := lib.Xsqlite3_value_blob(v.tls, v.ptrOrType)
	return libc.GoBytes(ptr, int(lib.Xsqlite3_value_bytes(v.tls, v.ptrOrType)))
}

type xfunc struct {
	xFunc  func(Context, []Value) (Value, error)
	xStep  func(Context, []Value)
	xFinal func(Context) (Value, error)
}

var xfuncs = struct {
	mu  sync.RWMutex
	m   map[uintptr]*xfunc
	ids idGen
}{
	m: make(map[uintptr]*xfunc),
}

// FunctionImpl describes an application-defined SQL function. Either Scalar or
// both AggregateStep and AggregateFinal must be set.
type FunctionImpl struct {
	// NArgs is the required number of arguments that the function accepts.
	// If NArgs is negative, then the function is variadic.
	//
	// Multiple function implementations may be registered with the same name
	// with different numbers of required arguments.
	NArgs int

	// Scalar is called when a scalar function is invoked in SQL.
	Scalar func(ctx Context, args []Value) (Value, error)

	// AggregateStep is called for each row of an aggregate function's
	// SQL invocation.
	AggregateStep func(ctx Context, rowArgs []Value)
	// AggregateFinal is called after all of the aggregate function's input rows
	// have been stepped through to construct the result.
	//
	// Use closure variables to pass information between AggregateStep and
	// AggregateFinal. The AggregateFinal function should also reset any shared
	// variables to their initial states before returning.
	AggregateFinal func(ctx Context) (Value, error)

	// If Deterministic is true, the function must always give the same output
	// when the input parameters are the same. This enables functions to be used
	// in additional contexts like the WHERE clause of partial indexes and enables
	// additional optimizations.
	//
	// See https://sqlite.org/c3ref/c_deterministic.html#sqlitedeterministic for
	// more details.
	Deterministic bool

	// If AllowIndirect is false, then the function may only be invoked from
	// top-level SQL. If AllowIndirect is true, then the function can be used in
	// VIEWs, TRIGGERs, and schema structures (e.g. CHECK constraints and DEFAULT
	// clauses).
	//
	// This is the inverse of SQLITE_DIRECTONLY. See
	// https://sqlite.org/c3ref/c_deterministic.html#sqlitedirectonly for more
	// details. This defaults to false for better security by default.
	AllowIndirect bool
}

// CreateFunction registers a Go function with SQLite
// for use in SQL queries.
//
// https://sqlite.org/appfunc.html
func (c *Conn) CreateFunction(name string, impl *FunctionImpl) error {
	if c == nil {
		return fmt.Errorf("sqlite: create function: nil connection")
	}
	if name == "" {
		return fmt.Errorf("sqlite: create function: no name provided")
	}
	if impl.NArgs > 127 {
		return fmt.Errorf("sqlite: create function %s: too many permitted arguments (%d)", name, impl.NArgs)
	}
	if impl.AggregateStep == nil {
		if impl.Scalar == nil {
			return fmt.Errorf("sqlite: create function %s: must specify one of Scalar or AggregateStep", name)
		}
		if impl.AggregateFinal != nil {
			return fmt.Errorf("sqlite: create function %s: must not specify AggregateFinal with Scalar", name)
		}
	} else {
		if impl.Scalar != nil {
			return fmt.Errorf("sqlite: create function %s: both Scalar and AggregateStep specified", name)
		}
		if impl.AggregateFinal == nil {
			return fmt.Errorf("sqlite: create function %s: AggregateStep specified without AggregateFinal", name)
		}
	}

	cname, err := libc.CString(name)
	if err != nil {
		return fmt.Errorf("sqlite: create function %s: %w", name, err)
	}
	defer libc.Xfree(c.tls, cname)

	eTextRep := int32(lib.SQLITE_UTF8)
	if impl.Deterministic {
		eTextRep |= lib.SQLITE_DETERMINISTIC
	}
	if !impl.AllowIndirect {
		eTextRep |= lib.SQLITE_DIRECTONLY
	}

	x := &xfunc{
		xFunc:  impl.Scalar,
		xStep:  impl.AggregateStep,
		xFinal: impl.AggregateFinal,
	}

	xfuncs.mu.Lock()
	id := xfuncs.ids.next()
	xfuncs.m[id] = x
	xfuncs.mu.Unlock()

	// The following are conversions from function values to uintptr. It assumes
	// the memory representation described in https://golang.org/s/go11func.
	//
	// It does this by doing the following in order:
	// 1) Create a Go struct containing a pointer to a pointer to
	//    the function. It is assumed that the pointer to the function will be
	//    stored in the read-only data section and thus will not move.
	// 2) Convert the pointer to the Go struct to a pointer to uintptr through
	//    unsafe.Pointer. This is permitted via Rule #1 of unsafe.Pointer.
	// 3) Dereference the pointer to uintptr to obtain the function value as a
	//    uintptr. This is safe as long as function values are passed as pointers.
	var funcfn, stepfn, finalfn uintptr
	if impl.Scalar == nil {
		stepfn = *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, uintptr, int32, uintptr)
		}{stepTrampoline}))
		finalfn = *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, uintptr)
		}{finalTrampoline}))
	} else {
		funcfn = *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, uintptr, int32, uintptr)
		}{funcTrampoline}))
	}
	destroyfn := *(*uintptr)(unsafe.Pointer(&struct {
		f func(*libc.TLS, uintptr)
	}{destroyFuncTrampoline}))

	numArgs := impl.NArgs
	if numArgs < 0 {
		numArgs = -1
	}
	res := ResultCode(lib.Xsqlite3_create_function_v2(
		c.tls,
		c.conn,
		cname,
		int32(numArgs),
		eTextRep,
		id,
		funcfn,
		stepfn,
		finalfn,
		destroyfn,
	))
	if err := reserr(res); err != nil {
		return fmt.Errorf("sqlite: create function %s: %w", name, err)
	}
	return nil
}

func getxfuncs(tls *libc.TLS, ctx uintptr) *xfunc {
	id := lib.Xsqlite3_user_data(tls, ctx)
	xfuncs.mu.RLock()
	x := xfuncs.m[id]
	xfuncs.mu.RUnlock()
	return x
}

func funcTrampoline(tls *libc.TLS, ctx uintptr, n int32, valarray uintptr) {
	vals := make([]Value, 0, int(n))
	for ; len(vals) < cap(vals); valarray += uintptr(ptrSize) {
		vals = append(vals, Value{
			tls:       tls,
			ptrOrType: *(*uintptr)(unsafe.Pointer(valarray)),
		})
	}
	goCtx := Context{tls: tls, ptr: ctx}
	goCtx.result(getxfuncs(tls, ctx).xFunc(goCtx, vals))
}

func stepTrampoline(tls *libc.TLS, ctx uintptr, n int32, valarray uintptr) {
	vals := make([]Value, 0, int(n))
	for ; len(vals) < cap(vals); valarray += uintptr(ptrSize) {
		vals = append(vals, Value{
			tls:       tls,
			ptrOrType: *(*uintptr)(unsafe.Pointer(valarray)),
		})
	}
	goCtx := Context{tls: tls, ptr: ctx}
	getxfuncs(tls, ctx).xStep(goCtx, vals)
}

func finalTrampoline(tls *libc.TLS, ctx uintptr) {
	x := getxfuncs(tls, ctx)
	goCtx := Context{tls: tls, ptr: ctx}
	goCtx.result(x.xFinal(goCtx))
}

func destroyFuncTrampoline(tls *libc.TLS, id uintptr) {
	xfuncs.mu.Lock()
	defer xfuncs.mu.Unlock()
	delete(xfuncs.m, id)
	xfuncs.ids.reclaim(id)
}

// idGen is an ID generator. The zero value is ready to use.
type idGen struct {
	bitset []uint64
}

func (gen *idGen) next() uintptr {
	base := uintptr(1)
	for i := 0; i < len(gen.bitset); i, base = i+1, base+64 {
		b := gen.bitset[i]
		if b != 1<<64-1 {
			n := uintptr(bits.TrailingZeros64(^b))
			gen.bitset[i] |= 1 << n
			return base + n
		}
	}
	gen.bitset = append(gen.bitset, 1)
	return base
}

func (gen *idGen) reclaim(id uintptr) {
	bit := id - 1
	gen.bitset[bit/64] &^= 1 << (bit % 64)
}
