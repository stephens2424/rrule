package rrule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonthDiffAbs(t *testing.T) {
	cases := []struct {
		A, B time.Month
		Diff int
	}{
		{
			A:    time.January,
			B:    time.February,
			Diff: 1,
		},
		{
			A:    time.January,
			B:    time.January,
			Diff: 12,
		},
		{
			A:    time.November,
			B:    time.February,
			Diff: 3,
		},
		{
			A:    time.December,
			B:    time.January,
			Diff: 1,
		},
	}

	for _, tc := range cases {
		absDiff := monthDiffAbs(tc.A, tc.B)
		assert.Equal(t, tc.Diff, absDiff, "%v -> %v should be %v, got %v", tc.A, tc.B, tc.Diff, absDiff)
	}
}

func TestMonthDiff(t *testing.T) {
	cases := []struct {
		A, B time.Time
		Diff int
	}{
		{
			A:    time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
			B:    time.Date(2008, 12, 31, 0, 0, 0, 0, time.UTC),
			Diff: 14,
		},
		{
			A:    time.Date(2008, 12, 31, 0, 0, 0, 0, time.UTC),
			B:    time.Date(2009, 10, 1, 0, 0, 0, 0, time.UTC),
			Diff: 10,
		},
		{
			A:    time.Date(2009, 10, 1, 0, 0, 0, 0, time.UTC),
			B:    time.Date(2008, 12, 31, 0, 0, 0, 0, time.UTC),
			Diff: -10,
		},
		{
			A:    time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
			B:    time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
			Diff: 0,
		},
	}

	for _, tc := range cases {
		diff := monthDiff(tc.A, tc.B)
		assert.Equal(t, tc.Diff, diff, "%v -> %v should be %v, got %v", tc.A, tc.B, tc.Diff, diff)
	}
}
