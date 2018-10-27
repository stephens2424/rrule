package rrule

type InvalidBehavior int

const (
	OmitInvalid InvalidBehavior = iota
	NextInvalid
	PrevInvalid
)
