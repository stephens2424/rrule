package rrule

import "time"

type groupIterator struct {
	currentMin *int
	iters      []Iterator
}

func groupIteratorFromRRules(rrules []*RRule) *groupIterator {
	gi := &groupIterator{}
	for _, rr := range rrules {
		gi.iters = append(gi.iters, rr.Iterator())
	}

	return gi
}

func (gi *groupIterator) Peek() *time.Time {
	if gi.currentMin != nil {
		return gi.iters[*gi.currentMin].Peek()
	}

	var min *time.Time
	var minIdx int = -1

	for i, iter := range gi.iters {
		t := iter.Peek()
		if t != nil {
			if min == nil {
				min = t
			} else {
				if t.Before(*min) {
					min = t
					minIdx = i
				}
			}
		}
	}

	if minIdx >= 0 {
		gi.currentMin = &minIdx
	}

	return min
}

func (gi *groupIterator) Next() *time.Time {
	if gi.currentMin == nil {
		gi.Peek()
	}

	if gi.currentMin == nil {
		// still don't have a min time, so the iterators must all have ended
		return nil
	}

	idx := *gi.currentMin
	gi.currentMin = nil
	return gi.iters[idx].Next()
}
