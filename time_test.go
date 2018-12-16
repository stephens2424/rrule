package rrule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		Input            string
		DefaultLoc       *time.Location
		Expected         time.Time
		ExpectedFloating bool
	}{
		{
			Input:            "20181027T183615",
			Expected:         time.Date(2018, time.October, 27, 18, 36, 15, 00, time.UTC),
			ExpectedFloating: true,
		},
		{
			Input:            "DTSTART=20181027T183615",
			DefaultLoc:       NewYork(),
			Expected:         time.Date(2018, time.October, 27, 18, 36, 15, 00, NewYork()),
			ExpectedFloating: true,
		},
		{
			Input:            "DTSTART=20181027T183615Z",
			DefaultLoc:       NewYork(),
			Expected:         time.Date(2018, time.October, 27, 18, 36, 15, 00, time.UTC),
			ExpectedFloating: false,
		},
		{
			Input:            "DTSTART=20181027T183615-0500",
			DefaultLoc:       NewYork(),
			Expected:         time.Date(2018, time.October, 27, 18, 36, 15, 00, time.FixedZone("-0500", int(-5*time.Hour/time.Second))),
			ExpectedFloating: false,
		},

		{
			Input:            "DTSTART;TZID=America/New_York:20181027T183615",
			Expected:         time.Date(2018, time.October, 27, 18, 36, 15, 00, NewYork()),
			ExpectedFloating: false,
		},
		{
			Input:            "UNTIL;TZID=America/New_York:20181027T183615Z",
			Expected:         time.Date(2018, time.October, 27, 18, 36, 15, 00, time.UTC),
			ExpectedFloating: false,
		},
		{
			Input:            "DTSTART;TZID=America/New_York:20071104T013000",
			Expected:         time.Date(2007, time.November, 4, 1, 30, 0, 0, NewYork()),
			ExpectedFloating: false,
		},
		{
			Input:            "DTSTART;TZID=America/New_York:20070311T023000",
			Expected:         time.Date(2007, time.March, 11, 3, 30, 0, 0, NewYork()),
			ExpectedFloating: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			got, gotFloating, err := parseTime(tc.Input, tc.DefaultLoc)
			require.NoError(t, err)
			assert.True(t, tc.Expected.Equal(got), tc.Expected, got)
			assert.Equal(t, tc.ExpectedFloating, gotFloating)
		})
	}
}
