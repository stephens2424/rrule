package rrule

import "time"

// validFunc is a kind of function that checks if a time is valid against a rule. It returns true if the time is valid.
// A pointer is accepted in order to avoid the memory copy of the entire time structure. Nil is never considered valid.
type validFunc func(t *time.Time) bool

func alwaysValid(t *time.Time) bool {
	return t != nil
}

func validSecond(seconds []int) validFunc {
	if len(seconds) == 0 {
		return alwaysValid
	}

	m := intmap(seconds)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Second()]
	}
}

func validMinute(minutes []int) validFunc {
	if len(minutes) == 0 {
		return alwaysValid
	}

	m := intmap(minutes)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Minute()]
	}
}

func validHour(hours []int) validFunc {
	if len(hours) == 0 {
		return alwaysValid
	}

	m := intmap(hours)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Hour()]
	}
}

// validWeekday ignores the N modifier of QualifiedWeekday
func validWeekday(weekdays []QualifiedWeekday) validFunc {
	if len(weekdays) == 0 {
		return alwaysValid
	}

	m := weekdaymap(weekdays)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Weekday()]
	}
}

func validMonthDay(monthdays []int) validFunc {
	if len(monthdays) == 0 {
		return alwaysValid
	}

	m := intmap(monthdays)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Day()]
	}
}

func validWeek(weeks []int) validFunc {
	if len(weeks) == 0 {
		return alwaysValid
	}

	m := intmap(weeks)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[1+t.YearDay()/7]
	}
}

func validMonth(months []time.Month) validFunc {
	if len(months) == 0 {
		return alwaysValid
	}

	m := monthmap(months)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Month()]
	}
}

func validYearDay(yeardays []int) validFunc {
	if len(yeardays) == 0 {
		return alwaysValid
	}

	m := intmap(yeardays)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.YearDay()]
	}
}
