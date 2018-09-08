package rrule

import (
	"strconv"
	"strings"
	"time"
)

func (rrule RRule) String() string {
	str := &strings.Builder{}
	str.WriteString("FREQ=")
	str.WriteString(rrule.Frequency.String())

	if !rrule.Until.IsZero() {
		str.WriteString(";UNTIL=")
		str.WriteString(rrule.Until.Format(rfc5545))
	}

	if !rrule.Dtstart.IsZero() {
		str.WriteString(";DTSTART=")
		str.WriteString(rrule.Dtstart.Format(rfc5545))
	}

	if rrule.Count != 0 {
		str.WriteString(";COUNT=")
		str.WriteString(strconv.FormatUint(rrule.Count, 10))
	}

	if rrule.Interval != 0 && rrule.Interval != 1 {
		str.WriteString(";INTERVAL=")
		str.WriteString(strconv.Itoa(rrule.Interval))
	}

	if len(rrule.BySeconds) > 0 {
		str.WriteString(";BYSECOND=")
		str.WriteString(intlist(rrule.BySeconds))
	}

	if len(rrule.ByMinutes) > 0 {
		str.WriteString(";BYMINUTE=")
		str.WriteString(intlist(rrule.ByMinutes))
	}

	if len(rrule.ByHours) > 0 {
		str.WriteString(";BYHOUR=")
		str.WriteString(intlist(rrule.ByHours))
	}

	if len(rrule.ByWeekdays) > 0 {
		str.WriteString(";BYDAY=")
		str.WriteString(weekdaylist(rrule.ByWeekdays))
	}

	if len(rrule.ByMonthDays) > 0 {
		str.WriteString(";BYMONTHDAY=")
		str.WriteString(intlist(rrule.ByMonthDays))
	}

	if len(rrule.ByYearDays) > 0 {
		str.WriteString(";BYYEARDAY=")
		str.WriteString(intlist(rrule.ByYearDays))
	}

	if len(rrule.ByMonths) > 0 {
		str.WriteString(";BYMONTH=")
		str.WriteString(monthlist(rrule.ByMonths))
	}

	if len(rrule.BySetPos) > 0 {
		str.WriteString(";BYSETPOS=")
		str.WriteString(intlist(rrule.BySetPos))
	}

	if rrule.WeekStart != nil {
		str.WriteString(";WKST=")
		str.WriteString(weekdayString(*rrule.WeekStart))
	}

	return str.String()
}

func intlist(ints []int) string {
	b := &strings.Builder{}
	for i, n := range ints {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(strconv.Itoa(n))
	}
	return b.String()
}

func weekdaylist(wds []QualifiedWeekday) string {
	b := &strings.Builder{}
	for i, wd := range wds {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(qualifiedWeekdayString(wd))
	}
	return b.String()
}

func monthlist(months []time.Month) string {
	b := &strings.Builder{}
	for i, n := range months {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(strconv.Itoa(int(n)))
	}
	return b.String()
}

func qualifiedWeekdayString(wd QualifiedWeekday) string {
	str := ""
	if wd.N != 0 {
		str += strconv.Itoa(wd.N)
	}
	str += weekdayString(wd.WD)
	return str
}

func weekdayString(wd time.Weekday) string {
	switch wd {
	case time.Monday:
		return "MO"
	case time.Tuesday:
		return "TU"
	case time.Wednesday:
		return "WE"
	case time.Thursday:
		return "TH"
	case time.Friday:
		return "FR"
	case time.Saturday:
		return "SA"
	case time.Sunday:
		return "SU"
	}

	return ""
}
