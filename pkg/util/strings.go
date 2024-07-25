package util

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goschtalt/approx"
	"github.com/itlightning/dateparse"
)

// StringSliceContainsKey returns true if a string is contained in a slice
func StringSliceContainsKey(strings []string, key string) bool {
	return slices.Contains(strings, key)
}

// same thing as above except we early exit at any given point going through the
// strings and keys
func StringSliceContainsAnyKey(strings []string, keys []string) bool {
	if len(strings) == 0 || len(keys) == 0 {
		return false
	}

	return slices.ContainsFunc(keys, func(key string) bool {
		return slices.Contains(strings, key)
	})
}

// GetTimestamp tries to parse given string as a golang duration,
// then as a time. If successful, it returns a Unix timestamp
// as string otherwise returns empty string with an error.
// In case of duration input, the returned timestamp is computed
// as the given reference time minus the amount of the duration.
// Formats Supported - see tests for examples:
//   - Durations: year, week and day in addition to standard golang durations.
//   - Times: any golang layout except for those that contain timezone abbreviations (e.g., UnixDate).
func GetTimestamp(value string, reference time.Time, noFuture bool) (string, error) {
	if value == "" || value == "0" {
		return "", fmt.Errorf("must provide a valid time or duration: %q", value)
	}

	// try duration first for two reasons:
	//   1. It's more likely to be specified over a time
	//   2. It's less likely to give a false positive "valid" format match over time
	var t time.Time
	if d, err := approx.ParseDuration(value); err == nil {
		t = reference.Add(-d)
	} else {
		t, err = parseTime(value, reference)
		if err != nil {
			return "", err
		}
	}

	if noFuture && t.Compare(reference) > 0 {
		return "", fmt.Errorf("%q cannot resolve to a time in the future: %v", value, t.Format(time.RFC3339))
	}

	return fmt.Sprintf("%d.%09d", t.Unix(), int64(t.Nanosecond())), nil
}

func ParseTimestamp(value string, defaultSeconds int64) (seconds int64, nanoseconds int64, err error) {
	if value == "" {
		return defaultSeconds, 0, nil
	}
	return parseTimestamp(value)
}

func parseTime(value string, reference time.Time) (time.Time, error) {
	// dateparse will treat certain times (depending on their hour/min/sec values) as dates
	// (d/m/y) so we avoid uncertainty by checking for timeonly first.  Not ideal
	// performance wise but more reliable outcome.
	if t, err := time.ParseInLocation(time.TimeOnly, value, reference.Location()); err == nil {
		// only time was provided so adjust to current date at time specified
		return forceCurrentDate(t, reference), nil
	}

	if t, err := dateparse.ParseIn(value, reference.Location()); err == nil {
		// golang will fabricate a timezone when any non-numeric timezone is specified in the value being parsed and
		// the timezone name isn't known by the current timezone (e.g., value contains MST but parsing in PST - see
		// https://pkg.go.dev/time#Parse for details). We are intentionally not supporting any formats that allow
		// for timezone names, however since we are using "loose" parsing to provide flexibility, dateparse will behave
		// just like time.Parse/ParseInLocation would for layouts like UnixDate so we'll end up with a fabricated
		// timezone (a name but zero offset). Given this, if there is a timezone name and the offset is zero, except
		// for when its UTC which is valid to have zero for offset, we treat it as a parsing error.
		// NOTE - This situation could be solved for and support for parsing timezone names implemented, however its
		// strongly discouraged due to amibguioty of abbreviations (they are not unique across the world for a given
		// offset).  If this ever needed to be solved for, the timezone abbreviation would need to be looked up in a list
		// (e.g., a package like go-timezone) along with the current country code to resolve to the correct timezone and
		// then apply the offset to the time.  Not recommended but possible understanding that it still wouldn't be 100%
		// reliable.
		tzName, tzOffset := t.Zone()
		if tzName != "" && tzName != "UTC" && tzOffset == 0 {
			return time.Time{}, fmt.Errorf("value contained an unexpected timezone identifier - please use numeric timezone values")
		}

		// certain formats (e.g. Stampli) permit no year
		if t.Year() == 0 {
			t = time.Date(reference.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
				t.Nanosecond(), reference.Location())
			return t, nil
		}
		return t, nil
	}

	// dateparse does not handle certain golang formats (see TimeOnly above)
	for _, l := range []string{time.Layout, time.Kitchen} {
		if t, err := time.ParseInLocation(l, value, reference.Location()); err == nil {
			if l == time.Kitchen {
				// only time was provided so adjust to current date at time specified
				return forceCurrentDate(t, reference), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse value as time or duration: %q", value)
}

func forceCurrentDate(t time.Time, reference time.Time) time.Time {
	return time.Date(reference.Year(), reference.Month(), reference.Day(), t.Hour(), t.Minute(),
		t.Second(), t.Nanosecond(), reference.Location())

}

func parseTimestamp(value string) (sec int64, nsec int64, err error) {
	s, n, ok := strings.Cut(value, ".")
	sec, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return sec, 0, err
	}
	if !ok {
		return sec, 0, nil
	}
	nsec, err = strconv.ParseInt(n, 10, 64)
	if err != nil {
		return sec, nsec, err
	}
	// should already be in nanoseconds but just in case convert n to nanoseconds
	nsec = int64(float64(nsec) * math.Pow(float64(10), float64(9-len(n))))
	return sec, nsec, nil
}
