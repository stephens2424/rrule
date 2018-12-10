package rrule

import (
	"strings"
	"time"
)

// Recurrence expresses a complex pattern of repeating events composed of individual
// patterns and extra days that are filtered by exclusion patterns and days.
type Recurrence struct {
	// Dtstart specifies the time to begin recurrence. The location of Dtstart is
	// the location that will be used to process the recurrence.
	Dtstart time.Time

	FloatingLocation bool // If true, Dtstart will be encoded in local time.

	// Patterns and instances to include. Repeated instances are included only
	// once, even if defined by multiple patterns.
	RRules []*RRule
	RDates []time.Time

	// Patterns and instances to exclude. These take precedence over the
	// inclusions. Note: this feature was deprecated in RFC5545, noting its
	// limited (and buggy) adoption and real-world use case. It is
	// implemented here, nonetheless, for maximum flexibility and
	// compatibility.
	ExRules []*RRule
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

func (r *Recurrence) SetDtstart() {
	for _, rr := range r.RRules {
		rr.Dtstart = r.Dtstart
	}
	for _, rr := range r.ExRules {
		rr.Dtstart = r.Dtstart
	}
}

// All returns all instances from the beginning of the iterator up to a limited
// number. If the limit is 0, all instances are returned, which may be an
// unbounded operation and loop forever.
//
// TODO: bound all operations to the maximum time.Time.
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

func (r Recurrence) Iterator() Iterator {
	r.SetDtstart()

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
