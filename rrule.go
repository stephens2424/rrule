package rrule

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// RRule represents a single pattern within a recurrence.
type RRule struct {
	Frequency Frequency

	// Either Until or Count may be set, but not both
	Until         time.Time
	UntilFloating bool // If true, the RRule will encode using local time (no offset).

	Count uint64

	// Dtstart is not actually part of the RRule when
	// encoded, but it's included here as a field because
	// it's required when expading the recurrence.
	//
	// If zero, time.Now is used when an iterator is generated
	Dtstart time.Time

	// 0 means the default value, which is 1.
	Interval int

	BySeconds     []int // 0 to 59
	ByMinutes     []int // 0 to 59
	ByHours       []int // 0 to 23
	ByWeekdays    []QualifiedWeekday
	ByMonthDays   []int // 1 to 31
	ByWeekNumbers []int // 1 to 53
	ByMonths      []time.Month
	ByYearDays    []int // 1 to 366
	BySetPos      []int // -366 to 366

	IB InvalidBehavior

	WeekStart *time.Weekday // if nil, Monday
}

func (rrule RRule) All(limit int) []time.Time {
	it := rrule.Iterator()
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

func (rrule RRule) Validate() error {
	if rrule.Frequency != Yearly && rrule.Frequency != Monthly {
		for _, wd := range rrule.ByWeekdays {
			if wd.N != 0 {
				return errors.New("BYDAY entries may only specify a numeric component when the frequency is YEARLY or MONTHLY")
			}
		}
	}
	if rrule.Frequency == Yearly && len(rrule.ByWeekNumbers) > 0 {
		for _, wd := range rrule.ByWeekdays {
			if wd.N != 0 {
				return errors.New("BYDAY entries must not specify a numeric component when the frequency is YEARLY and a BYWEEKNO rule is present")
			}
		}
	}

	if rrule.Frequency == Weekly && len(rrule.ByMonthDays) > 0 {
		return errors.New("WEEKLY recurrences must not include BYMONTHDAY")
	}

	if len(rrule.BySetPos) != 0 {
		if len(rrule.BySeconds) == 0 &&
			len(rrule.ByMinutes) == 0 &&
			len(rrule.ByHours) == 0 &&
			len(rrule.ByWeekdays) == 0 &&
			len(rrule.ByMonthDays) == 0 &&
			len(rrule.ByWeekNumbers) == 0 &&
			len(rrule.ByMonths) == 0 &&
			len(rrule.ByYearDays) == 0 {
			return errors.New("BYSETPOS rules must be used in conjunction with at least one other BYXXX rule part")
		}
	}

	if rrule.Count != 0 && !rrule.Until.IsZero() {
		return errors.New("COUNT and UNTIL must not appear in the same RRULE")
	}

	for _, sp := range rrule.BySetPos {
		if sp == 0 || sp < -366 || sp > 366 {
			return errors.New("BYSETPOS values must be between [-366,-1] or [1,366].")
		}
	}

	return nil
}

func (rrule RRule) Iterator() Iterator {
	err := rrule.Validate()
	if err != nil {
		panic(err)
	}

	switch rrule.Frequency {
	case Secondly:
		return setSecondly(rrule)
	case Minutely:
		return setMinutely(rrule)
	case Hourly:
		return setHourly(rrule)
	case Daily:
		return setDaily(rrule)
	case Weekly:
		return setWeekly(rrule)
	case Monthly:
		return setMonthly(rrule)
	case Yearly:
		return setYearly(rrule)
	default:
		panic(fmt.Sprintf("invalid frequency %v", rrule.Frequency))
	}
}

func setSecondly(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	current := start

	nextFn := func() *time.Time {
		ret := current // copy current
		current = current.Add(time.Duration(interval) * time.Second)
		return &ret
	}

	// An rrule with Interval of 1 and BySeconds will potentially cycle through
	// many seconds that get skipped. This is a fairly expensive case, but can be
	// short-circuited by skipping to each subsequent BySeconds point instead of
	// each second.
	if interval == 1 && len(rrule.BySeconds) > 0 {
		seconds := []int{}
		for _, s := range rrule.BySeconds {
			if s < 0 {
				s += 60
			}
			seconds = append(seconds, s)
		}

		sort.Ints(seconds)
		initialSecond := start.Second()
		loopIdx := 0
		wentPastInitial := false
		var firstDiff time.Duration

		var secondsLooper []time.Duration
		for i, s := range seconds {
			if !wentPastInitial && s > initialSecond {
				wentPastInitial = true
				loopIdx = i
				firstDiff = time.Duration(s-initialSecond) * time.Second
			}

			nextIdx := i + 1
			if nextIdx == len(seconds) {
				secondsLooper = append(secondsLooper, time.Duration(60+seconds[0]-seconds[i])*time.Second)
			} else {
				secondsLooper = append(secondsLooper, time.Duration(seconds[nextIdx]-seconds[i])*time.Second)
			}
		}

		if !wentPastInitial {
			// all the BySecond terms are lower numbers than the start time second, so we need to wrap around for the first diff
			firstDiff = time.Duration(seconds[0]+60-initialSecond) * time.Second
		}

		secondsLooperFn := func() *time.Time {
			ret := current // copy
			current = current.Add(secondsLooper[loopIdx])
			loopIdx++
			if loopIdx >= len(secondsLooper) {
				loopIdx = 0
			}
			return &ret
		}

		var afterFirst bool

		// return an initial function that does the first initial
		nextFn = func() *time.Time {
			if afterFirst {
				return secondsLooperFn()
			}

			ret := current // copy
			current = current.Add(firstDiff)
			afterFirst = true
			return &ret
		}
	}

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		queueCap: rrule.Count,
		setpos:   rrule.BySetPos,
		next:     nextFn,

		valid: combineLimiters(
			validSecond(rrule.BySeconds),
			validMinute(rrule.ByMinutes),
			validHour(rrule.ByHours),
			validWeekday(rrule.ByWeekdays),
			validMonthDay(rrule.ByMonthDays),
			validMonth(rrule.ByMonths),
			validWeek(rrule.ByWeekNumbers),
			validYearDay(rrule.ByYearDays),
		),

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}
			return []time.Time{*t}
		},
	}
}

func setMinutely(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	current := start

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		setpos:   rrule.BySetPos,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.Add(time.Duration(interval) * time.Minute)
			return &ret
		},

		valid: combineLimiters(
			validMonth(rrule.ByMonths),
			validWeek(rrule.ByWeekNumbers),
			validYearDay(rrule.ByYearDays),
			validMonthDay(rrule.ByMonthDays),
			validWeekday(rrule.ByWeekdays),
			validHour(rrule.ByHours),
			validMinute(rrule.ByMinutes),
		),

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}
			tt := expandBySeconds([]time.Time{*t}, rrule.BySeconds...)
			tt = limitBySetPos(tt, rrule.BySetPos)
			return tt
		},
	}
}

func setHourly(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	current := start

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		setpos:   rrule.BySetPos,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.Add(time.Duration(interval) * time.Hour)
			return &ret
		},

		valid: combineLimiters(
			validMonth(rrule.ByMonths),
			validWeek(rrule.ByWeekNumbers),
			validYearDay(rrule.ByYearDays),
			validMonthDay(rrule.ByMonthDays),
			validWeekday(rrule.ByWeekdays),
			validHour(rrule.ByHours),
		),

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}
			tt := expandByMinutes([]time.Time{*t}, rrule.ByMinutes...)
			tt = expandBySeconds(tt, rrule.BySeconds...)
			tt = limitBySetPos(tt, rrule.BySetPos)
			return tt
		},
	}
}

func setMonthly(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	current := start

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		setpos:   rrule.BySetPos,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(0, interval, 0)
			return &ret
		},

		valid: func(t *time.Time) bool {
			if t == nil {
				return false
			}
			if len(rrule.ByMonthDays) > 0 {
				return checkLimiters(t,
					validMonth(rrule.ByMonths),
					validWeekday(rrule.ByWeekdays),
				)
			} else {
				return checkLimiters(t,
					validMonth(rrule.ByMonths),
				)
			}
		},

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}
			tt := expandBySeconds([]time.Time{*t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)
			if len(rrule.ByMonthDays) > 0 {
				tt = expandByMonthDays(tt, rrule.ByMonthDays...)
			} else if len(rrule.ByWeekdays) > 0 {
				tt = expandMonthByWeekdays(tt, rrule.IB, rrule.BySetPos, rrule.ByWeekdays...)
			}
			return tt
		},
	}
}

func setDaily(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	current := start

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		setpos:   rrule.BySetPos,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(0, 0, interval)
			return &ret
		},

		valid: combineLimiters(
			validMonth(rrule.ByMonths),
			validMonthDay(rrule.ByMonthDays),
			validWeekday(rrule.ByWeekdays),
		),

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}
			tt := expandBySeconds([]time.Time{*t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)
			tt = limitBySetPos(tt, rrule.BySetPos)
			return tt
		},
	}
}

func setWeekly(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	current := start

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		setpos:   rrule.BySetPos,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(0, 0, interval*7)
			return &ret
		},

		valid: combineLimiters(
			validMonth(rrule.ByMonths),
		),

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}
			tt := expandBySeconds([]time.Time{*t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)
			tt = limitBySetPos(tt, rrule.BySetPos)
			tt = expandByWeekdays(tt, rrule.weekStart(), rrule.ByWeekdays...)
			return tt
		},
	}
}

func setYearly(rrule RRule) *iterator {
	start := rrule.Dtstart
	if start.IsZero() {
		start = time.Now()
	}

	interval := 1
	if rrule.Interval != 0 {
		interval = rrule.Interval
	}

	current := start

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		setpos:   rrule.BySetPos,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(interval, 0, 0)
			return &ret
		},

		valid: func(t *time.Time) bool {
			if t == nil {
				return false
			}

			// see note 2 on page 44 of RFC 5545, including erratum 3747.
			if len(rrule.ByYearDays) > 0 || len(rrule.ByMonthDays) > 0 {
				return checkLimiters(t,
					validMonth(rrule.ByMonths),
					validWeekday(rrule.ByWeekdays),
				)
			}

			return checkLimiters(t,
				validMonth(rrule.ByMonths),
			)
		},

		variations: func(t *time.Time) []time.Time {
			if t == nil {
				return nil
			}

			tt := expandBySeconds([]time.Time{*t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)

			tt = expandByMonthDays(tt, rrule.ByMonthDays...)
			tt = expandByYearDays(tt, rrule.ByYearDays...)
			tt = expandByWeekNumbers(tt, rrule.weekStart(), rrule.ByWeekNumbers...)
			tt = expandByMonths(tt, rrule.IB, rrule.ByMonths...)

			// see note 2 on page 44 of RFC 5545, including erratum 3779.
			if len(rrule.ByYearDays) == 0 && len(rrule.ByMonthDays) == 0 {
				if len(rrule.ByMonths) != 0 {
					tt = expandMonthByWeekdays(tt, rrule.IB, nil, rrule.ByWeekdays...)
				} else {
					tt = expandYearByWeekdays(tt, rrule.IB, rrule.ByWeekdays...)
				}
			}

			tt = limitBySetPos(tt, rrule.BySetPos)
			return tt
		},
	}
}

func (rrule *RRule) weekStart() time.Weekday {
	if rrule.WeekStart == nil {
		return time.Monday
	}
	return *rrule.WeekStart
}
