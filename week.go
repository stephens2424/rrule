package rrule

import "time"

// diffWeekdayAbs returns the number of days from a to b.
func diffWeekdayAbs(a, b time.Weekday) int {
	diff := int(b - a)
	if diff < 0 {
		diff += 7
	}
	return diff
}
