package rrule

import (
	"log"
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
func weekdaysInMonth(t time.Time, weekdays []QualifiedWeekday, ib InvalidBehavior) []time.Time {
	firstDay := firstOfMonth(t)
	firstWeekday := firstDay.Weekday()
	lastDay := lastOfMonth(t)
	lastDate := lastDay.Day()

	dates := make([]int, 0, len(weekdays))
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
			}
		}

		if weekday.N < 0 {
			needWDBefore := lastDay.Day() + (7 * (weekday.N + 1))
			date := needWDBefore - daysFrom(lastDay.Weekday(), weekday.WD)
			log.Println(lastDay, weekday.N+1, needWDBefore, date)
			if date > 0 {
				dates = append(dates, date)
			}
		}
	}

	sort.Ints(dates)

	out := make([]time.Time, len(dates))
	for i, date := range dates {
		out[i] = firstDay.AddDate(0, 0, date-1)
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

func monthDates(t time.Time, days []int) []time.Time {
	dates := make([]time.Time, len(days))
	for i, d := range days {
		dates[i] = time.Date(t.Year(), t.Month(), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	}

	return dates
}
