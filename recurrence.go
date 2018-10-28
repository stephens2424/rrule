package rrule

import "time"

type Recurrence struct {
	// Dtstart specifies the time to begin recurrence. The location of Dtstart is
	// the location that will be used to process the recurrence.
	Dtstart time.Time

	// FLoatingLocation, if true, Dtstart will be encoded in local time.
	FloatingLocation bool

	RRules  []*RRule
	ExRules []*RRule // note this feature was deprecated in RFC5545
	RDates  []time.Time
	ExDates []time.Time
}

func (r *Recurrence) SetDtstart() {
	for _, rr := range r.RRules {
		rr.Dtstart = r.Dtstart
	}
	for _, rr := range r.ExRules {
		rr.Dtstart = r.Dtstart
	}
}

func (r Recurrence) Iterator() Iterator {
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
	rdates  *iterator
	exdates *iterator
}

func (ri *recurrenceIterator) Peek() *time.Time {
	next := ri.rrules.Peek()

	if next == nil {
		return nil
	}

	nextException := ri.exrules.Peek()

	for {
		if nextException != nil && nextException.Before(*next) {
			ri.exrules.Next()
			continue
		}

		if nextException != nil && nextException.Equal(*next) {
			ri.rrules.Next()
			continue
		}

		break
	}

	return next
}

func (ri *recurrenceIterator) Next() *time.Time {
	t := ri.rrules.Peek()
	ri.rrules.Next()
	return t
}
