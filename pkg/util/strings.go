package util

import (
	"fmt"
	"slices"
	"time"

	"github.com/goschtalt/approx"
	"github.com/itlightning/dateparse"
)

type timeFormat struct {
	Format   string
	TimeOnly bool
}

var additionalTimeFormats = []timeFormat{
	{time.Layout, false},
	{time.Kitchen, true},
	{"15:04", true},
	{"3:04 PM", true},
	{"03:04 PM", true},
	{"03:04PM", true},
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

// GetTime tries to parse given string as a golang duration, then as a time and returns
// the time that it represents. In case of duration input, the returned timestamp is computed
// as the given reference time minus the amount of the duration.
// Formats Supported - see tests for examples:
//   - Durations: year, week and day in addition to standard golang durations.
//   - Times: any golang layout except for those that contain timezone abbreviations (e.g., UnixDate).
func GetTime(value string, reference time.Time, noFuture bool) (time.Time, error) {
	if value == "" || value == "0" {
		return time.Time{}, fmt.Errorf("value is not a valid time or duration: %q", value)
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
			return time.Time{}, err
		}
	}

	if noFuture && t.Compare(reference) > 0 {
		return time.Time{}, fmt.Errorf("value %q cannot represent a time in the future: %v", value, t.Format(time.RFC3339))
	}

	return t, nil
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
		// Per golang docs (https://pkg.go.dev/time#Parse), when parsing a time with a zone abbreviation, golang
		// will fabricate a location with the abbreviation and a zero offset when the abbreviation is not defined in the
		// location (e.g., value contains MST but parsing in America/Los_Angeles which contains PST/PDT).  If the timezone
		// is known, the value will be parsed accurately and there is no straightforward method for us to detect a tz
		// abbreviation was used (outside of doing string inspection).  Except for the case where a timezone abbreviation
		// in the value being parsed is "known" by the location in the reference value (e.g. value contains MST and reference
		// location contains MST zone), we are intentionally not supporting any formats that allow for timezone names.  However,
		// since we are using "loose" parsing to provide flexibility, dateparse will behave just like time.Parse/ParseInLocation
		// would for layouts like UnixDate so we'll end up with a fabricated location (a name but zero offset). Given this,
		// if there is a timezone name and the offset is zero, except for when its UTC which is valid to have zero for offset,
		// we treat it as a parsing error.
		// NOTE - Timezone abbreviations not known to reference location could be solved for and support for parsing implemented,
		// however its strongly discouraged due to amibguioty of abbreviations (they are not unique across the world for a given
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

	// dateparse does not handle certain golang formats (see also TimeOnly above)
	// and some formats that previous code supported
	for _, tf := range additionalTimeFormats {
		if t, err := time.ParseInLocation(tf.Format, value, reference.Location()); err == nil {
			if tf.TimeOnly {
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
