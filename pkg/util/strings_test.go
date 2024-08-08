package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skuid/skuid-cli/pkg/util"
)

func TestStringSliceContainsAnyKey(t *testing.T) {
	for _, tc := range []struct {
		description string
		array       []string
		keys        []string
		expected    bool
	}{
		{
			description: "found",
			array:       []string{"a", "b", "c"},
			keys:        []string{"b"},
			expected:    true,
		},
		{
			description: "found",
			array:       []string{"a", "b", "c"},
			keys:        []string{"z", "b"},
			expected:    true,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			keys:        []string{"d"},
			expected:    false,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			keys:        []string{"z", "f"},
			expected:    false,
		},
		{
			description: "not found",
			keys:        []string{"b"},
			expected:    false,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			expected:    false,
		},
		{
			description: "not found",
			expected:    false,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			result := util.StringSliceContainsAnyKey(tc.array, tc.keys)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetTime(t *testing.T) {
	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	require.NoError(t, err, "unable to load timezone location information for %q", locName)
	reference := time.Date(2006, 1, 2, 15, 4, 5, 0, loc)

	// inSameClock returns a copy of t in the provided location. The copy
	// will have the same clock reading as t and will represent a different
	// time instant depending on the provided location.
	// If t contains a monotonic clock reading this will also be stripped.
	inSameClock := func(t time.Time, loc *time.Location) time.Time {
		y, m, d := t.Date()
		return time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	}

	expectedTimeSecs := reference
	expectedTimeNoSecs := reference.Add(-5 * time.Second)
	expectedTimeNanoSecs := reference.Add(999999999 * time.Nanosecond)
	expectedTimeUTC := inSameClock(reference, time.UTC)
	expectedTimeNanoSecsUTC := inSameClock(expectedTimeNanoSecs, time.UTC)
	expectedTimeMidnight := time.Date(reference.Year(), reference.Month(), reference.Day(), 0, 0, 0, 0, reference.Location())
	expectedTimeMorning := reference.Add(-12 * time.Hour)
	// for times containing timezone offset that doesn't match -0800 which would be the correct
	// offset for America/Los_Angeles on 2006.01.02 @ 15:04
	expectedTimeSecsMinusHour := expectedTimeSecs.Add(-1 * time.Hour)
	expectedTimeNoSecsMinusHour := expectedTimeNoSecs.Add(-1 * time.Hour)
	expectedTimeNanoSecsMinusHour := expectedTimeNanoSecs.Add(-1 * time.Hour)

	testCases := []struct {
		testDescription string
		giveValue       string
		giveNoFuture    bool
		wantValue       time.Time
		wantError       bool
	}{
		// Times - Golang Layouts
		{"time.Layout", "01/02 03:04:05PM '06 -0700", false, expectedTimeSecsMinusHour, false},
		{"time.ANSIC", "Mon Jan 2 15:04:05 2006", false, expectedTimeSecs, false},
		{"time.UnixDate", "Mon Jan 2 15:04:05 MST 2006", false, time.Time{}, true},
		{"time.RubyDate", "Mon Jan 02 15:04:05 -0700 2006", false, expectedTimeSecsMinusHour, false},
		{"time.RFC822", "02 Jan 06 15:04 MST", false, time.Time{}, true},
		{"time.RFC822Z", "02 Jan 06 15:04 -0700", false, expectedTimeNoSecsMinusHour, false},
		{"time.RFC822Z-nofuture", reference.Add(1 * time.Minute).Format(time.RFC822Z), true, time.Time{}, true},
		{"time.RFC850", "Monday, 02-Jan-06 15:04:05 MST", false, time.Time{}, true},
		{"time.RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", false, time.Time{}, true},
		{"time.RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700", false, expectedTimeSecsMinusHour, false},
		{"time.RFC3339", "2006-01-02T15:04:05-07:00", false, expectedTimeSecsMinusHour, false},
		{"time.RFC3339-UTC", "2006-01-02T15:04:05Z", false, expectedTimeUTC, false},
		{"time.RFC3339-UTC-spaces", "20 06-01-0 2T15:04:05Z", false, time.Time{}, true},
		{"time.RFC3339Nano", "2006-01-02T15:04:05.999999999-07:00", false, expectedTimeNanoSecsMinusHour, false},
		{"time.RFC3339Nano-UTC", "2006-01-02T15:04:05.999999999Z", false, expectedTimeNanoSecsUTC, false},
		{"time.RFC3339Nano-nofuture", reference.Add(1 * time.Nanosecond).Format(time.RFC3339Nano), true, time.Time{}, true},
		{"time.Kitchen", "3:04PM", false, expectedTimeNoSecs, false},
		{"time.Stamp", "Jan 2 15:04:05", false, expectedTimeSecs, false},
		{"time.StampMilli", "Jan 2 15:04:05.000", false, expectedTimeSecs, false},
		{"time.StampNano", "Jan 2 15:04:05.000000000", false, expectedTimeSecs, false},
		{"time.DateTime", "2006-01-02 15:04:05", false, expectedTimeSecs, false},
		{"time.DateOnly", "2006-01-02", false, expectedTimeMidnight, false},
		{"time.DateOnly-space", "20 06-01-02", false, time.Time{}, true},
		{"time.DateOnly-spaces", "2006- 01- 02", false, time.Time{}, true},
		{"time.DateOnly-nofuture", reference.Add(48 * time.Hour).Format(time.DateOnly), true, time.Time{}, true},
		{"time.TimeOnly-afternoon", "15:04:05", false, expectedTimeSecs, false},
		{"time.TimeOnly-morning", "03:04:05", false, expectedTimeMorning, false},
		{"time.TimeOnly-midnight", "00:00:00", false, expectedTimeMidnight, false},

		// // Times - Non-Golang layouts
		{"2006/01/02 15:04:05", "2006/01/02 15:04:05", false, expectedTimeSecs, false},
		{"2006-01-02 03:04:05 PM", "2006-01-02 03:04:05 PM", false, expectedTimeSecs, false},
		{"2006/01/02 03:04:05 PM", "2006/01/02 03:04:05 PM", false, expectedTimeSecs, false},
		{"2006-01-02 03:04:05PM", "2006-01-02 03:04:05PM", false, expectedTimeSecs, false},
		{"2006/01/02 03:04:05PM", "2006/01/02 03:04:05PM", false, expectedTimeSecs, false},
		{"2006-01-02 15:04", "2006-01-02 15:04", false, expectedTimeNoSecs, false},
		{"2006/01/02 15:04", "2006/01/02 15:04", false, expectedTimeNoSecs, false},
		{"2006-01-02 03:04 PM", "2006-01-02 03:04 PM", false, expectedTimeNoSecs, false},
		{"2006/01/02 03:04 PM", "2006/01/02 03:04 PM", false, expectedTimeNoSecs, false},
		{"2006-01-02 03:04PM", "2006-01-02 03:04PM", false, expectedTimeNoSecs, false},
		{"2006/01/02 03:04PM", "2006/01/02 03:04PM", false, expectedTimeNoSecs, false},
		{"2006/01/02", "2006/01/02", false, expectedTimeMidnight, false},
		{"15:04", "15:04", false, expectedTimeNoSecs, false},
		{"3:04 PM", "3:04 PM", false, expectedTimeNoSecs, false},
		{"03:04 PM", "03:04 PM", false, expectedTimeNoSecs, false},
		{"03:04PM", "03:04PM", false, expectedTimeNoSecs, false},

		// Times - Golang layouts with timezone abbreviation "known" to the location
		{"time.UnixDate-PST", "Mon Jan 2 15:04:05 PST 2006", false, expectedTimeSecs, false},
		{"time.UnixDate-PDT", "Mon Jan 2 15:04:05 PDT 2006", false, expectedTimeSecsMinusHour, false},
		{"time.RFC822-PST", "02 Jan 06 15:04 PST", false, expectedTimeNoSecs, false},
		{"time.RFC822-PDT", "02 Jan 06 15:04 PDT", false, expectedTimeNoSecsMinusHour, false},
		{"time.RFC850-PST", "Monday, 02-Jan-06 15:04:05 PST", false, expectedTimeSecs, false},
		{"time.RFC850-PDT", "Monday, 02-Jan-06 15:04:05 PDT", false, expectedTimeSecsMinusHour, false},
		{"time.RFC1123-PST", "Mon, 02 Jan 2006 15:04:05 PST", false, expectedTimeSecs, false},
		{"time.RFC1123-PDT", "Mon, 02 Jan 2006 15:04:05 PDT", false, expectedTimeSecsMinusHour, false},

		// Durations
		{"1ns", "1ns", false, reference.Add(-1 * time.Nanosecond), false},
		{"1.1µs", "1.1µs", false, reference.Add(-1100 * time.Nanosecond), false},
		{"1.us", "1.1us", false, reference.Add(-1100 * time.Nanosecond), false},
		{"2.2ms", "2.2ms", false, reference.Add(-2200 * time.Microsecond), false},
		{"0s", "0s", false, reference.Add(0), false},
		{"1s", "1s", false, reference.Add(-1 * time.Second), false},
		{"3.3s", "3.3s", false, reference.Add(-3300 * time.Millisecond), false},
		{"-5.5s", "-5.5s", false, reference.Add(5500 * time.Millisecond), false},
		{"-5.5s-nofuture", "-5.5s", true, time.Time{}, true},
		{"1m", "1m", false, reference.Add(-1 * time.Minute), false},
		{"1.5h", "1.5h", false, reference.Add(-90 * time.Minute), false},
		{"1.012h", "1.012h", false, reference.Add(-1 * (time.Hour + 43*time.Second + 200*time.Millisecond)), false},
		{"1h30m", "1h30m", false, reference.Add(-90 * time.Minute), false},
		{"1d", "1d", false, reference.Add(-1 * 24 * time.Hour), false},
		{"1w", "1w", false, reference.Add(-7 * 24 * time.Hour), false},
		{"1y", "1y", false, reference.Add(-365 * 24 * time.Hour), false},
		{"1y2w3d4h5m6s", "1y2w3d4h5m6s", false, reference.Add(-1 * ((365 * 24 * time.Hour) + (2 * 7 * 24 * time.Hour) + (3 * 24 * time.Hour) + (4 * time.Hour) + (5 * time.Minute) + (6 * time.Second))), false},
		{"4y0w0d0h0m0s", "4y0w0d0h0m0s", false, reference.Add(-1 * (4 * 365 * 24 * time.Hour)), false},
		{"0y0w0d0h0m0s", "0y0w0d0h0m0s", false, reference.Add(0), false},
		{"-3y", "-3y", false, reference.Add((3 * 365 * 24 * time.Hour)), false},
		{"-3y-nofuture", "-3y", true, time.Time{}, true},

		{"invalid", "invalid", false, time.Time{}, true},
		{"empty", "", false, time.Time{}, true},
		{"0", "0", false, time.Time{}, true},
		{"3days", "3days", false, time.Time{}, true},
		{"3 days", "3 days", false, time.Time{}, true},
		{"3day", "3day", false, time.Time{}, true},
		{"3months", "3months", false, time.Time{}, true},
		{"3 months", "3 months", false, time.Time{}, true},
		{"3month", "3month", false, time.Time{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := util.GetTime(tc.giveValue, reference, tc.giveNoFuture)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, tc.wantValue.Equal(actualValue))
		})
	}
}
