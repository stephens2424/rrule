package rrule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var recurrenceCases = []struct {
	Name string

	Recurrence *Recurrence
	String     string
	Dates      []string
}{{
	Name: "Simple",
	Recurrence: &Recurrence{
		Dtstart: now,
		RRules: []*RRule{
			{Frequency: Daily, Count: 5},
		},
		ExRules: []*RRule{
			{Frequency: Monthly, ByWeekdays: []QualifiedWeekday{{N: -1, WD: time.Tuesday}}},
		},
	},
	Dates:  []string{"2018-08-25T09:08:07Z", "2018-08-26T09:08:07Z", "2018-08-27T09:08:07Z", "2018-08-29T09:08:07Z"},
	String: "DTSTART:20180825T090807Z\nRRULE:FREQ=DAILY;COUNT=5\nEXRULE:FREQ=MONTHLY;BYDAY=-1TU\n",
}, {
	Name: "More",
	Recurrence: &Recurrence{
		Dtstart: now,
		RRules: []*RRule{
			{Frequency: Daily, Interval: 3, Count: 6},
		},
		ExRules: []*RRule{
			{Frequency: Daily, Interval: 2},
			{Frequency: Monthly, ByWeekdays: []QualifiedWeekday{{N: -1, WD: time.Tuesday}}},
		},
	},
	Dates:  []string{"2018-09-03T09:08:07Z", "2018-09-09T09:08:07Z"},
	String: "DTSTART:20180825T090807Z\nRRULE:FREQ=DAILY;COUNT=6;INTERVAL=3\nEXRULE:FREQ=DAILY;INTERVAL=2\nEXRULE:FREQ=MONTHLY;BYDAY=-1TU\n",
}, {
	Name: "Multiple",
	Recurrence: &Recurrence{
		Dtstart: now,
		RRules: []*RRule{
			{Frequency: Daily, Count: 4},
			{Frequency: Daily, Interval: 2, Count: 8},
		},
		RDates: []time.Time{time.Date(2018, time.September, 2, 9, 8, 7, 0, time.UTC), time.Date(2018, time.September, 2, 9, 8, 7, 0, time.UTC)},
		ExRules: []*RRule{
			{Frequency: Daily, Interval: 4},
			{Frequency: Daily, Interval: 8},
		},
		ExDates: []time.Time{time.Date(2018, time.September, 2, 9, 8, 7, 0, time.UTC)},
	},
	Dates:  []string{"2018-08-26T09:08:07Z", "2018-08-28T09:08:07Z", "2018-08-31T09:08:07Z", "2018-09-04T09:08:07Z", "2018-09-08T09:08:07Z"},
	String: "DTSTART:20180825T090807Z\nRRULE:FREQ=DAILY;COUNT=4\nRRULE:FREQ=DAILY;COUNT=8;INTERVAL=2\nEXRULE:FREQ=DAILY;INTERVAL=4\nEXRULE:FREQ=DAILY;INTERVAL=8\nRDATE:20180902T090807Z\nRDATE:20180902T090807Z\nEXDATE:20180902T090807Z\n",
}}

func TestRecurrence(t *testing.T) {
	for _, tc := range recurrenceCases {
		if tc.String == "" {
			continue
		}

		t.Run(tc.Name, func(t *testing.T) {
			src := tc.Recurrence.String()

			parsed, err := ParseRecurrence([]byte(src), tc.Recurrence.Dtstart.Location())
			require.NoError(t, err)
			require.NotNil(t, parsed)

			t.Log(src)
			assert.Equal(t, tc.String, src)

			dates := All(tc.Recurrence.Iterator(), 0)
			assert.Equal(t, tc.Dates, rfcAll(dates))
		})
	}
}
