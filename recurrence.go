package rrule

import (
	"strings"
	"time"
)

// Recurrence expresses a complex pattern of repeating events composed of individual
// patterns and extra days that are filtered by exclusion patterns and days.
type Recurrence struct {
	// Dtstart specifies the time to begin recurrence. If zero, time.Now is
	// used when an iterator is generated.  The location of Dtstart is the
	// location that will be used to process the recurrence, which is
	// particularly relevant for calculations affected by Daylight Savings.
	Dtstart time.Time

	// FloatingLocation determines how the Recurrence is encoded to string.
	// If true, Dtstart, RDates, and ExDates will be written in local time,
	// excluding the offset or timezone indicator, to represent a local time
	// independent of timezone. See ParseRecurrence or RFC 5545 for more
	// detail.
	FloatingLocation bool

	// Patterns and instances to include. Repeated instances are included only
	// once, even if defined by multiple patterns.
	//
	// The Dtstart property of RRule and ExRule patterns are
	// ignored, including when the above Dtstart property is zero.
	RRules []RRule
	RDates []time.Time

	// Patterns and instances to exclude. These take precedence over the
	// inclusions. Note: this feature was deprecated in RFC5545, noting its
	// limited (and buggy) adoption and real-world use case. It is
	// implemented here, nonetheless, for maximum flexibility and
	// compatibility.
	ExRules []RRule
	ExDates []time.Time
}

func (r *Recurrence) String() string {
	b := &strings.Builder{}
	if !r.Dtstart.IsZero() {
		b.WriteString(formatTime("DTSTART", r.Dtstart, r.FloatingLocation))
		b.WriteString("\n")
	}
	for _, rrule := range r.RRules {
		b.WriteString("RRULE:")
		b.WriteString(rrule.String())
		b.WriteString("\n")
	}
	for _, exrule := range r.ExRules {
		b.WriteString("EXRULE:")
		b.WriteString(exrule.String())
		b.WriteString("\n")
	}
	for _, rdate := range r.RDates {
		b.WriteString(formatTime("RDATE", rdate, r.FloatingLocation))
		b.WriteString("\n")
	}
	for _, exdate := range r.ExDates {
		b.WriteString(formatTime("EXDATE", exdate, r.FloatingLocation))
		b.WriteString("\n")
	}

	return b.String()
}

func (r *Recurrence) setDtstart() {
	for i, rr := range r.RRules {
		rr.Dtstart = r.Dtstart
		r.RRules[i] = rr
	}
	for i, rr := range r.ExRules {
		rr.Dtstart = r.Dtstart
		r.ExRules[i] = rr
	}
}

// All returns all instances from the beginning of the iterator up to a limited
// number. If the limit is 0, all instances are returned, which will include all
// instances until Go's maximum useful time.Time, in the year 219248499.
func All(it Iterator, limit int) []time.Time {
	all := make([]time.Time, 0)
	for {
		next := it.Next()
		if next == nil {
			break
		}
		all = append(all, *next)
		if limit > 0 && len(all) == limit {
			break
		}
	}
	return all
}

// Iterator returns an iterator for the recurrence.
func (r Recurrence) Iterator() Iterator {
	r.setDtstart()

	ri := &recurrenceIterator{
		rrules:  groupIteratorFromRRules(r.RRules),
		exrules: groupIteratorFromRRules(r.ExRules),
	}

	ri.rrules.iters = append(ri.rrules.iters, &iterator{queue: r.RDates})
	ri.exrules.iters = append(ri.exrules.iters, &iterator{queue: r.ExDates})

	return ri
}

type recurrenceIterator struct {
	rrules  *groupIterator
	exrules *groupIterator
}

func (ri *recurrenceIterator) Peek() *time.Time {
	next := ri.rrules.Peek()

	for {
		if next == nil {
			return nil
		}

		nextException := ri.exrules.Peek()

		if nextException != nil && nextException.Before(*next) {
			ri.exrules.Next()
			continue
		}

		if nextException != nil && nextException.Equal(*next) {
			next = ri.rrules.Next()

			continue
		}

		break
	}

	return next
}

func (ri *recurrenceIterator) Next() *time.Time {
	t := ri.Peek()
	ri.rrules.Next()
	return t
}
