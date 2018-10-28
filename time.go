package rrule

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	rfc5545_WithOffset    = "20060102T150405Z0700"
	rfc5545_WithoutOffset = "20060102T150405"
)

func parseTime(str string, defaultLoc *time.Location) (time.Time, error) {
	//        DTSTART;TZID=America/New_York:19970902T090000

	var t time.Time

	if defaultLoc == nil {
		defaultLoc = time.UTC
	}
	loc := defaultLoc

	if idBeg := strings.Index(str, ";TZID="); idBeg >= 0 {
		locBeg := idBeg + 6
		locEnd := locBeg + strings.Index(str[locBeg:], ":")
		if locEnd < 0 {
			return t, errors.New("no end to TZID")
		}

		var err error
		loc, err = LoadLocation(str[locBeg:locEnd])
		if err != nil {
			return t, err
		}

		str = str[locEnd+1:]
	} else {
		colonIdx := strings.IndexAny(str, ":=")
		str = str[colonIdx+1:]
	}

	t, err := time.ParseInLocation(rfc5545_WithOffset, str, loc)
	if err != nil {
		t, err = time.ParseInLocation(rfc5545_WithoutOffset, str, loc)
	}

	// From RFC 5545:
	//
	//     If, based on the definition of the referenced time zone, the local
	//     time described occurs more than once (when changing from daylight
	//     to standard time), the DATE-TIME value refers to the first
	//     occurrence of the referenced time.  Thus, TZID=America/
	//     New_York:20071104T013000 indicates November 4, 2007 at 1:30 A.M.
	//     EDT (UTC-04:00).  If the local time described does not occur (when
	//     changing from standard to daylight time), the DATE-TIME value is
	//     interpreted using the UTC offset before the gap in local times.
	//     Thus, TZID=America/New_York:20070311T023000 indicates March 11,
	//     2007 at 3:30 A.M. EDT (UTC-04:00), one hour after 1:30 A.M. EST
	//     (UTC-05:00).
	//
	// However, Go's time.ParseInLocation makes no guarantee about how it
	// behaves relative to "fall-back" repetition of an hour in DST
	// transitions. (time.Date explicitly documents the same concept is
	// undefined.) Therefore, here, we normalize according to the spec by
	// trying to remove an hour and see if the local time is the same, and
	// if so, we keep that difference. Otherwise, if the original string was
	// in the 2am range, but the parsed time is less than 2 o'clock, advance an hour.
	if tMinusHour := t.Add(-1 * time.Hour); t.Hour() == tMinusHour.Hour() {
		t = tMinusHour
	} else if twoAMRegex.MatchString(str) {
		t = t.Add(1 * time.Hour)
	}

	return t, err
}

var twoAMRegex = regexp.MustCompile("T02[0-9]{4}(Z|[0-9]{4})?$")
