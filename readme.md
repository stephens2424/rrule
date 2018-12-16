# RRule

[![GoDoc](https://godoc.org/github.com/stephens2424/rrule?status.svg)](https://godoc.org/github.com/stephens2424/rrule)

Package RRule implements recurrence processing according to RFC 5545. See the
[godoc](https://godoc.org/github.com/stephens2424/rrule) for usage information.

This implementation was written to overcome performance issues in previous
implementations. Those previous ones were generally implemented as direct
translations of the venerable python-dateutil, however the algorithms were
complicated and probably didn't use Go's language features effectively enough
for performance optimization. The observed problem was particularly acute under
GopherJS.

The library here is essentially complete. A fair number of various patterns are
tested, particularly simple ones. The library has not seen, at the time of this
writing, any production usage, however. Issue reports with implementation
accuracy or performance problems are particularly welcome.

Licensed under BSD-3. See the LICENSE file.
