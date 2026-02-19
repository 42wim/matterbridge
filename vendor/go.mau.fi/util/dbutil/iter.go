// Copyright (c) 2023 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"errors"
	"fmt"
	"runtime"

	"go.mau.fi/util/exzerolog"
)

var ErrAlreadyIterated = errors.New("this iterator has been already iterated")

// RowIter is a wrapper for [Rows] that allows conveniently iterating over rows
// with a predefined scanner function.
type RowIter[T any] interface {
	// Iter iterates over the rows and calls the given function for each row.
	//
	// If the function returns false, the iteration is stopped.
	// If the function returns an error, the iteration is stopped and the error is
	// returned.
	Iter(func(T) (bool, error)) error

	// AsList collects all rows into a slice.
	AsList() ([]T, error)
}

type ConvertRowFn[T any] func(Scannable) (T, error)

// NewRowIter is a proxy for NewRowIterWithError for more convenient usage.
//
// For example:
//
//	func exampleConvertRowFn(rows Scannable) (*YourType, error) {
//		...
//	}
//	func exampleFunction() {
//		iter := dbutil.ConvertRowFn[*YourType](exampleConvertRowFn).NewRowIter(
//			db.Query("SELECT ..."),
//		)
//	}
func (crf ConvertRowFn[T]) NewRowIter(rows Rows, err error) RowIter[T] {
	return newRowIterWithError(rows, crf, err)
}

type rowIterImpl[T any] struct {
	Rows
	ConvertRow ConvertRowFn[T]

	iterated bool
	caller   string
	err      error
}

// NewRowIter creates a new RowIter from the given Rows and scanner function.
//
// Deprecated: use NewRowIterWithError instead to avoid an unnecessary separate error check on the Query result.
//
// Instead of
//
//	func DoQuery(...) (dbutil.RowIter, error) {
//		rows, err := db.Query(...)
//		if err != nil {
//			return nil, err
//		}
//		return dbutil.NewRowIter(rows, convertFn), nil
//	}
//
// you should use
//
//	func DoQuery(...) dbutil.RowIter {
//		rows, err := db.Query(...)
//		return dbutil.NewRowIterWithError(rows, convertFn, err)
//	}
//
// or alternatively pre-wrap the convertFn
//
//	var converter = dbutil.ConvertRowFn(convertFn)
//	func DoQuery(...) dbutil.RowIter {
//		return converter.NewRowIter(db.Query(...))
//	}
//
// Embedding the error in the iterator allows the caller to do only one error check instead of two:
//
//	iter, err := DoQuery(...)
//	if err != nil { ... }
//	result, err := iter.Iter(...)
//	if err != nil { ... }
//
// vs
//
//	result, err := DoQuery(...).Iter(...)
//	if err != nil { ... }
func NewRowIter[T any](rows Rows, convertFn ConvertRowFn[T]) RowIter[T] {
	return newRowIterWithError(rows, convertFn, nil)
}

// NewRowIterWithError creates a new RowIter from the given Rows and scanner function with default error. If not nil, it will be returned without calling iterator function.
func NewRowIterWithError[T any](rows Rows, convertFn ConvertRowFn[T], err error) RowIter[T] {
	return newRowIterWithError(rows, convertFn, err)
}

func newRowIterWithError[T any](rows Rows, convertFn ConvertRowFn[T], err error) RowIter[T] {
	ri := &rowIterImpl[T]{Rows: rows, ConvertRow: convertFn, err: err}
	if err == nil {
		callerSkip := 2
		if pc, file, line, ok := runtime.Caller(callerSkip); ok {
			ri.caller = exzerolog.CallerWithFunctionName(pc, file, line)
		}
		runtime.SetFinalizer(ri, (*rowIterImpl[T]).destroy)
	}
	return ri
}

func ScanSingleColumn[T any](rows Scannable) (val T, err error) {
	err = rows.Scan(&val)
	return
}

type NewableDataStruct[T any] interface {
	DataStruct[T]
	New() T
}

func ScanDataStruct[T NewableDataStruct[T]](rows Scannable) (T, error) {
	var val T
	return val.New().Scan(rows)
}

func (i *rowIterImpl[T]) destroy() {
	if !i.iterated {
		panic(fmt.Errorf("RowIter created at %s wasn't iterated", i.caller))
	}
}

func (i *rowIterImpl[T]) Iter(fn func(T) (bool, error)) error {
	if i == nil {
		return nil
	} else if i.Rows == nil || i.err != nil {
		return i.err
	}
	defer func() {
		_ = i.Rows.Close()
		i.iterated = true
	}()

	for i.Rows.Next() {
		if item, err := i.ConvertRow(i.Rows); err != nil {
			i.err = err
			return err
		} else if cont, err := fn(item); err != nil {
			i.err = err
			return err
		} else if !cont {
			break
		}
	}

	err := i.Rows.Err()
	if err != nil {
		i.err = err
	} else {
		i.err = ErrAlreadyIterated
	}
	return err
}

func (i *rowIterImpl[T]) AsList() (list []T, err error) {
	err = i.Iter(func(item T) (bool, error) {
		list = append(list, item)
		return true, nil
	})
	return
}

func RowIterAsMap[T any, Key comparable, Value any](ri RowIter[T], getKeyValue func(T) (Key, Value)) (map[Key]Value, error) {
	m := make(map[Key]Value)
	err := ri.Iter(func(item T) (bool, error) {
		k, v := getKeyValue(item)
		m[k] = v
		return true, nil
	})
	return m, err
}

type sliceIterImpl[T any] struct {
	items []T
	err   error
}

func NewSliceIter[T any](items []T) RowIter[T] {
	return &sliceIterImpl[T]{items: items}
}

func NewSliceIterWithError[T any](items []T, err error) RowIter[T] {
	return &sliceIterImpl[T]{items: items, err: err}
}

func (i *sliceIterImpl[T]) Iter(fn func(T) (bool, error)) error {
	if i == nil {
		return nil
	} else if i.err != nil {
		return i.err
	}

	for _, item := range i.items {
		if cont, err := fn(item); err != nil {
			i.err = err
			return err
		} else if !cont {
			break
		}
	}

	i.err = ErrAlreadyIterated
	return nil
}

func (i *sliceIterImpl[T]) AsList() ([]T, error) {
	if i == nil {
		return nil, nil
	} else if i.err != nil {
		return nil, i.err
	}

	i.err = ErrAlreadyIterated
	return i.items, nil
}
