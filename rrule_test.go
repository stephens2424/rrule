package rrule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = time.Date(2018, 8, 25, 9, 8, 7, 6, time.UTC)

var cases = []struct {
	Name     string
	RRule    RRule
	Dates    []string
	Terminal bool
}{
	{
		Name: "simple secondly",
		RRule: RRule{
			Frequency: Secondly,
			Count:     3,
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-08-25T09:08:08Z", "2018-08-25T09:08:09Z"},
		Terminal: true,
	},
	{
		Name: "simple daily",
		RRule: RRule{
			Frequency: Daily,
			Count:     3,
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-08-26T09:08:07Z", "2018-08-27T09:08:07Z"},
		Terminal: true,
	},
	{
		Name: "simple monthly",
		RRule: RRule{
			Frequency: Monthly,
			Count:     3,
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-09-25T09:08:07Z", "2018-10-25T09:08:07Z"},
		Terminal: true,
	},
	{
		Name: "monthly by weekday",
		RRule: RRule{
			Frequency:  Monthly,
			Count:      3,
			Dtstart:    now,
			ByWeekdays: []QualifiedWeekday{{N: 1, WD: time.Tuesday}},
		},
		Dates:    []string{"2018-09-04T09:08:07Z", "2018-10-02T09:08:07Z", "2018-11-06T09:08:07Z"},
		Terminal: true,
	},

	{
		Name: "simple weekly",
		RRule: RRule{
			Frequency: Weekly,
			Count:     3,
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-09-01T09:08:07Z", "2018-09-08T09:08:07Z"},
		Terminal: true,
	},

	{
		Name: "weekly by weekday",
		RRule: RRule{
			Frequency:  Weekly,
			Count:      3,
			Dtstart:    now,
			ByWeekdays: []QualifiedWeekday{{WD: time.Tuesday}},
		},
		Dates:    []string{"2018-08-28T09:08:07Z", "2018-09-04T09:08:07Z", "2018-09-11T09:08:07Z"},
		Terminal: true,
	},
}

func TestRRule(t *testing.T) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			dates := tc.RRule.All(0)
			assert.Equal(t, tc.Dates, rfcAll(dates))
		})
	}
}

func BenchmarkRRule(b *testing.B) {
	for _, tc := range cases {
		b.Run(tc.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				dates := tc.RRule.All(0)
				assert.Equal(b, tc.Dates, rfcAll(dates))
			}
		})
	}
}

func rfcAll(times []time.Time) []string {
	strs := make([]string, len(times))
	for i, t := range times {
		strs[i] = t.Format(time.RFC3339)
	}
	return strs
}
