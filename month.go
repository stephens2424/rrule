package rrule

import (
	"time"
)

// monthDiffAbs returns the number of months between a and b. Passing november and
// february, respectively, counts december, januray, and feburary to return 3.
func monthDiffAbs(a, b time.Month) int {
	if a < b {
		return int(b - a)
	}

	return int((12 - a) + b)
}

// monthDiff returns the number of months that elapse to go from a to b. The
// result will be negative if b occurs before a. The month of a is not counted, but b is.
func monthDiff(a, b time.Time) int {
	if a.Equal(b) {
		return 0
	}

	// force order
	if b.Before(a) {
		return -1 * monthDiff(b, a)
	}

	years := b.Year() - a.Year()
	months := 0

	switch {
	case a.Month() < b.Month():
		months = int(b.Month() - a.Month())
	case a.Month() > b.Month():
		years--
		months = monthDiffAbs(a.Month(), b.Month())
	}

	absolute := years*12 + months
	return absolute
}
