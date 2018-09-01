package rrule

import (
	"sync"
	"time"
)

type setposIterator struct {
	validPos   []int
	underlying Iterator

	readOnce sync.Once
	queue    []*time.Time
}

func (si *setposIterator) Next() *time.Time {
	si.readOnce.Do(func() {
		var list []*time.Time

		for i := 0; true; i++ {
			t := si.underlying.Next()
			if t == nil {
				break
			}

			list = append(list, t)
		}

		m := map[int]struct{}{}
		for _, p := range si.validPos {
			if p < 0 {
				p = len(list) + p
			}
			m[p] = struct{}{}
		}

		for i, t := range list {
			if _, ok := m[i]; ok {
				si.queue = append(si.queue, t)
			}
		}
	})

	if len(si.queue) == 0 {
		return nil
	}

	v := si.queue[0]
	si.queue = si.queue[1:]

	return v
}
