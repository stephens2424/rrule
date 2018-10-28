package rrule

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecurrence(t *testing.T) {
	for _, tc := range cases {
		if tc.NoTest {
			continue
		}

		if tc.String == "" {
			continue
		}

		t.Run(tc.Name, func(t *testing.T) {
			src := fmt.Sprintf("DTSTART:%s\nRRULE:%s", tc.RRule.Dtstart.Format(rfc5545_WithOffset), tc.RRule.String())
			if tc.RRule.Dtstart.Location() != time.UTC {
				src = fmt.Sprintf("DTSTART;TZID=%s:%s\nRRULE:%s", tc.RRule.Dtstart.Location(), tc.RRule.Dtstart.Format(rfc5545_WithoutOffset), tc.RRule.String())
			}

			parsed, err := ParseRecurrence([]byte(src), tc.RRule.Dtstart.Location())
			require.NoError(t, err)
			require.NotNil(t, parsed)

			t.Log(src)

			assert.Len(t, parsed.ExRules, 0)
			assert.Len(t, parsed.RDates, 0)
			assert.Len(t, parsed.ExDates, 0)
			require.Len(t, parsed.RRules, 1)

			tc.RRule.Dtstart = tc.RRule.Dtstart.Truncate(time.Second)

			assert.Equal(t, tc.RRule, *parsed.RRules[0])
		})
	}
}
