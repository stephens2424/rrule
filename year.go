package rrule

import (
	"time"
)

// yearStart returns a time on the first day of the year specified by t. Time
// and location are copied from t.
func yearStart(t time.Time, wkstart time.Weekday) time.Time {
	jan1 := time.Date(t.Year(), time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	fw := forwardToWeekday(jan1, wkstart)

	// if by going forward, we're on or before the 4th, we are in the first week.
	if fw.Day() <= 4 {
		return fw
	}

	// otherwise we must go backward to the start of the first week
	return backToWeekday(jan1, wkstart)
}
