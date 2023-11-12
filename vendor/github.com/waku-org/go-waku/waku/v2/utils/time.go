package utils

import (
	"time"

	"google.golang.org/protobuf/proto"
)

// GetUnixEpochFrom converts a time into a unix timestamp with nanoseconds
func GetUnixEpochFrom(now time.Time) *int64 {
	return proto.Int64(now.UnixNano())
}

type Timesource interface {
	Now() time.Time
}

// GetUnixEpoch returns the current time in unix timestamp with the integer part
// representing seconds and the decimal part representing subseconds.
// Optionally receives a timesource to obtain the time from
func GetUnixEpoch(timesource ...Timesource) *int64 {
	if len(timesource) != 0 {
		return GetUnixEpochFrom(timesource[0].Now())
	}

	return GetUnixEpochFrom(time.Now())
}
