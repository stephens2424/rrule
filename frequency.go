package rrule

// Frequency defines a set of constants for a base factor for how often recurrences happen.
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
