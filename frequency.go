package rrule

import "log"

// Frequency defines a set of constants for a base factor for how often recurrences happen.
type Frequency int

// String returns the RFC 5545 string for supported frequencies, and panics otherwise.
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
	log.Panicf("%d is not a supported frequency constant", f)
	return ""
}

// Frequencies specified in RFC 5545.
const (
	Secondly Frequency = iota
	Minutely
	Hourly
	Daily
	Weekly
	Monthly
	Yearly
)
