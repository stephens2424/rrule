package rrule

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekdaysInMonth(t *testing.T) {
	cases := []struct {
		Name     string
		Time     time.Time
		Weekdays []QualifiedWeekday
		IB       InvalidBehavior
		Expect   []time.Time
	}{
		{
			Name:     "simple august",
			Time:     time.Date(2018, 8, 12, 0, 0, 0, 0, time.UTC),
			Weekdays: []QualifiedWeekday{{N: 0, WD: time.Wednesday}},
			Expect: []time.Time{
				time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2018, 8, 8, 0, 0, 0, 0, time.UTC),
				time.Date(2018, 8, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2018, 8, 22, 0, 0, 0, 0, time.UTC),
				time.Date(2018, 8, 29, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			Name: "positive august",
			Time: time.Date(2018, 8, 12, 0, 0, 0, 0, time.UTC),
			Weekdays: []QualifiedWeekday{
				{N: 1, WD: time.Tuesday},
				{N: 5, WD: time.Tuesday},
			},
			Expect: []time.Time{
				time.Date(2018, 8, 7, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			Name: "negative august",
			Time: time.Date(2018, 8, 12, 0, 0, 0, 0, time.UTC),
			Weekdays: []QualifiedWeekday{
				{N: -1, WD: time.Thursday},
				{N: -4, WD: time.Tuesday},
			},
			Expect: []time.Time{
				time.Date(2018, 8, 7, 0, 0, 0, 0, time.UTC),
				time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			out := weekdaysInMonth(tt.Time, tt.Weekdays, nil, tt.IB)
			assert.Equal(t, tt.Expect, out)
		})
	}
}

func TestDaysTil(t *testing.T) {
	assert.Equal(t, 0, daysTil(time.Tuesday, time.Tuesday))
	assert.Equal(t, 1, daysTil(time.Tuesday, time.Wednesday))
	assert.Equal(t, 6, daysTil(time.Sunday, time.Saturday))
	assert.Equal(t, 2, daysTil(time.Saturday, time.Monday))
}

func TestDaysFrom(t *testing.T) {
	assert.Equal(t, 0, daysFrom(time.Tuesday, time.Tuesday))
	assert.Equal(t, 6, daysFrom(time.Tuesday, time.Wednesday))
	assert.Equal(t, 1, daysFrom(time.Sunday, time.Saturday))
	assert.Equal(t, 5, daysFrom(time.Saturday, time.Monday))
}

func TestCountWeekdaysInMonth(t *testing.T) {
	cases := []struct {
		Month   time.Month
		Year    int
		Weekday time.Weekday
		Expect  int
	}{{
		time.August,
		2018,
		time.Wednesday,
		5,
	}, {
		time.August,
		2018,
		time.Tuesday,
		4,
	}, {
		time.August,
		2018,
		time.Friday,
		5,
	}, {
		time.August,
		2018,
		time.Saturday,
		4,
	}}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("%v %v %v", tt.Year, tt.Month, tt.Weekday), func(t *testing.T) {
			ld := time.Date(tt.Year, tt.Month+1, 0, 0, 0, 0, 0, time.UTC)
			got := countWeekdaysInMonth(tt.Weekday, ld)
			assert.Equal(t, tt.Expect, got)
		})
	}
}
