// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package jsontime

import (
	"encoding/json"
	"time"
)

func UM(time time.Time) UnixMilli {
	return UnixMilli{Time: time}
}

func UMInt(ts int64) UnixMilli {
	return UM(time.UnixMilli(ts))
}

func UnixMilliNow() UnixMilli {
	return UM(time.Now())
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
	var val int64
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	if val == 0 {
		um.Time = time.Time{}
	} else {
		um.Time = time.UnixMilli(val)
	}
	return nil
}

func U(time time.Time) Unix {
	return Unix{Time: time}
}

func UInt(ts int64) Unix {
	return U(time.Unix(ts, 0))
}

func UnixNow() Unix {
	return U(time.Now())
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
	var val int64
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	if val == 0 {
		u.Time = time.Time{}
	} else {
		u.Time = time.Unix(val, 0)
	}
	return nil
}
