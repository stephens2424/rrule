package rrule

import "time"

func intmap(ints []int) map[int]bool {
	m := make(map[int]bool, len(ints))
	for _, v := range ints {
		m[v] = true
	}
	return m
}

func weekdaymap(weekdays []QualifiedWeekday) map[time.Weekday]bool {
	m := make(map[time.Weekday]bool, len(weekdays))
	for _, v := range weekdays {
		m[v.WD] = true
	}
	return m
}

func monthmap(months []time.Month) map[time.Month]bool {
	m := make(map[time.Month]bool, len(months))
	for _, v := range months {
		m[v] = true
	}
	return m
}
