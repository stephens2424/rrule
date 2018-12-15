package rrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupIterator(t *testing.T) {
	iter := groupIteratorFromRRules(
		[]*RRule{
			MustRRule("FREQ=WEEKLY;COUNT=5;BYDAY=MO"),
			MustRRule("FREQ=WEEKLY;COUNT=5;BYDAY=TU"),
		},
	)

	var count int
	for {
		t1 := iter.Peek()
		t2 := iter.Next()
		if t1 == nil && t2 == nil {
			break
		} else if t1 == nil || t2 == nil {
			t.Fatal("Peek followed by Next should always return the same thing")
			break
		}

		count++
		assert.Equal(t, *t1, *t2)
	}

	require.Equal(t, 10, count)

}
