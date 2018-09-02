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

	count uint64
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

				// check that we don't queue past the count limit
				if si.count > 0 && uint64(len(si.queue)) >= si.count {
					return
				}
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

type posSetposIterator struct {
	underlying Iterator
	index      int
	valid      map[int]bool

	returnCount uint64
	maxCount    uint64
}

func (psi *posSetposIterator) Next() *time.Time {
	if psi.returnCount >= psi.maxCount {
		return nil
	}

	for {
		t := psi.underlying.Next()
		if t == nil {
			return nil
		}

		// capture the current index and increment for the next iteration
		idx := psi.index
		psi.index++

		if psi.valid[idx] {
			psi.returnCount++
			return t
		}
	}
}
