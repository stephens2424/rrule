package rrule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		Input      string
		DefaultLoc *time.Location
		Expected   time.Time
	}{
		{
			Input:    "20181027T183615",
			Expected: time.Date(2018, time.October, 27, 18, 36, 15, 00, time.UTC),
		},
		{
			Input:      "DTSTART=20181027T183615",
			DefaultLoc: NewYork(),
			Expected:   time.Date(2018, time.October, 27, 18, 36, 15, 00, NewYork()),
		},
		{
			Input:    "DTSTART;TZID=America/New_York:20181027T183615",
			Expected: time.Date(2018, time.October, 27, 18, 36, 15, 00, NewYork()),
		},
		{
			Input:    "UNTIL;TZID=America/New_York:20181027T183615Z",
			Expected: time.Date(2018, time.October, 27, 18, 36, 15, 00, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			got, err := parseTime(tc.Input, tc.DefaultLoc)
			require.NoError(t, err)
			assert.Equal(t, tc.Expected, got)
		})
	}
}
