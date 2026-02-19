// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"reflect"
)

func reflectScan[T any](row Scannable) (*T, error) {
	t := new(T)
	val := reflect.ValueOf(t).Elem()
	fields := reflect.VisibleFields(val.Type())
	scanInto := make([]any, len(fields))
	for i, field := range fields {
		scanInto[i] = val.FieldByIndex(field.Index).Addr().Interface()
	}
	err := row.Scan(scanInto...)
	return t, err
}

// NewSimpleReflectRowIter creates a new RowIter that uses reflection to scan rows into the given type.
//
// This is a simplified implementation that always scans to all struct fields. It does not support any kind of struct tags.
func NewSimpleReflectRowIter[T any](rows Rows, err error) RowIter[*T] {
	return ConvertRowFn[*T](reflectScan[T]).NewRowIter(rows, err)
}
