package rrule

import (
	"errors"
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

	return t, err
}
