// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package jsontime

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrNotInteger = errors.New("value is not an integer")

func parseTime(data []byte, unixConv func(int64) time.Time, into *time.Time) error {
	var val int64
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	if val == 0 {
		*into = time.Time{}
	} else {
		*into = unixConv(val)
	}
	return nil
}

func anyIntegerToTime(src any, unixConv func(int64) time.Time, into *time.Time) error {
	switch v := src.(type) {
	case int:
		*into = unixConv(int64(v))
	case int8:
		*into = unixConv(int64(v))
	case int16:
		*into = unixConv(int64(v))
	case int32:
		*into = unixConv(int64(v))
	case int64:
		*into = unixConv(int64(v))
	default:
		return fmt.Errorf("%w: %T", ErrNotInteger, src)
	}

	return nil
}

var _ sql.Scanner = &UnixMilli{}
var _ driver.Valuer = UnixMilli{}

type UnixMilli struct {
	time.Time
}

func (um UnixMilli) MarshalJSON() ([]byte, error) {
	if um.IsZero() {
		return []byte{'0'}, nil
	}
	return json.Marshal(um.UnixMilli())
}

func (um *UnixMilli) UnmarshalJSON(data []byte) error {
	return parseTime(data, time.UnixMilli, &um.Time)
}

func (um UnixMilli) Value() (driver.Value, error) {
	return um.UnixMilli(), nil
}

func (um *UnixMilli) Scan(src any) error {
	return anyIntegerToTime(src, time.UnixMilli, &um.Time)
}

var _ sql.Scanner = &UnixMicro{}
var _ driver.Valuer = UnixMicro{}

type UnixMicro struct {
	time.Time
}

func (um UnixMicro) MarshalJSON() ([]byte, error) {
	if um.IsZero() {
		return []byte{'0'}, nil
	}
	return json.Marshal(um.UnixMicro())
}

func (um *UnixMicro) UnmarshalJSON(data []byte) error {
	return parseTime(data, time.UnixMicro, &um.Time)
}

func (um UnixMicro) Value() (driver.Value, error) {
	return um.UnixMicro(), nil
}

func (um *UnixMicro) Scan(src any) error {
	return anyIntegerToTime(src, time.UnixMicro, &um.Time)
}

var _ sql.Scanner = &UnixNano{}
var _ driver.Valuer = UnixNano{}

type UnixNano struct {
	time.Time
}

func (un UnixNano) MarshalJSON() ([]byte, error) {
	if un.IsZero() {
		return []byte{'0'}, nil
	}
	return json.Marshal(un.UnixNano())
}

func (un *UnixNano) UnmarshalJSON(data []byte) error {
	return parseTime(data, func(i int64) time.Time {
		return time.Unix(0, i)
	}, &un.Time)
}

func (un UnixNano) Value() (driver.Value, error) {
	return un.UnixNano(), nil
}

func (un *UnixNano) Scan(src any) error {
	return anyIntegerToTime(src, func(i int64) time.Time {
		return time.Unix(0, i)
	}, &un.Time)
}

type Unix struct {
	time.Time
}

func (u Unix) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte{'0'}, nil
	}
	return json.Marshal(u.Unix())
}

var _ sql.Scanner = &Unix{}
var _ driver.Valuer = Unix{}

func (u *Unix) UnmarshalJSON(data []byte) error {
	return parseTime(data, func(i int64) time.Time {
		return time.Unix(i, 0)
	}, &u.Time)
}

func (u Unix) Value() (driver.Value, error) {
	return u.Unix(), nil
}

func (u *Unix) Scan(src any) error {
	return anyIntegerToTime(src, func(i int64) time.Time {
		return time.Unix(i, 0)
	}, &u.Time)
}
