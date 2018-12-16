package rrule

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Describe returns a rough English description of the recurrence.  This is
// probably not suitable for truly polished UIs, but may be useful in some
// circumstances.
func (rrule RRule) Describe() string {
	b := &strings.Builder{}

	b.WriteString("every ")
	if rrule.Interval > 1 {
		fmt.Fprintf(b, "%d ", rrule.Interval)
	}

	b.WriteString(freqStrs[rrule.Frequency])
	if rrule.Interval > 1 {
		b.WriteString("s")
	}

	if rrule.Count != 0 {
		plural := ""
		if rrule.Count > 1 {
			plural = "s"
		}
		fmt.Fprintf(b, ", for %d occurrence%s", rrule.Count, plural)
	}
	if rrule.WeekStart != nil {
		fmt.Fprintf(b, ", with weeks starting on %v", rrule.WeekStart)
	}
	if !rrule.Until.IsZero() {
		fmt.Fprintf(b, ", until %v", rrule.Until.Format(time.UnixDate))
	}
	byMonthDesc(b, rrule.ByMonths)
	byTimeDesc(b, rrule.ByMonthDays, "day of the month")
	byTimeDesc(b, rrule.ByYearDays, "day of the year")
	byTimeDesc(b, rrule.ByWeekNumbers, "week of the yar")
	byWeekday(b, rrule.ByWeekdays)
	byTimeDesc(b, rrule.ByHours, "hour")
	byTimeDesc(b, rrule.ByMinutes, "minute")
	byTimeDesc(b, rrule.BySeconds, "second")

	setByPosDesc(b, rrule.BySetPos)

	return b.String()
}

var freqStrs = map[Frequency]string{
	Yearly:   "year",
	Monthly:  "month",
	Weekly:   "week",
	Daily:    "day",
	Hourly:   "hour",
	Minutely: "minute",
	Secondly: "second",
}

func byWeekday(w io.Writer, weekdays []QualifiedWeekday) {
	if len(weekdays) == 0 {
		return
	}
	seen := map[QualifiedWeekday]bool{}
	strs := []string{}
	for _, w := range weekdays {
		if !seen[w] {
			if w.N == 0 {
				strs = append(strs, w.WD.String())
			} else {
				strs = append(strs, fmt.Sprintf("the %v %v", ordinalWithLastFrom(w.N), w.WD.String()))
			}
		}
		seen[w] = true
	}

	fmt.Fprintf(w, ", on %s", joinConj(strs, ", ", "and"))

}

func byMonthDesc(w io.Writer, months []time.Month) {
	if len(months) == 0 {
		return
	}
	seen := [12]bool{}
	strs := []string{}
	for _, m := range months {
		if !seen[m-1] {
			strs = append(strs, m.String())
		}
		seen[m-1] = true
	}

	fmt.Fprintf(w, ", in %s", joinConj(strs, ", ", "and"))
}

func byTimeDesc(w io.Writer, ints []int, unit string) {
	if len(ints) == 0 {
		return
	}

	fmt.Fprintf(w, ", on the %v %s", ordinalList(ints, ", ", "and"), unit)
}

func setByPosDesc(w io.Writer, ints []int) {
	pos := []int{}
	neg := []int{}

	for _, x := range ints {
		if x > 0 {
			pos = append(pos, x)
		} else if x < 0 {
			neg = append(neg, x)
		}
	}

	if len(pos) > 0 && len(neg) == 0 {
		fmt.Fprintf(w, ", including only the %v instances", ordinalList(pos, ", ", "and"))
		return
	}
	if len(neg) > 0 && len(pos) == 0 {
		fmt.Fprintf(w, ", including only the %v instances from the end", ordinalList(neg, ", ", "and"))
		return
	}

	if len(neg) > 0 && len(pos) > 0 {
		fmt.Fprintf(w, ", including only the %v instances and the %v instances from the end", ordinalList(pos, ", ", "and"), ordinalList(neg, ", ", "and"))
		return
	}
}

func joinConj(strs []string, sep, listConj string) string {
	switch len(strs) {
	case 0:
		return ""
	case 1:
		return strs[0]
	case 2:
		return fmt.Sprintf("%s %s %s", strs[0], listConj, strs[1])
	default:
		cp := make([]string, len(strs))
		copy(cp, strs)
		cp[len(strs)-1] = listConj + " " + strs[len(strs)-1]
		return strings.Join(cp, sep)
	}

}

func ordinalList(ints []int, sep, listConj string) string {
	s := make([]string, len(ints))
	for i, x := range ints {
		s[i] = ordinal(x)
	}

	return joinConj(s, sep, listConj)
}

func ordinal(i int) string {
	suffix := "th"
	switch i % 10 {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}

	return fmt.Sprintf("%d%s", i, suffix)
}

func ordinalWithLastFrom(i int) string {
	if i >= 0 {
		return ordinal(i)
	}
	if i == -1 {
		return "last"
	}

	return fmt.Sprintf("%v from last", ordinal(i))
}
