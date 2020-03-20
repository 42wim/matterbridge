package stdlib

import (
	"time"

	"github.com/d5/tengo/v2"
)

var timesModule = map[string]tengo.Object{
	"format_ansic":        &tengo.String{Value: time.ANSIC},
	"format_unix_date":    &tengo.String{Value: time.UnixDate},
	"format_ruby_date":    &tengo.String{Value: time.RubyDate},
	"format_rfc822":       &tengo.String{Value: time.RFC822},
	"format_rfc822z":      &tengo.String{Value: time.RFC822Z},
	"format_rfc850":       &tengo.String{Value: time.RFC850},
	"format_rfc1123":      &tengo.String{Value: time.RFC1123},
	"format_rfc1123z":     &tengo.String{Value: time.RFC1123Z},
	"format_rfc3339":      &tengo.String{Value: time.RFC3339},
	"format_rfc3339_nano": &tengo.String{Value: time.RFC3339Nano},
	"format_kitchen":      &tengo.String{Value: time.Kitchen},
	"format_stamp":        &tengo.String{Value: time.Stamp},
	"format_stamp_milli":  &tengo.String{Value: time.StampMilli},
	"format_stamp_micro":  &tengo.String{Value: time.StampMicro},
	"format_stamp_nano":   &tengo.String{Value: time.StampNano},
	"nanosecond":          &tengo.Int{Value: int64(time.Nanosecond)},
	"microsecond":         &tengo.Int{Value: int64(time.Microsecond)},
	"millisecond":         &tengo.Int{Value: int64(time.Millisecond)},
	"second":              &tengo.Int{Value: int64(time.Second)},
	"minute":              &tengo.Int{Value: int64(time.Minute)},
	"hour":                &tengo.Int{Value: int64(time.Hour)},
	"january":             &tengo.Int{Value: int64(time.January)},
	"february":            &tengo.Int{Value: int64(time.February)},
	"march":               &tengo.Int{Value: int64(time.March)},
	"april":               &tengo.Int{Value: int64(time.April)},
	"may":                 &tengo.Int{Value: int64(time.May)},
	"june":                &tengo.Int{Value: int64(time.June)},
	"july":                &tengo.Int{Value: int64(time.July)},
	"august":              &tengo.Int{Value: int64(time.August)},
	"september":           &tengo.Int{Value: int64(time.September)},
	"october":             &tengo.Int{Value: int64(time.October)},
	"november":            &tengo.Int{Value: int64(time.November)},
	"december":            &tengo.Int{Value: int64(time.December)},
	"sleep": &tengo.UserFunction{
		Name:  "sleep",
		Value: timesSleep,
	}, // sleep(int)
	"parse_duration": &tengo.UserFunction{
		Name:  "parse_duration",
		Value: timesParseDuration,
	}, // parse_duration(str) => int
	"since": &tengo.UserFunction{
		Name:  "since",
		Value: timesSince,
	}, // since(time) => int
	"until": &tengo.UserFunction{
		Name:  "until",
		Value: timesUntil,
	}, // until(time) => int
	"duration_hours": &tengo.UserFunction{
		Name:  "duration_hours",
		Value: timesDurationHours,
	}, // duration_hours(int) => float
	"duration_minutes": &tengo.UserFunction{
		Name:  "duration_minutes",
		Value: timesDurationMinutes,
	}, // duration_minutes(int) => float
	"duration_nanoseconds": &tengo.UserFunction{
		Name:  "duration_nanoseconds",
		Value: timesDurationNanoseconds,
	}, // duration_nanoseconds(int) => int
	"duration_seconds": &tengo.UserFunction{
		Name:  "duration_seconds",
		Value: timesDurationSeconds,
	}, // duration_seconds(int) => float
	"duration_string": &tengo.UserFunction{
		Name:  "duration_string",
		Value: timesDurationString,
	}, // duration_string(int) => string
	"month_string": &tengo.UserFunction{
		Name:  "month_string",
		Value: timesMonthString,
	}, // month_string(int) => string
	"date": &tengo.UserFunction{
		Name:  "date",
		Value: timesDate,
	}, // date(year, month, day, hour, min, sec, nsec) => time
	"now": &tengo.UserFunction{
		Name:  "now",
		Value: timesNow,
	}, // now() => time
	"parse": &tengo.UserFunction{
		Name:  "parse",
		Value: timesParse,
	}, // parse(format, str) => time
	"unix": &tengo.UserFunction{
		Name:  "unix",
		Value: timesUnix,
	}, // unix(sec, nsec) => time
	"add": &tengo.UserFunction{
		Name:  "add",
		Value: timesAdd,
	}, // add(time, int) => time
	"add_date": &tengo.UserFunction{
		Name:  "add_date",
		Value: timesAddDate,
	}, // add_date(time, years, months, days) => time
	"sub": &tengo.UserFunction{
		Name:  "sub",
		Value: timesSub,
	}, // sub(t time, u time) => int
	"after": &tengo.UserFunction{
		Name:  "after",
		Value: timesAfter,
	}, // after(t time, u time) => bool
	"before": &tengo.UserFunction{
		Name:  "before",
		Value: timesBefore,
	}, // before(t time, u time) => bool
	"time_year": &tengo.UserFunction{
		Name:  "time_year",
		Value: timesTimeYear,
	}, // time_year(time) => int
	"time_month": &tengo.UserFunction{
		Name:  "time_month",
		Value: timesTimeMonth,
	}, // time_month(time) => int
	"time_day": &tengo.UserFunction{
		Name:  "time_day",
		Value: timesTimeDay,
	}, // time_day(time) => int
	"time_weekday": &tengo.UserFunction{
		Name:  "time_weekday",
		Value: timesTimeWeekday,
	}, // time_weekday(time) => int
	"time_hour": &tengo.UserFunction{
		Name:  "time_hour",
		Value: timesTimeHour,
	}, // time_hour(time) => int
	"time_minute": &tengo.UserFunction{
		Name:  "time_minute",
		Value: timesTimeMinute,
	}, // time_minute(time) => int
	"time_second": &tengo.UserFunction{
		Name:  "time_second",
		Value: timesTimeSecond,
	}, // time_second(time) => int
	"time_nanosecond": &tengo.UserFunction{
		Name:  "time_nanosecond",
		Value: timesTimeNanosecond,
	}, // time_nanosecond(time) => int
	"time_unix": &tengo.UserFunction{
		Name:  "time_unix",
		Value: timesTimeUnix,
	}, // time_unix(time) => int
	"time_unix_nano": &tengo.UserFunction{
		Name:  "time_unix_nano",
		Value: timesTimeUnixNano,
	}, // time_unix_nano(time) => int
	"time_format": &tengo.UserFunction{
		Name:  "time_format",
		Value: timesTimeFormat,
	}, // time_format(time, format) => string
	"time_location": &tengo.UserFunction{
		Name:  "time_location",
		Value: timesTimeLocation,
	}, // time_location(time) => string
	"time_string": &tengo.UserFunction{
		Name:  "time_string",
		Value: timesTimeString,
	}, // time_string(time) => string
	"is_zero": &tengo.UserFunction{
		Name:  "is_zero",
		Value: timesIsZero,
	}, // is_zero(time) => bool
	"to_local": &tengo.UserFunction{
		Name:  "to_local",
		Value: timesToLocal,
	}, // to_local(time) => time
	"to_utc": &tengo.UserFunction{
		Name:  "to_utc",
		Value: timesToUTC,
	}, // to_utc(time) => time
}

func timesSleep(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	time.Sleep(time.Duration(i1))
	ret = tengo.UndefinedValue

	return
}

func timesParseDuration(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	s1, ok := tengo.ToString(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	dur, err := time.ParseDuration(s1)
	if err != nil {
		ret = wrapError(err)
		return
	}

	ret = &tengo.Int{Value: int64(dur)}

	return
}

func timesSince(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(time.Since(t1))}

	return
}

func timesUntil(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(time.Until(t1))}

	return
}

func timesDurationHours(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Float{Value: time.Duration(i1).Hours()}

	return
}

func timesDurationMinutes(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Float{Value: time.Duration(i1).Minutes()}

	return
}

func timesDurationNanoseconds(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: time.Duration(i1).Nanoseconds()}

	return
}

func timesDurationSeconds(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Float{Value: time.Duration(i1).Seconds()}

	return
}

func timesDurationString(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.String{Value: time.Duration(i1).String()}

	return
}

func timesMonthString(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.String{Value: time.Month(i1).String()}

	return
}

func timesDate(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 7 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}
	i2, ok := tengo.ToInt(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}
	i3, ok := tengo.ToInt(args[2])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "third",
			Expected: "int(compatible)",
			Found:    args[2].TypeName(),
		}
		return
	}
	i4, ok := tengo.ToInt(args[3])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "fourth",
			Expected: "int(compatible)",
			Found:    args[3].TypeName(),
		}
		return
	}
	i5, ok := tengo.ToInt(args[4])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "fifth",
			Expected: "int(compatible)",
			Found:    args[4].TypeName(),
		}
		return
	}
	i6, ok := tengo.ToInt(args[5])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "sixth",
			Expected: "int(compatible)",
			Found:    args[5].TypeName(),
		}
		return
	}
	i7, ok := tengo.ToInt(args[6])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "seventh",
			Expected: "int(compatible)",
			Found:    args[6].TypeName(),
		}
		return
	}

	ret = &tengo.Time{
		Value: time.Date(i1,
			time.Month(i2), i3, i4, i5, i6, i7, time.Now().Location()),
	}

	return
}

func timesNow(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 0 {
		err = tengo.ErrWrongNumArguments
		return
	}

	ret = &tengo.Time{Value: time.Now()}

	return
}

func timesParse(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	s1, ok := tengo.ToString(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	s2, ok := tengo.ToString(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	parsed, err := time.Parse(s1, s2)
	if err != nil {
		ret = wrapError(err)
		return
	}

	ret = &tengo.Time{Value: parsed}

	return
}

func timesUnix(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	i1, ok := tengo.ToInt64(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	i2, ok := tengo.ToInt64(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	ret = &tengo.Time{Value: time.Unix(i1, i2)}

	return
}

func timesAdd(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	i2, ok := tengo.ToInt64(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	ret = &tengo.Time{Value: t1.Add(time.Duration(i2))}

	return
}

func timesSub(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	t2, ok := tengo.ToTime(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "time(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Sub(t2))}

	return
}

func timesAddDate(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 4 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	i2, ok := tengo.ToInt(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	i3, ok := tengo.ToInt(args[2])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "third",
			Expected: "int(compatible)",
			Found:    args[2].TypeName(),
		}
		return
	}

	i4, ok := tengo.ToInt(args[3])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "fourth",
			Expected: "int(compatible)",
			Found:    args[3].TypeName(),
		}
		return
	}

	ret = &tengo.Time{Value: t1.AddDate(i2, i3, i4)}

	return
}

func timesAfter(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	t2, ok := tengo.ToTime(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "time(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	if t1.After(t2) {
		ret = tengo.TrueValue
	} else {
		ret = tengo.FalseValue
	}

	return
}

func timesBefore(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	t2, ok := tengo.ToTime(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	if t1.Before(t2) {
		ret = tengo.TrueValue
	} else {
		ret = tengo.FalseValue
	}

	return
}

func timesTimeYear(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Year())}

	return
}

func timesTimeMonth(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Month())}

	return
}

func timesTimeDay(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Day())}

	return
}

func timesTimeWeekday(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Weekday())}

	return
}

func timesTimeHour(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Hour())}

	return
}

func timesTimeMinute(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Minute())}

	return
}

func timesTimeSecond(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Second())}

	return
}

func timesTimeNanosecond(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: int64(t1.Nanosecond())}

	return
}

func timesTimeUnix(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: t1.Unix()}

	return
}

func timesTimeUnixNano(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Int{Value: t1.UnixNano()}

	return
}

func timesTimeFormat(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	s2, ok := tengo.ToString(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	s := t1.Format(s2)
	if len(s) > tengo.MaxStringLen {

		return nil, tengo.ErrStringLimit
	}

	ret = &tengo.String{Value: s}

	return
}

func timesIsZero(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	if t1.IsZero() {
		ret = tengo.TrueValue
	} else {
		ret = tengo.FalseValue
	}

	return
}

func timesToLocal(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Time{Value: t1.Local()}

	return
}

func timesToUTC(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.Time{Value: t1.UTC()}

	return
}

func timesTimeLocation(args ...tengo.Object) (
	ret tengo.Object,
	err error,
) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.String{Value: t1.Location().String()}

	return
}

func timesTimeString(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
		return
	}

	t1, ok := tengo.ToTime(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "time(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	ret = &tengo.String{Value: t1.String()}

	return
}
