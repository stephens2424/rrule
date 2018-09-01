package rrule

import (
	"fmt"
	"time"
)

type RRule struct {
	Frequency Frequency

	// Either Until or Count may be set, but not both
	Until time.Time
	Count uint64

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

// validFunc is a kind of function that checks if a time is valid against a rule. It returns true if the time is valid.
// A pointer is accepted in order to avoid the memory copy of the entire time structure. Nil is never considered valid.
type validFunc func(t *time.Time) bool

func alwaysValid(t *time.Time) bool {
	return t != nil
}

func validSecond(seconds []int) validFunc {
	m := intmap(seconds)

	if len(seconds) == 0 {
		return alwaysValid
	}

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Second()]
	}
}

func intmap(ints []int) map[int]bool {
	m := make(map[int]bool, len(ints))
	for _, v := range ints {
		m[v] = true
	}
	return m
}

func weekdaymap(weekdays []QualifiedWeekday) map[time.Weekday]bool {
	m := make(map[time.Weekday]bool, len(weekdays))
	for _, v := range weekdays {
		m[v.WD] = true
	}
	return m
}

func monthmap(months []time.Month) map[time.Month]bool {
	m := make(map[time.Month]bool, len(months))
	for _, v := range months {
		m[v] = true
	}
	return m
}

func validMinute(minutes []int) validFunc {
	m := intmap(minutes)

	if len(minutes) == 0 {
		return alwaysValid
	}
	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Minute()]
	}
}

func validHour(hours []int) validFunc {
	m := intmap(hours)

	if len(hours) == 0 {
		return alwaysValid
	}
	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Hour()]
	}
}

// validWeekday ignores the N modifier of QualifiedWeekday
func validWeekday(weekdays []QualifiedWeekday) validFunc {
	m := weekdaymap(weekdays)

	if len(weekdays) == 0 {
		return alwaysValid
	}

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Weekday()]
	}
}

func validMonthDay(monthdays []int) validFunc {
	m := intmap(monthdays)

	if len(monthdays) == 0 {
		return alwaysValid
	}
	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Day()]
	}
}

func validWeek(weeks []int) validFunc {
	m := intmap(weeks)

	if len(weeks) == 0 {
		return alwaysValid
	}
	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[1+t.YearDay()/7]
	}
}

func validMonth(months []time.Month) validFunc {
	m := monthmap(months)

	if len(months) == 0 {
		return alwaysValid
	}
	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Month()]
	}
}

func validYearDay(yeardays []int) validFunc {
	m := intmap(yeardays)

	if len(yeardays) == 0 {
		return alwaysValid
	}
	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.YearDay()]
	}
}

func checkLimiters(t time.Time, ll ...validFunc) bool {
	for _, l := range ll {
		if !l(&t) {
			return false
		}
	}
	return true
}

type Frequency int

const (
	Secondly Frequency = iota
	Minutely
	Hourly
	Daily
	Weekly
	Monthly
	Yearly
)

type Iterator interface {
	Next() *time.Time
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

func (rrule RRule) Iterator() Iterator {
	iter := rrule.noSetposIterator()

	if len(rrule.BySetPos) == 0 {
		return iter
	}

	if rrule.Frequency != Yearly && rrule.Frequency != Monthly {
		for _, wd := range rrule.ByWeekdays {
			if wd.N != 0 {
				panic("BYDAY entries may only specify a numeric component when the frequency is YEARLY or MONTHLY")
			}
		}
	}
	if rrule.Frequency == Yearly && len(rrule.ByWeekNumbers) > 0 {
		for _, wd := range rrule.ByWeekdays {
			if wd.N != 0 {
				panic("BYDAY entries must not specify a numeric component when the frequency is YEARLY and a BYWEEKNO rule is present")
			}
		}
	}

	// TODO: don't need a pre-caching iterator when all the BySetPos values
	// are >= 0.
	// TODO: COUNT and UNTIL need to be evaluated after BYSETPOS rules.
	return &setposIterator{
		validPos:   rrule.BySetPos,
		underlying: iter,
	}
}

func (rrule RRule) noSetposIterator() Iterator {
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

	return &iterator{
		minTime:  start,
		maxTime:  rrule.Until,
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.Add(time.Duration(interval) * time.Second)
			return &ret
		},

		valid: func(t time.Time) bool {
			return true
			return checkLimiters(t,
				validMonth(rrule.ByMonths),
				validWeek(rrule.ByWeekNumbers),
				validYearDay(rrule.ByYearDays),
				validMonthDay(rrule.ByMonthDays),
				validWeekday(rrule.ByWeekdays),
				validHour(rrule.ByHours),
				validMinute(rrule.ByMinutes),
			)
		},

		variations: func(t time.Time) []time.Time {
			return []time.Time{t}
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
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.Add(time.Duration(interval) * time.Minute)
			return &ret
		},

		valid: func(t time.Time) bool {
			return checkLimiters(t,
				validMonth(rrule.ByMonths),
				validWeek(rrule.ByWeekNumbers),
				validYearDay(rrule.ByYearDays),
				validMonthDay(rrule.ByMonthDays),
				validWeekday(rrule.ByWeekdays),
				validHour(rrule.ByHours),
				validMinute(rrule.ByMinutes),
			)
		},

		variations: func(t time.Time) []time.Time {
			tt := expandBySeconds([]time.Time{t}, rrule.BySeconds...)
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
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.Add(time.Duration(interval) * time.Hour)
			return &ret
		},

		valid: func(t time.Time) bool {
			return true
			return checkLimiters(t,
				validMonth(rrule.ByMonths),
				validWeek(rrule.ByWeekNumbers),
				validYearDay(rrule.ByYearDays),
				validMonthDay(rrule.ByMonthDays),
				validWeekday(rrule.ByWeekdays),
				validHour(rrule.ByHours),
			)
		},

		variations: func(t time.Time) []time.Time {
			tt := expandByMinutes([]time.Time{t}, rrule.ByMinutes...)
			tt = expandBySeconds(tt, rrule.BySeconds...)
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
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(0, interval, 0)
			return &ret
		},

		valid: func(t time.Time) bool {
			if len(rrule.ByMonthDays) > 0 {
				return checkLimiters(t,
					validMonth(rrule.ByMonths),
					validMonthDay(rrule.ByMonthDays),
					validWeekday(rrule.ByWeekdays),
				)
			} else {
				return checkLimiters(t,
					validMonth(rrule.ByMonths),
					validMonthDay(rrule.ByMonthDays),
				)
			}
		},

		variations: func(t time.Time) []time.Time {
			tt := expandBySeconds([]time.Time{t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)
			if len(rrule.ByMonthDays) > 0 {
				tt = expandByMonthDays(tt, rrule.ByMonthDays...)
			} else if len(rrule.ByWeekdays) > 0 {
				tt = expandMonthByWeekdays(tt, rrule.IB, rrule.ByWeekdays...)
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
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(0, 0, interval)
			return &ret
		},

		valid: func(t time.Time) bool {
			return checkLimiters(t,
				validMonth(rrule.ByMonths),
				validMonthDay(rrule.ByMonthDays),
				validWeekday(rrule.ByWeekdays),
			)
		},

		variations: func(t time.Time) []time.Time {
			tt := expandBySeconds([]time.Time{t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)
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
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(0, 0, interval*7)
			return &ret
		},

		valid: func(t time.Time) bool {
			return checkLimiters(t,
				validMonth(rrule.ByMonths),
			)
		},

		variations: func(t time.Time) []time.Time {
			tt := expandBySeconds([]time.Time{t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)
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
		queueCap: rrule.Count,
		next: func() *time.Time {
			ret := current // copy current
			current = current.AddDate(interval, 0, 0)
			return &ret
		},

		valid: func(t time.Time) bool {

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

		variations: func(t time.Time) []time.Time {
			tt := expandBySeconds([]time.Time{t}, rrule.BySeconds...)
			tt = expandByMinutes(tt, rrule.ByMinutes...)
			tt = expandByHours(tt, rrule.ByHours...)

			tt = expandByMonthDays(tt, rrule.ByMonthDays...)
			tt = expandByYearDays(tt, rrule.ByYearDays...)
			tt = expandByWeekNumbers(tt, rrule.weekStart(), rrule.ByWeekNumbers...)
			tt = expandByMonths(tt, rrule.IB, rrule.ByMonths...)

			// see note 2 on page 44 of RFC 5545, including erratum 3779.
			if len(rrule.ByYearDays) == 0 && len(rrule.ByMonthDays) == 0 {
				if len(rrule.ByMonths) != 0 {
					tt = expandMonthByWeekdays(tt, rrule.IB, rrule.ByWeekdays...)
				} else {
					tt = expandYearByWeekdays(tt, rrule.IB, rrule.ByWeekdays...)
				}
			}
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

func expandBySeconds(tt []time.Time, seconds ...int) []time.Time {
	if len(seconds) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(seconds))
	for _, t := range tt {
		for _, s := range seconds {
			e = append(e, time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), s, t.Nanosecond(), t.Location()))
		}
	}

	return e
}

func expandByMinutes(tt []time.Time, minutes ...int) []time.Time {
	if len(minutes) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(minutes))
	for _, t := range tt {
		for _, m := range minutes {
			e = append(e, time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, t.Second(), t.Nanosecond(), t.Location()))
		}
	}

	return e
}

func expandByHours(tt []time.Time, hours ...int) []time.Time {
	if len(hours) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(hours))
	for _, t := range tt {
		for _, h := range hours {
			e = append(e, time.Date(t.Year(), t.Month(), t.Day(), h, t.Minute(), t.Second(), t.Nanosecond(), t.Location()))
		}
	}

	return e
}

func expandByWeekdays(tt []time.Time, weekStart time.Weekday, weekdays ...QualifiedWeekday) []time.Time {
	if len(weekdays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(weekdays))
	for _, t := range tt {
		t = backToWeekday(t, weekStart)
		for _, wd := range weekdays {
			e = append(e, forwardToWeekday(t, wd.WD))
		}
	}

	return e
}

func backToWeekday(t time.Time, day time.Weekday) time.Time {
	for t.Weekday() != day {
		t = t.AddDate(0, 0, -1)
	}
	return t
}

func forwardToWeekday(t time.Time, day time.Weekday) time.Time {
	for t.Weekday() != day {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

func expandMonthByWeekdays(tt []time.Time, ib InvalidBehavior, weekdays ...QualifiedWeekday) []time.Time {
	if len(weekdays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt))
	for _, t := range tt {
		for _, wd := range weekdays {
			e = append(e, weekdaysInMonth(t, wd, ib)...)
		}
	}

	return e
}

func expandYearByWeekdays(tt []time.Time, ib InvalidBehavior, weekdays ...QualifiedWeekday) []time.Time {
	if len(weekdays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt))
	for _, t := range tt {
		for _, wd := range weekdays {
			e = append(e, weekdaysInYear(t, wd, ib)...)
		}
	}

	return e

}

type InvalidBehavior int

const (
	OmitInvalid InvalidBehavior = iota
	NextInvalid
	PrevInvalid
)

func weekdaysInMonth(t time.Time, wd QualifiedWeekday, ib InvalidBehavior) []time.Time {
	allWDs := make([]time.Time, 0, 5)

	day := time.Date(t.Year(), t.Month(), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	for day.Weekday() != wd.WD {
		day = day.AddDate(0, 0, 1)
	}

	for {
		allWDs = append(allWDs, day)
		day = day.AddDate(0, 0, 7)
		if day.Month() != t.Month() {
			break
		}
	}

	if wd.N == 0 {
		return allWDs
	}

	if wd.N > 0 {
		if wd.N > len(allWDs) {
			switch ib {
			case OmitInvalid:
				return nil
			case PrevInvalid:
				idx := len(allWDs) - 1
				return allWDs[idx:idx]
			case NextInvalid:
				return []time.Time{allWDs[len(allWDs)-1].AddDate(0, 0, 7)}
			}
		}
		return []time.Time{allWDs[wd.N-1]}
	}

	// an example of the following logic:
	//
	// -1 in a list of 4 ..
	// 	- the index becomes 3, which is the last index
	//	  which corresponds to "the last instance"
	// -3 in a list of 4 ..
	//	- the index becomes 1, which is the third from last
	// -7 in a list of 4 ..
	//	- the index becomes -3, which should trigger invalid behavior
	idx := len(allWDs) + wd.N

	if idx < 0 || idx > len(allWDs) {
		switch ib {
		case OmitInvalid:
			return nil
		case PrevInvalid:
			return []time.Time{allWDs[0].AddDate(0, 0, -7)}
		case NextInvalid:
			return allWDs[0:0]
		}
	}

	return []time.Time{allWDs[idx]}

}

func weekdaysInYear(t time.Time, wd QualifiedWeekday, ib InvalidBehavior) []time.Time {
	allWDs := make([]time.Time, 0, 5)

	// start on first of year
	day := time.Date(t.Year(), 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	// scan til the first relevant weekday of the year
	for day.Weekday() != wd.WD {
		day = day.AddDate(0, 0, 1)
	}

	// scan over every week of the year
	for {
		allWDs = append(allWDs, day)
		day = day.AddDate(0, 0, 7)
		if day.Year() != t.Year() {
			break
		}
	}

	if wd.N == 0 {
		// no index specified, return all.
		return allWDs
	}

	if wd.N > 0 {
		// positive index specified. count to the correct instance
		if wd.N > len(allWDs) {
			switch ib {
			case OmitInvalid:
				return nil
			case PrevInvalid:
				idx := len(allWDs) - 1
				return allWDs[idx:idx]
			case NextInvalid:
				return []time.Time{allWDs[len(allWDs)-1].AddDate(0, 0, 7)}
			}
		}
		return []time.Time{allWDs[wd.N-1]}
	}

	// negative index specified. count backwards to the correct instance

	// an example of the following logic:
	//
	// -1 in a list of 4 ..
	// 	- the index becomes 3, which is the last index
	//	  which corresponds to "the last instance"
	// -3 in a list of 4 ..
	//	- the index becomes 1, which is the third from last
	// -7 in a list of 4 ..
	//	- the index becomes -3, which should trigger invalid behavior
	idx := len(allWDs) + wd.N

	if idx < 0 || idx > len(allWDs) {
		switch ib {
		case OmitInvalid:
			return nil
		case PrevInvalid:
			return []time.Time{allWDs[0].AddDate(0, 0, -7)}
		case NextInvalid:
			return allWDs[0:0]
		}
	}

	return []time.Time{allWDs[idx]}
}

func expandByMonthDays(tt []time.Time, monthdays ...int) []time.Time {
	if len(monthdays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(monthdays))
	for _, t := range tt {
		for _, md := range monthdays {
			e = append(e, time.Date(t.Year(), t.Month(), md, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()))
		}
	}

	return e
}

func expandByYearDays(tt []time.Time, yeardays ...int) []time.Time {
	if len(yeardays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(yeardays))
	for _, t := range tt {
		yearStart := time.Date(t.Year(), time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

		for _, yd := range yeardays {
			e = append(e, yearStart.AddDate(0, 0, yd))
		}
	}

	return e
}

func expandByWeekNumbers(tt []time.Time, weekStarts time.Weekday, weekNumbers ...int) []time.Time {
	if len(weekNumbers) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(weekNumbers))
	for _, t := range tt {
		yearStart := time.Date(t.Year(), time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		yearStart = forwardToWeekday(yearStart, t.Weekday())

		for _, w := range weekNumbers {
			e = append(e, yearStart.AddDate(0, 0, (w-1)*7))
		}
	}

	return e
}

func expandByMonths(tt []time.Time, ib InvalidBehavior, months ...time.Month) []time.Time {
	if len(months) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(months))
	for _, t := range tt {
		for _, m := range months {
			set := time.Date(t.Year(), m, t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
			if set.Month() != m {
				switch ib {
				case PrevInvalid:
					set = time.Date(t.Year(), t.Month()+1, -1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
					e = append(e, set)
				case NextInvalid:
					set = time.Date(t.Year(), t.Month()+1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
					e = append(e, set)
				case OmitInvalid:
					// do nothing
				}
			} else {
				e = append(e, set)
			}
		}
	}

	return e
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
	variations func(t time.Time) []time.Time

	// valid determines if a particular key time is a valid recurrence.
	valid func(t time.Time) bool
}

func (i *iterator) Next() *time.Time {
	if len(i.queue) > 0 {
		r := i.queue[0]
		i.queue = i.queue[1:]
		return &r
	}

	if i.queueCap > 0 {
		if i.totalQueued >= i.queueCap {
			return nil
		}
	}

	for {
		if i.pastMaxTime {
			return nil
		}

		key := i.next()
		if key == nil {
			return nil
		}

		if !i.valid(*key) {
			continue
		}

		variations := i.variations(*key)

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

		i.queue = variations[1:]
		return &variations[0]
	}
}

// QualifiedWeekday can represent a day of the week, or a certain instance
// of that day of the week.
type QualifiedWeekday struct {
	// N, when specified says which instance of the weekday relative to
	// some greater duration. -3 would be "third from the last"
	N  int
	WD time.Weekday
}
