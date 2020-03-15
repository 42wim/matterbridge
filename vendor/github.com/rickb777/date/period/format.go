// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rickb777/plural"
)

// Format converts the period to human-readable form using the default localisation.
func (period Period) Format() string {
	return period.FormatWithPeriodNames(PeriodYearNames, PeriodMonthNames, PeriodWeekNames, PeriodDayNames, PeriodHourNames, PeriodMinuteNames, PeriodSecondNames)
}

// FormatWithPeriodNames converts the period to human-readable form in a localisable way.
func (period Period) FormatWithPeriodNames(yearNames, monthNames, weekNames, dayNames, hourNames, minNames, secNames plural.Plurals) string {
	period = period.Abs()

	parts := make([]string, 0)
	parts = appendNonBlank(parts, yearNames.FormatFloat(absFloat10(period.years)))
	parts = appendNonBlank(parts, monthNames.FormatFloat(absFloat10(period.months)))

	if period.days > 0 || (period.IsZero()) {
		if len(weekNames) > 0 {
			weeks := period.days / 70
			mdays := period.days % 70
			//fmt.Printf("%v %#v - %d %d\n", period, period, weeks, mdays)
			if weeks > 0 {
				parts = appendNonBlank(parts, weekNames.FormatInt(int(weeks)))
			}
			if mdays > 0 || weeks == 0 {
				parts = appendNonBlank(parts, dayNames.FormatFloat(absFloat10(mdays)))
			}
		} else {
			parts = appendNonBlank(parts, dayNames.FormatFloat(absFloat10(period.days)))
		}
	}
	parts = appendNonBlank(parts, hourNames.FormatFloat(absFloat10(period.hours)))
	parts = appendNonBlank(parts, minNames.FormatFloat(absFloat10(period.minutes)))
	parts = appendNonBlank(parts, secNames.FormatFloat(absFloat10(period.seconds)))

	return strings.Join(parts, ", ")
}

func appendNonBlank(parts []string, s string) []string {
	if s == "" {
		return parts
	}
	return append(parts, s)
}

// PeriodDayNames provides the English default format names for the days part of the period.
// This is a sequence of plurals where the first match is used, otherwise the last one is used.
// The last one must include a "%v" placeholder for the number.
var PeriodDayNames = plural.FromZero("%v days", "%v day", "%v days")

// PeriodWeekNames is as for PeriodDayNames but for weeks.
var PeriodWeekNames = plural.FromZero("", "%v week", "%v weeks")

// PeriodMonthNames is as for PeriodDayNames but for months.
var PeriodMonthNames = plural.FromZero("", "%v month", "%v months")

// PeriodYearNames is as for PeriodDayNames but for years.
var PeriodYearNames = plural.FromZero("", "%v year", "%v years")

// PeriodHourNames is as for PeriodDayNames but for hours.
var PeriodHourNames = plural.FromZero("", "%v hour", "%v hours")

// PeriodMinuteNames is as for PeriodDayNames but for minutes.
var PeriodMinuteNames = plural.FromZero("", "%v minute", "%v minutes")

// PeriodSecondNames is as for PeriodDayNames but for seconds.
var PeriodSecondNames = plural.FromZero("", "%v second", "%v seconds")

// String converts the period to ISO-8601 form.
func (period Period) String() string {
	if period.IsZero() {
		return "P0D"
	}

	buf := &bytes.Buffer{}
	if period.Sign() < 0 {
		buf.WriteByte('-')
	}

	buf.WriteByte('P')

	if period.years != 0 {
		fmt.Fprintf(buf, "%gY", absFloat10(period.years))
	}
	if period.months != 0 {
		fmt.Fprintf(buf, "%gM", absFloat10(period.months))
	}
	if period.days != 0 {
		if period.days%70 == 0 {
			fmt.Fprintf(buf, "%gW", absFloat10(period.days/7))
		} else {
			fmt.Fprintf(buf, "%gD", absFloat10(period.days))
		}
	}
	if period.hours != 0 || period.minutes != 0 || period.seconds != 0 {
		buf.WriteByte('T')
	}
	if period.hours != 0 {
		fmt.Fprintf(buf, "%gH", absFloat10(period.hours))
	}
	if period.minutes != 0 {
		fmt.Fprintf(buf, "%gM", absFloat10(period.minutes))
	}
	if period.seconds != 0 {
		fmt.Fprintf(buf, "%gS", absFloat10(period.seconds))
	}

	return buf.String()
}

func absFloat10(v int16) float32 {
	f := float32(v) / 10
	if v < 0 {
		return -f
	}
	return f
}
