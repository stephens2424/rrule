package rrule

import (
	"sort"
	"time"
)

func limitBySetPos(tt []time.Time, setpos []int) []time.Time {
	if len(setpos) == 0 {
		return tt
	}

	// a map of tt indexes to include
	include := map[int]bool{}

	for _, sp := range setpos {
		if sp < 0 {
			sp = len(tt) + sp
		} else {
			sp-- // setpos is 1-indexed in the rrule. adjust here
		}

		include[sp] = true
	}

	ret := make([]time.Time, 0, len(include))
	for included := range include {
		if len(tt) > included {
			ret = append(ret, tt[included])
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Before(ret[j])
	})

	return ret
}

func limitInstancesBySetPos(tt []int, setpos []int) []int {
	if len(setpos) == 0 {
		return tt
	}

	// a map of tt indexes to include
	include := make(map[int]bool, len(setpos))

	for _, sp := range setpos {
		if sp < 0 {
			sp = len(tt) + sp
		} else {
			sp-- // setpos is 1-indexed in the rrule. adjust here
		}

		include[sp] = true
	}

	ret := make([]int, 0, len(include))
	for included := range include {
		if len(tt) > included {
			ret = append(ret, tt[included])
		}
	}

	sort.Ints(ret)

	return ret
}

func combineLimiters(ll ...validFunc) func(t *time.Time) bool {
	return func(t *time.Time) bool {
		for _, l := range ll {
			if !l(t) {
				return false
			}
		}
		return true
	}
}

func checkLimiters(t *time.Time, ll ...validFunc) bool {
	for _, l := range ll {
		if !l(t) {
			return false
		}
	}
	return true
}
