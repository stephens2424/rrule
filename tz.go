package rrule

import "time"

// LoadLocation defaults to the standard library's implementation,
// but that implementation does not work on every platform. Set this
// to an alternative implementation when necessary.
var LoadLocation = time.LoadLocation
