package rrule

import (
	"fmt"
	"time"
)

// QualifiedWeekday can represent a day of the week, or a certain instance
// of that day of the week.
type QualifiedWeekday struct {
	// N, when specified says which instance of the weekday relative to
	// some greater duration. -3 would be "third from the last"
	N  int
	WD time.Weekday
}

func (wd QualifiedWeekday) String() string {
	wdStr := WeekdayString(wd.WD)

	if wd.N == 0 {
		return wdStr
	}

	return fmt.Sprintf("%d%s", wd.N, wdStr)
}

// WeekdayString returns a weekday formatted as the two-letter string used in RFC5545.
func WeekdayString(wd time.Weekday) string {
	var wdStr string
	switch wd {
	case time.Sunday:
		wdStr = "SU"
	case time.Monday:
		wdStr = "MO"
	case time.Tuesday:
		wdStr = "TU"
	case time.Wednesday:
		wdStr = "WE"
	case time.Thursday:
		wdStr = "TH"
	case time.Friday:
		wdStr = "FR"
	case time.Saturday:
		wdStr = "SA"
	}
	return wdStr
}

func weekdaysInYear(t time.Time, wd QualifiedWeekday, ib invalidBehavior) []time.Time {
	allWDs := make([]time.Time, 0, 5)

	// start on first of year
	day := time.Date(t.Year(), 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	// scan til the first relevant weekday of the year
	for day.Weekday() != wd.WD {
		day = day.AddDate(0, 0, 1)
	}

	// scan over every week of the year
	for {
		allWDs = append(allWDs, day)
		day = day.AddDate(0, 0, 7)
		if day.Year() != t.Year() {
			break
		}
	}

	if wd.N == 0 {
		// no index specified, return all.
		return allWDs
	}

	if wd.N > 0 {
		// positive index specified. count to the correct instance
		if wd.N > len(allWDs) {
			switch ib {
			case OmitInvalid:
				return nil
			case PrevInvalid:
				idx := len(allWDs) - 1
				return allWDs[idx:idx]
			case NextInvalid:
				return []time.Time{allWDs[len(allWDs)-1].AddDate(0, 0, 7)}
			}
		}
		return []time.Time{allWDs[wd.N-1]}
	}

	// negative index specified. count backwards to the correct instance

	// an example of the following logic:
	//
	// -1 in a list of 4 ..
	// 	- the index becomes 3, which is the last index
	//	  which corresponds to "the last instance"
	// -3 in a list of 4 ..
	//	- the index becomes 1, which is the third from last
	// -7 in a list of 4 ..
	//	- the index becomes -3, which should trigger invalid behavior
	idx := len(allWDs) + wd.N

	if idx < 0 || idx > len(allWDs) {
		switch ib {
		case OmitInvalid:
			return nil
		case PrevInvalid:
			return []time.Time{allWDs[0].AddDate(0, 0, -7)}
		case NextInvalid:
			return allWDs[0:0]
		}
	}

	return []time.Time{allWDs[idx]}
}

func backToWeekday(t time.Time, day time.Weekday) time.Time {
	for t.Weekday() != day {
		t = t.AddDate(0, 0, -1)
	}
	return t
}

func forwardToWeekday(t time.Time, day time.Weekday) time.Time {
	for t.Weekday() != day {
		t = t.AddDate(0, 0, 1)
	}
	return t
}
