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
