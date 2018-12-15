package rrule

func ExampleParseRRule() {
	ParseRRule("FREQ=WEEKLY;BYDAY=1MO,2TU;COUNT=2")
}
