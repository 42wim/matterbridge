package ddp

import (
	"encoding/json"
	"io"
	"strconv"
	"time"
)

// utcOffset in milliseconds for the current local time (east of UTC).
var utcOffset int64

func init() {
	_, offsetSeconds := time.Now().Zone()
	utcOffset = int64(offsetSeconds * 1000)
}

// Time is an alias for time.Time with custom json marshalling implementations to support ejson.
type Time struct {
	time.Time
}

// UnixMilli creates a new Time from the given unix millis but in UTC (as opposed to time.UnixMilli which returns
// time in the local time zone). This supports the proper loading of times from EJSON $date objects.
func UnixMilli(i int64) Time {
	return Time{Time: time.UnixMilli(i - utcOffset)}
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var data map[string]float64
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	val, ok := data["$date"]
	if !ok {
		return io.ErrUnexpectedEOF
	}
	// The time MUST be UTC but time.UnixMilli uses local time.
	// We see what time it is in local time and calculate the offset to UTC
	*t = UnixMilli(int64(val))

	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte("{\"$date\":" + strconv.FormatInt(t.UnixMilli(), 10) + "}"), nil
}

