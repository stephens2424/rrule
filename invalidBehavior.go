package rrule

type invalidBehavior int

const (
	omitInvalid invalidBehavior = iota
	nextInvalid
	prevInvalid
)
