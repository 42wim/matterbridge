// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package jsontime

import (
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

func UMicro(time time.Time) UnixMicro {
	return UnixMicro{Time: time}
}

func UMicroInto(ts int64) UnixMicro {
	return UMicro(time.UnixMicro(ts))
}

func UnixMicroNow() UnixMicro {
	return UMicro(time.Now())
}

func UN(time time.Time) UnixNano {
	return UnixNano{Time: time}
}

func UNInt(ts int64) UnixNano {
	return UN(time.Unix(0, ts))
}

func UnixNanoNow() UnixNano {
	return UN(time.Now())
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
