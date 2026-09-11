// Package deadline turns a command-line argument into the instant a timer
// should fire.
//
// It deliberately does no I/O and never calls time.Now itself: the caller
// supplies "now", so every rule below can be tested against a fixed clock.
package deadline

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Layouts accepted for the absolute forms. Seconds are optional in both.
const (
	layoutDateTime      = "2006-01-02T15:04"
	layoutDateTimeSecs  = "2006-01-02T15:04:05"
	layoutTimeOfDay     = "15:04"
	layoutTimeOfDaySecs = "15:04:05"
)

// Parse resolves arg against now and returns the instant the timer fires.
//
// Accepted forms:
//
//	2h15m, 45m, 8m20s    a duration from now, as understood by time.ParseDuration
//	10                   a bare number is that many seconds
//	2027-03-14T09:00     an absolute local date and time
//	T07:30               a time of day today, rolling to tomorrow if it has passed
//
// The result is always strictly after now, so callers can subtract without
// having to defend against a negative countdown.
func Parse(arg string, now time.Time) (time.Time, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return time.Time{}, fmt.Errorf("no timer given")
	}

	switch {
	case arg[0] == 'T' || arg[0] == 't':
		return parseTimeOfDay(arg[1:], now)
	case strings.Contains(arg[1:], "-"):
		// Only the date forms contain an interior hyphen. A leading one is a
		// negative duration, which parseDuration rejects with a clearer message.
		return parseAbsolute(arg, now)
	default:
		return parseDuration(arg, now)
	}
}

// bareNumber matches an argument carrying no unit at all, like "10" or "1.5".
// Anything with a unit in it, even a partial "1h30", is left to ParseDuration.
var bareNumber = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// parseDuration handles the relative forms, which are the common case.
func parseDuration(arg string, now time.Time) (time.Time, error) {
	if bareNumber.MatchString(arg) {
		arg += "s" // seconds are the unit people leave off
	}

	d, err := time.ParseDuration(arg)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not a duration like 2h15m, 45m, 8m20s or 10", arg)
	}
	if d <= 0 {
		return time.Time{}, fmt.Errorf("%q is not a positive duration", arg)
	}
	return now.Add(d), nil
}

// parseAbsolute handles a full date and time, which must be in the future.
// Unlike a bare time of day there is no sensible way to roll it forward.
func parseAbsolute(arg string, now time.Time) (time.Time, error) {
	fire, err := parseFirst(arg, now.Location(), layoutDateTime, layoutDateTimeSecs)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not a date like 2027-03-14T09:00", arg)
	}
	if !fire.After(now) {
		return time.Time{}, fmt.Errorf("%s has already passed", fire.Format(layoutDateTime))
	}
	return fire, nil
}

// parseTimeOfDay handles the T07:30 form. arg has already had its T stripped.
func parseTimeOfDay(arg string, now time.Time) (time.Time, error) {
	clock, err := parseFirst(arg, now.Location(), layoutTimeOfDay, layoutTimeOfDaySecs)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not a time of day like T07:30", arg)
	}

	// A parsed clock reading carries year zero as its date, so graft the hours
	// and minutes onto today.
	fire := time.Date(now.Year(), now.Month(), now.Day(),
		clock.Hour(), clock.Minute(), clock.Second(), 0, now.Location())
	if !fire.After(now) {
		fire = fire.AddDate(0, 0, 1) // asking for T07:30 in the afternoon means tomorrow
	}
	return fire, nil
}

// parseFirst returns the result of the first layout that accepts arg.
func parseFirst(arg string, loc *time.Location, layouts ...string) (time.Time, error) {
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, arg, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("no layout matched %q", arg)
}
