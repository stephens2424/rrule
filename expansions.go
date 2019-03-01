package rrule

import (
	"sort"
	"time"
)

func expandBySeconds(tt []time.Time, seconds ...int) []time.Time {
	if len(seconds) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(seconds))
	for _, t := range tt {
		tmpl := t.Add(time.Duration(-1*t.Second()) * time.Second)
		for _, s := range seconds {
			if s < 0 {
				s += 60
			}
			e = append(e, tmpl.Add((time.Duration(s) * time.Second)))
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
		tmpl := t.Add(time.Duration(-1*t.Minute()) * time.Minute)
		for _, m := range minutes {
			if m < 0 {
				m += 60
			}
			e = append(e, tmpl.Add(time.Duration(m)*time.Minute))
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
		tmpl := t.Add(time.Duration(-1*t.Hour()) * time.Hour)
		for _, h := range hours {
			if h < 0 {
				h += 24
			}
			e = append(e, tmpl.Add(time.Duration(h)*time.Hour))
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

func expandByYearDays(tt []time.Time, ib InvalidBehavior, yeardays ...int) []time.Time {
	if len(yeardays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(yeardays))
	for _, t := range tt {
		yearStart := time.Date(t.Year(), time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		startYear := yearStart.Year()

		for _, yd := range yeardays {
			added := yearStart.AddDate(0, 0, yd-1) // subtract one because we start on the 1st, so if we want yearday 1, we actually want to advance 0.
			if added.Year() != startYear {
				switch ib {
				case OmitInvalid:
					// do nothing
				case NextInvalid:
					e = append(e, added)
				case PrevInvalid:
					e = append(e, added.AddDate(0, 0, -1))
				}
			} else {
				e = append(e, added)
			}
		}
	}

	return e
}

func expandByWeekNumbers(tt []time.Time, ib InvalidBehavior, weekStarts time.Weekday, byWeekdays []time.Weekday, weekNumbers ...int) []time.Time {
	if len(weekNumbers) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt)*len(weekNumbers))
	for _, t := range tt {
		ys := yearStart(t, weekStarts)

		byWeekdays := byWeekdays
		if len(byWeekdays) == 0 {
			// NOTE: the spec is not 100% clear on what to do in this case.
			// rrule.js, for instance, will default to returning the full
			// week. lib-recur seems to copy the weekday from the input
			// time. I'm going with the latter, since it seems more consistent
			// with the behavior you'd get on a BYMONTH clause.
			byWeekdays = []time.Weekday{t.Weekday()}
		}

		for _, w := range weekNumbers {
			ws := ys.AddDate(0, 0, (w-1)*7)

			if weekYearStart := yearStart(ws, weekStarts); weekYearStart.Year() != ys.Year() {
				// check that the week we generated is still within the proper
				// year, or if it ran over because the year did not have enough
				// weeks

				nextYearStart := yearStart(ys.AddDate(1, 0, 0), weekStarts)
				switch ib {
				case OmitInvalid:
					// do nothing
				case NextInvalid:
					for _, wd := range byWeekdays {
						e = append(e, forwardToWeekday(nextYearStart, wd))
					}
				case PrevInvalid:
					for _, wd := range byWeekdays {
						e = append(e, backToWeekday(nextYearStart, wd))
					}
				}
				continue
			}

			for _, wd := range byWeekdays {
				e = append(e, forwardToWeekday(ws, wd))
			}
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

// expandMonthByWeekdays does a special expansion of the month by weekdays. If
// bySetPos is not nil, it is assumed tt is the full set of instances within the
// monthly iteration, and only the instances matching the posisions of bySetPos
// are returned. This is an optimization.
func expandMonthByWeekdays(tt []time.Time, ib InvalidBehavior, bySetPos []int, weekdays ...QualifiedWeekday) []time.Time {
	if len(weekdays) == 0 {
		return tt
	}

	e := make([]time.Time, 0, len(tt))
	for _, t := range tt {
		e = append(e, weekdaysInMonth(t, weekdays, bySetPos, ib)...)
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
			res := weekdaysInYear(t, wd, ib)
			e = append(e, res...)
		}
	}

	sort.Slice(e, func(i, j int) bool {
		return e[i].Before(e[j])
	})

	return e

}
