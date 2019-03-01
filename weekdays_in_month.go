package rrule

import (
	"sort"
	"time"
)

// weekdaysInMonth finds all the applicable weekdays in the month of t.
//
// weekdaysInMonth is a more complex function than I prefer, but the time savings
// by only calculating the first of the month once, plus returning an already sorted
// list, outweighs the concerns.
//
// weekdays must have at least one element
//
// If ib is not OmitInvalid, the returned set will have instances in the
// preceeding and following months if the requested weekdays go beyond the
// bounds of the month.
func weekdaysInMonth(t time.Time, weekdays []QualifiedWeekday, bySetPos []int, ib InvalidBehavior) []time.Time {
	firstDay := firstOfMonth(t)
	firstWeekday := firstDay.Weekday()
	lastDay := lastOfMonth(t)
	lastDate := lastDay.Day()

	dates := make([]int, 0, 30)
	var addLastPrevMonth bool
	var addFirstNextMonth bool

	for _, weekday := range weekdays {
		countOfWD := countWeekdaysInMonth(weekday.WD, lastDay)

		if weekday.N == 0 {
			daysTil := daysTil(firstWeekday, weekday.WD)
			for i := 0; i < countOfWD; i++ {
				date := i*7 + daysTil + 1
				if date <= lastDate {
					dates = append(dates, date)
				}
			}
		}

		if weekday.N > 0 {
			daysTil := daysTil(firstWeekday, weekday.WD)
			date := ((weekday.N - 1) * 7) + daysTil + 1
			if date <= lastDay.Day() {
				dates = append(dates, date)
			} else {
				switch ib {
				case NextInvalid:
					addFirstNextMonth = true
				case PrevInvalid:
					dates = append(dates, lastDay.Day())
				}
			}
		}

		if weekday.N < 0 {
			needWDBefore := lastDay.Day() + (7 * (weekday.N + 1))
			date := needWDBefore - daysFrom(lastDay.Weekday(), weekday.WD)
			if date > 0 {
				dates = append(dates, date)
			} else {
				switch ib {
				case NextInvalid:
					dates = append(dates, date)
				case PrevInvalid:
					addLastPrevMonth = true
				}
			}
		}
	}

	sort.Ints(dates)
	dates = limitInstancesBySetPos(dates, bySetPos)

	excessDates := 0
	if addLastPrevMonth {
		excessDates++
	}
	if addFirstNextMonth {
		excessDates++
	}

	out := make([]time.Time, 0, len(dates)+excessDates)
	if addLastPrevMonth {
		out = append(out, firstDay.AddDate(0, 0, -1))
	}

	// it's possible we get duplicates with invalid behavior. avoid that.
	// but avoid the extra allocation when we don't need this.
	var addedMap map[int]bool
	if ib != OmitInvalid {
		addedMap = make(map[int]bool, cap(out))
	}

	for _, date := range dates {
		if addedMap != nil {
			if has := addedMap[date]; has {
				continue
			}
			addedMap[date] = true
		}

		out = append(out, firstDay.AddDate(0, 0, date-1))
	}

	if addFirstNextMonth {
		out = append(out, firstDay.AddDate(0, 1, 0))
	}

	return out
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func lastOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func daysTil(from, to time.Weekday) int {
	if from == to {
		return 0
	}
	diff := to - from
	if diff < 0 {
		diff += 7
	}
	return int(diff)
}

func daysFrom(startFrom, backTo time.Weekday) int {
	if startFrom == backTo {
		return 0
	}
	diff := startFrom - backTo
	if diff < 0 {
		diff += 7
	}
	return int(diff)
}

func countWeekdaysInMonth(wd time.Weekday, lastDayOfMonth time.Time) int {
	lastDate := lastDayOfMonth.Day()
	lastWD := lastDayOfMonth.Weekday()

	daysBack := daysFrom(lastWD, wd)
	if daysBack < (lastDate - 28) {
		return 5
	}

	return 4
}
