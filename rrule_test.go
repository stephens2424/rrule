package rrule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	rrule "github.com/teambition/rrule-go"
)

var now = time.Date(2018, 8, 25, 9, 8, 7, 6, time.UTC)

var cases = []struct {
	Name     string
	RRule    RRule
	Dates    []string
	Terminal bool

	NoBenchmark bool
	NoTest      bool
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
		Name: "simple minutely",
		RRule: RRule{
			Frequency: Minutely,
			Count:     3,
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-08-25T09:09:07Z", "2018-08-25T09:10:07Z"},
		Terminal: true,
	},

	{
		Name: "simple hourly",
		RRule: RRule{
			Frequency: Hourly,
			Count:     3,
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-08-25T10:08:07Z", "2018-08-25T11:08:07Z"},
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
		Name: "daily setpos",
		RRule: RRule{
			Frequency:  Monthly,
			ByWeekdays: []QualifiedWeekday{{N: 0, WD: time.Monday}, {N: 0, WD: time.Tuesday}, {N: 0, WD: time.Wednesday}, {N: 0, WD: time.Thursday}, {N: 0, WD: time.Friday}, {N: 0, WD: time.Saturday}, {N: 0, WD: time.Sunday}},
			Count:      4,
			Dtstart:    now,
			ByMonths:   []time.Month{time.August, time.September},
			BySetPos:   []int{1, 3, -1},
		},
		Dates:    []string{"2018-08-31T09:08:07Z", "2018-09-01T09:08:07Z", "2018-09-03T09:08:07Z", "2018-09-30T09:08:07Z"},
		Terminal: true,
	},

	{
		Name: "daily until",
		RRule: RRule{
			Frequency: Daily,
			Until:     time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC),
			Dtstart:   now,
		},
		Dates:    []string{"2018-08-25T09:08:07Z", "2018-08-26T09:08:07Z", "2018-08-27T09:08:07Z", "2018-08-28T09:08:07Z", "2018-08-29T09:08:07Z"},
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
		Name: "long monthly",
		RRule: RRule{
			Frequency: Monthly,
			Count:     300,
			Dtstart:   now,
		},
		Terminal: true,
		NoTest:   true,
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
		if tc.NoTest {
			continue
		}

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
				tc.RRule.All(0)
			}
		})
	}
}

func rruleToROption(r RRule) rrule.ROption {
	converted := rrule.ROption{
		Dtstart: r.Dtstart,

		Until:    r.Until,
		Count:    int(r.Count),
		Interval: r.Interval,

		Bysecond:   r.BySeconds,
		Byminute:   r.ByMinutes,
		Byhour:     r.ByHours,
		Bymonthday: r.ByMonthDays,
		Byweekno:   r.ByWeekNumbers,
		Byyearday:  r.ByYearDays,
		Bysetpos:   r.BySetPos,

		Bymonth:   make([]int, 0, len(r.ByMonths)),
		Byweekday: make([]rrule.Weekday, 0, len(r.ByWeekdays)),
	}

	switch r.Frequency {
	case Secondly:
		converted.Freq = rrule.SECONDLY
	case Minutely:
		converted.Freq = rrule.MINUTELY
	case Hourly:
		converted.Freq = rrule.HOURLY
	case Daily:
		converted.Freq = rrule.DAILY
	case Weekly:
		converted.Freq = rrule.WEEKLY
	case Monthly:
		converted.Freq = rrule.MONTHLY
	case Yearly:
		converted.Freq = rrule.YEARLY
	}

	for _, m := range r.ByMonths {
		converted.Bymonth = append(converted.Bymonth, int(m))
	}
	for _, wd := range r.ByWeekdays {
		switch wd.WD {
		case time.Sunday:
			converted.Byweekday = append(converted.Byweekday, rrule.SU)
		case time.Monday:
			converted.Byweekday = append(converted.Byweekday, rrule.MO)
		case time.Tuesday:
			converted.Byweekday = append(converted.Byweekday, rrule.TU)
		case time.Wednesday:
			converted.Byweekday = append(converted.Byweekday, rrule.WE)
		case time.Thursday:
			converted.Byweekday = append(converted.Byweekday, rrule.TH)
		case time.Friday:
			converted.Byweekday = append(converted.Byweekday, rrule.FR)
		case time.Saturday:
			converted.Byweekday = append(converted.Byweekday, rrule.SA)
		}
	}

	return converted
}

func BenchmarkTeambition(b *testing.B) {
	for _, tc := range cases {

		ro := rruleToROption(tc.RRule)

		b.Run(tc.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				teambitionRRule, _ := rrule.NewRRule(ro)
				teambitionRRule.All()
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
