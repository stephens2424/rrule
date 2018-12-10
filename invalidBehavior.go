package rrule

type invalidBehavior int

const (
	OmitInvalid invalidBehavior = iota
	NextInvalid
	PrevInvalid
)
