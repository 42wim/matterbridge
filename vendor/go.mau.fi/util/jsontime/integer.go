// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package jsontime

import (
	"encoding/json"
	"time"
)

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

type Unix struct {
	time.Time
}

func (u Unix) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte{'0'}, nil
	}
	return json.Marshal(u.Unix())
}

func (u *Unix) UnmarshalJSON(data []byte) error {
	return parseTime(data, func(i int64) time.Time {
		return time.Unix(i, 0)
	}, &u.Time)
}
