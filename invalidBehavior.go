package rrule

// InvalidBehavior specifies how to behave when a pattern generates a date that
// wouldn't exist, like February 31st.
type InvalidBehavior int

const (
	// OmitInvalid skips invalid dates. This is the only choice for RFC
	// 5545 and RFC 2445.
	OmitInvalid InvalidBehavior = iota

	// NextInvalid chooses the next valid date. So if February 31st were generated,
	// March 1st would be used.
	NextInvalid

	// PrevInvalid chooses the previously valid date. So if February 31st were generated,
	// the result would be February 28th (or 29th on a leap year).
	PrevInvalid
)
