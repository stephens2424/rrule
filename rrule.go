package rrule

import (
	"errors"
	"fmt"
	"sort"
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
	if len(seconds) == 0 {
		return alwaysValid
	}

	m := intmap(seconds)

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
	if len(minutes) == 0 {
		return alwaysValid
	}

	m := intmap(minutes)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Minute()]
	}
}

func validHour(hours []int) validFunc {
	if len(hours) == 0 {
		return alwaysValid
	}

	m := intmap(hours)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Hour()]
	}
}

// validWeekday ignores the N modifier of QualifiedWeekday
func validWeekday(weekdays []QualifiedWeekday) validFunc {
	if len(weekdays) == 0 {
		return alwaysValid
	}

	m := weekdaymap(weekdays)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Weekday()]
	}
}

func validMonthDay(monthdays []int) validFunc {
	if len(monthdays) == 0 {
		return alwaysValid
	}

	m := intmap(monthdays)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Day()]
	}
}

func validWeek(weeks []int) validFunc {
	if len(weeks) == 0 {
		return alwaysValid
	}

	m := intmap(weeks)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[1+t.YearDay()/7]
	}
}

func validMonth(months []time.Month) validFunc {
	if len(months) == 0 {
		return alwaysValid
	}

	m := monthmap(months)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.Month()]
	}
}

func validYearDay(yeardays []int) validFunc {
	if len(yeardays) == 0 {
		return alwaysValid
	}

	m := intmap(yeardays)

	return func(t *time.Time) bool {
		if t == nil {
			return false
		}
		return m[t.YearDay()]
	}
}

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

type Frequency int

func (f Frequency) String() string {
	switch f {
	case Secondly:
		return "SECONDLY"
	case Minutely:
		return "MINUTELY"
	case Hourly:
		return "HOURLY"
	case Daily:
		return "DAILY"
	case Weekly:
		return "WEEKLY"
	case Monthly:
		return "MONTHLY"
	case Yearly:
		return "YEARLY"
	}
	return ""
}

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

type InvalidBehavior int

const (
	OmitInvalid InvalidBehavior = iota
	NextInvalid
	PrevInvalid
)

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

func (wd QualifiedWeekday) String() string {
	wdStr := WeekdayString(wd.WD)

	if wd.N == 0 {
		return wdStr
	}

	return fmt.Sprintf("%d%s", wd.N, wdStr)
}

// WeekdayString returns a weekday formatted as the two-letter string used in RFC5545.
func WeekdayString(wd time.Weekday) string {
	var wdStr string
	switch wd {
	case time.Sunday:
		wdStr = "SU"
	case time.Monday:
		wdStr = "MO"
	case time.Tuesday:
		wdStr = "TU"
	case time.Wednesday:
		wdStr = "WE"
	case time.Thursday:
		wdStr = "TH"
	case time.Friday:
		wdStr = "FR"
	case time.Saturday:
		wdStr = "SA"
	}
	return wdStr
}
