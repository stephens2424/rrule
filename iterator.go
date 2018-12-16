package rrule

import (
	"time"
)

// Iterator scans over a series of times.
type Iterator interface {
	// Peek returns the next time without advancing the iterator, or nil if the iterator has ended.
	Peek() *time.Time

	// Next returns the next time and advances the iterator. Nil is returned if the iterator has ended.
	Next() *time.Time
}

type iterator struct {
	queue       []time.Time
	totalQueued uint64
	queueCap    uint64
	minTime     time.Time
	maxTime     time.Time
	pastMaxTime bool

	// next finds the next key time.
	next func() *time.Time

	// variations returns all the possible variations
	// of the key time t
	variations func(t *time.Time) []time.Time

	// valid determines if a particular key time is a valid recurrence.
	valid func(t *time.Time) bool

	setpos []int
}

func (i *iterator) Next() *time.Time {
	t := i.Peek()
	if len(i.queue) > 1 {
		i.queue = i.queue[1:]
	} else if len(i.queue) == 1 {
		i.queue = nil
	}
	return t
}

func (i *iterator) Peek() *time.Time {
	if len(i.queue) > 0 {
		r := i.queue[0]
		return &r
	}

	if i.queueCap > 0 {
		if i.totalQueued >= i.queueCap {
			return nil
		}
	}

	if i.next == nil {
		return nil
	}

	for {
		if i.pastMaxTime {
			return nil
		}

		key := i.next()
		if key == nil {
			return nil
		}

		if !i.valid(key) {
			continue
		}

		variations := i.variations(key)

		// remove any variations before the min time
		for len(variations) > 0 && variations[0].Before(i.minTime) {
			variations = variations[1:]
		}

		// remove any variations after the max time
		if !i.maxTime.IsZero() {
			for idx, v := range variations {
				if v.After(i.maxTime) {
					variations = variations[:idx]
					i.pastMaxTime = true
					break
				}
			}
		}

		// if we're left with nothing (or started there) skip this key time
		if len(variations) == 0 {
			continue
		}

		if i.queueCap > 0 {
			if i.totalQueued+uint64(len(variations)) > i.queueCap {
				variations = variations[:i.queueCap-i.totalQueued]
			}
		}

		i.totalQueued += uint64(len(variations))

		i.queue = variations[:]
		return &variations[0]
	}
}

// https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go
var absoluteMaxTime = time.Date(219248499, 01, 01, 0, 0, 0, 0, time.UTC)
