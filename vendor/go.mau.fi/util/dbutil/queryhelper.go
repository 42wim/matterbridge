// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
)

// DataStruct is an interface for structs that represent a single database row.
type DataStruct[T any] interface {
	Scan(row Scannable) (T, error)
}

// QueryHelper is a generic helper struct for SQL query execution boilerplate.
//
// After implementing the Scan and Init methods in a data struct, the query
// helper allows writing query functions in a single line.
type QueryHelper[T DataStruct[T]] struct {
	db      *Database
	newFunc func(qh *QueryHelper[T]) T
}

// MakeQueryHelperSimple is a form of MakeQueryHelper where the constructor function does not
// take parameters. It can be used to reduce boilerplate when not using an active record model.
//
// If the provided function is nil, the behavior is same as MakeQueryHelper.
func MakeQueryHelperSimple[T DataStruct[T]](db *Database, new func() T) *QueryHelper[T] {
	if new == nil {
		return MakeQueryHelper[T](db, nil)
	}
	return MakeQueryHelper(db, func(qh *QueryHelper[T]) T {
		return new()
	})
}

// MakeQueryHelperReflect is a form of MakeQueryHelper that uses reflection to
// create new instances of T. This requires that T is a pointer to a struct type.
func MakeQueryHelperReflect[T DataStruct[T]](db *Database) *QueryHelper[T] {
	return MakeQueryHelper[T](db, nil)
}

func reflectBuilder[T DataStruct[T]](ref reflect.Type) func(*QueryHelper[T]) T {
	return func(_ *QueryHelper[T]) T {
		return reflect.New(ref).Interface().(T)
	}
}

// MakeQueryHelper creates a new QueryHelper for the given type T, using the provided constructor function.
//
// If the constructor function is nil, it will be generated using reflection.
// This requires that T is a pointer to a struct type.
func MakeQueryHelper[T DataStruct[T]](db *Database, new func(qh *QueryHelper[T]) T) *QueryHelper[T] {
	if new == nil {
		var zero T
		new = reflectBuilder[T](reflect.TypeOf(zero).Elem())
	}
	return &QueryHelper[T]{db: db, newFunc: new}
}

// ValueOrErr is a helper function that returns the value if err is nil, or
// returns nil and the error if err is not nil. It can be used to avoid
// `if err != nil { return nil, err }` boilerplate in certain cases like
// DataStruct.Scan implementations.
func ValueOrErr[T any](val *T, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	return val, nil
}

// StrPtr returns a pointer to the given string, or nil if the string is empty.
func StrPtr[T ~string](val T) *string {
	if val == "" {
		return nil
	}
	strVal := string(val)
	return &strVal
}

// NumPtr returns a pointer to the given number, or nil if the number is zero.
func NumPtr[T constraints.Integer | constraints.Float](val T) *T {
	if val == 0 {
		return nil
	}
	return &val
}

// UnixPtr returns a pointer to the given time as unix seconds, or nil if the time is zero.
func UnixPtr(val time.Time) *int64 {
	return ConvertedPtr(val, time.Time.Unix)
}

// UnixMilliPtr returns a pointer to the given time as unix milliseconds, or nil if the time is zero.
func UnixMilliPtr(val time.Time) *int64 {
	return ConvertedPtr(val, time.Time.UnixMilli)
}

type Zeroable interface {
	IsZero() bool
}

// ConvertedPtr returns a pointer to the converted version of the given value, or nil if the input is zero.
//
// This is primarily meant for time.Time, but it can be used with any type that has implements `IsZero() bool`.
//
//	yourTime := time.Now()
//	unixMSPtr := dbutil.TimePtr(yourTime, time.Time.UnixMilli)
func ConvertedPtr[Input Zeroable, Output any](val Input, converter func(Input) Output) *Output {
	if val.IsZero() {
		return nil
	}
	converted := converter(val)
	return &converted
}

func (qh *QueryHelper[T]) GetDB() *Database {
	return qh.db
}

func (qh *QueryHelper[T]) New() T {
	return qh.newFunc(qh)
}

// Exec executes a query with ExecContext and returns the error.
//
// It omits the sql.Result return value, as it is rarely used. When the result
// is wanted, use `qh.GetDB().Exec(...)` instead, which is
// otherwise equivalent.
func (qh *QueryHelper[T]) Exec(ctx context.Context, query string, args ...any) error {
	_, err := qh.db.Exec(ctx, query, args...)
	return err
}

func (qh *QueryHelper[T]) scanNew(row Scannable) (T, error) {
	return qh.New().Scan(row)
}

// QueryOne executes a query with QueryRowContext, uses the associated DataStruct
// to scan it, and returns the value. If the query returns no rows, it returns nil
// and no error.
func (qh *QueryHelper[T]) QueryOne(ctx context.Context, query string, args ...any) (val T, err error) {
	val, err = qh.scanNew(qh.db.QueryRow(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return *new(T), nil
	}
	return val, err
}

// QueryMany executes a query with QueryContext, uses the associated DataStruct
// to scan each row, and returns the values. If the query returns no rows, it
// returns a non-nil zero-length slice and no error.
func (qh *QueryHelper[T]) QueryMany(ctx context.Context, query string, args ...any) ([]T, error) {
	return qh.QueryManyIter(ctx, query, args...).AsList()
}

// QueryManyIter executes a query with QueryContext and returns a RowIter
// that will use the associated DataStruct to scan each row.
func (qh *QueryHelper[T]) QueryManyIter(ctx context.Context, query string, args ...any) RowIter[T] {
	rows, err := qh.db.Query(ctx, query, args...)
	return NewRowIterWithError(rows, qh.scanNew, err)
}
