package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skuid/skuid-cli/pkg/util"
)

func TestStringSliceContainsKey(t *testing.T) {
	for _, tc := range []struct {
		description string
		array       []string
		key         string
		expected    bool
	}{
		{
			description: "found",
			array:       []string{"a", "b", "c"},
			key:         "b",
			expected:    true,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			key:         "d",
			expected:    false,
		},
		{
			description: "not found",
			key:         "b",
			expected:    false,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			key:         "",
			expected:    false,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			result := util.StringSliceContainsKey(tc.array, tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

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

func TestGetTimestamp(t *testing.T) {
	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	require.NoError(t, err, "unable to load timezone location information for %q", locName)
	reference := time.Date(2006, 1, 2, 15, 4, 5, 0, loc)

	addDuration := func(d time.Duration) string {
		return util.FormatTimestamp(reference.Add(d))
	}

	testCases := []struct {
		testDescription string
		giveValue       string
		giveNoFuture    bool
		wantValue       string
		wantError       bool
	}{
		// Times - Golang Layouts
		{"time.Layout", "01/02 03:04:05PM '06 -0700", false, "1136239445.000000000", false},
		{"time.ANSIC", "Mon Jan 2 15:04:05 2006", false, "1136243045.000000000", false},
		{"time.UnixDate", "Mon Jan 2 15:04:05 MST 2006", false, "", true},
		{"time.RubyDate", "Mon Jan 02 15:04:05 -0700 2006", false, "1136239445.000000000", false},
		{"time.RFC822", "02 Jan 06 15:04 MST", false, "", true},
		{"time.RFC822Z", "02 Jan 06 15:04 -0700", false, "1136239440.000000000", false},
		{"time.RFC822Z-nofuture", reference.Add(1 * time.Minute).Format(time.RFC822Z), true, "", true},
		{"time.RFC850", "Monday, 02-Jan-06 15:04:05 MST", false, "", true},
		{"time.RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", false, "", true},
		{"time.RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700", false, "1136239445.000000000", false},
		{"time.RFC3339", "2006-01-02T15:04:05-07:00", false, "1136239445.000000000", false},
		{"time.RFC3339-UTC", "2006-01-02T15:04:05Z", false, "1136214245.000000000", false},
		{"time.RFC3339-UTC-spaces", "20 06-01-0 2T15:04:05Z", false, "", true},
		{"time.RFC3339Nano", "2006-01-02T15:04:05.999999999-07:00", false, "1136239445.999999999", false},
		{"time.RFC3339Nano-UTC", "2006-01-02T15:04:05.999999999Z", false, "1136214245.999999999", false},
		{"time.RFC3339Nano-nofuture", reference.Add(1 * time.Nanosecond).Format(time.RFC3339Nano), true, "", true},
		{"time.Kitchen", "3:04PM", false, "1136243040.000000000", false},
		{"time.Stamp", "Jan 2 15:04:05", false, "1136243045.000000000", false},
		{"time.StampMilli", "Jan 2 15:04:05.000", false, "1136243045.000000000", false},
		{"time.StampNano", "Jan 2 15:04:05.000000000", false, "1136243045.000000000", false},
		{"time.DateTime", "2006-01-02 15:04:05", false, "1136243045.000000000", false},
		{"time.DateOnly", "2006-01-02", false, "1136188800.000000000", false},
		{"time.DateOnly-space", "20 06-01-02", false, "", true},
		{"time.DateOnly-spaces", "2006- 01- 02", false, "", true},
		{"time.DateOnly-nofuture", reference.Add(48 * time.Hour).Format(time.DateOnly), true, "", true},
		{"time.TimeOnly-afternoon", "15:04:05", false, "1136243045.000000000", false},
		{"time.TimeOnly-morning", "03:04:05", false, "1136199845.000000000", false},
		{"time.TimeOnly-midnight", "00:00:00", false, "1136188800.000000000", false},

		// Times - Non-Golang layouts
		{"2006/01/02 15:04:05", "2006/01/02 15:04:05", false, "1136243045.000000000", false},
		{"2006-01-02 03:04:05 PM", "2006-01-02 03:04:05 PM", false, "1136243045.000000000", false},
		{"2006/01/02 03:04:05 PM", "2006/01/02 03:04:05 PM", false, "1136243045.000000000", false},
		{"2006-01-02 03:04:05PM", "2006-01-02 03:04:05PM", false, "1136243045.000000000", false},
		{"2006/01/02 03:04:05PM", "2006/01/02 03:04:05PM", false, "1136243045.000000000", false},
		{"2006-01-02 15:04", "2006-01-02 15:04", false, "1136243040.000000000", false},
		{"2006/01/02 15:04", "2006/01/02 15:04", false, "1136243040.000000000", false},
		{"2006-01-02 03:04 PM", "2006-01-02 03:04 PM", false, "1136243040.000000000", false},
		{"2006/01/02 03:04 PM", "2006/01/02 03:04 PM", false, "1136243040.000000000", false},
		{"2006-01-02 03:04PM", "2006-01-02 03:04PM", false, "1136243040.000000000", false},
		{"2006/01/02 03:04PM", "2006/01/02 03:04PM", false, "1136243040.000000000", false},
		{"2006/01/02", "2006/01/02", false, "1136188800.000000000", false},
		{"15:04", "15:04", false, "1136243040.000000000", false},
		{"3:04 PM", "3:04 PM", false, "1136243040.000000000", false},
		{"03:04 PM", "03:04 PM", false, "1136243040.000000000", false},
		{"03:04PM", "03:04PM", false, "1136243040.000000000", false},

		// Times - Golang layouts with timezone abbreviation "known" to the location
		{"time.UnixDate-PST", "Mon Jan 2 15:04:05 PST 2006", false, "1136243045.000000000", false},
		{"time.UnixDate-PDT", "Mon Jan 2 15:04:05 PDT 2006", false, "1136239445.000000000", false},
		{"time.RFC822-PST", "02 Jan 06 15:04 PST", false, "1136243040.000000000", false},
		{"time.RFC822-PDT", "02 Jan 06 15:04 PDT", false, "1136239440.000000000", false},
		{"time.RFC850-PST", "Monday, 02-Jan-06 15:04:05 PST", false, "1136243045.000000000", false},
		{"time.RFC850-PDT", "Monday, 02-Jan-06 15:04:05 PDT", false, "1136239445.000000000", false},
		{"time.RFC1123-PST", "Mon, 02 Jan 2006 15:04:05 PST", false, "1136243045.000000000", false},
		{"time.RFC1123-PDT", "Mon, 02 Jan 2006 15:04:05 PDT", false, "1136239445.000000000", false},

		// Durations
		{"1ns", "1ns", false, addDuration(-1 * time.Nanosecond), false},
		{"1.1µs", "1.1µs", false, addDuration(-1100 * time.Nanosecond), false},
		{"1.us", "1.1us", false, addDuration(-1100 * time.Nanosecond), false},
		{"2.2ms", "2.2ms", false, addDuration(-2200 * time.Microsecond), false},
		{"0s", "0s", false, addDuration(0), false},
		{"1s", "1s", false, addDuration(-1 * time.Second), false},
		{"3.3s", "3.3s", false, addDuration(-3300 * time.Millisecond), false},
		{"-5.5s", "-5.5s", false, addDuration(5500 * time.Millisecond), false},
		{"-5.5s-nofuture", "-5.5s", true, "", true},
		{"1m", "1m", false, addDuration(-1 * time.Minute), false},
		{"1.5h", "1.5h", false, addDuration(-90 * time.Minute), false},
		{"1.012h", "1.012h", false, addDuration(-1 * (time.Hour + 43*time.Second + 200*time.Millisecond)), false},
		{"1h30m", "1h30m", false, addDuration(-90 * time.Minute), false},
		{"1d", "1d", false, addDuration(-1 * 24 * time.Hour), false},
		{"1w", "1w", false, addDuration(-7 * 24 * time.Hour), false},
		{"1y", "1y", false, addDuration(-365 * 24 * time.Hour), false},
		{"1y2w3d4h5m6s", "1y2w3d4h5m6s", false, addDuration(-1 * ((365 * 24 * time.Hour) + (2 * 7 * 24 * time.Hour) + (3 * 24 * time.Hour) + (4 * time.Hour) + (5 * time.Minute) + (6 * time.Second))), false},
		{"4y0w0d0h0m0s", "4y0w0d0h0m0s", false, addDuration(-1 * (4 * 365 * 24 * time.Hour)), false},
		{"0y0w0d0h0m0s", "0y0w0d0h0m0s", false, addDuration(0), false},
		{"-3y", "-3y", false, addDuration((3 * 365 * 24 * time.Hour)), false},
		{"-3y-nofuture", "-3y", true, "", true},

		{"invalid", "invalid", false, "", true},
		{"empty", "", false, "", true},
		{"0", "0", false, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := util.GetTimestamp(tc.giveValue, reference, tc.giveNoFuture)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantValue, actualValue)
		})
	}
}

func TestParseTimestamps(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveValue       string
		giveDefault     int64
		wantSeconds     int64
		wantNano        int64
		wantError       bool
	}{
		// unix timestamps
		{"1136073600", "1136073600", 0, 1136073600, 0, false},
		{"1136073600.000000001", "1136073600.000000001", 0, 1136073600, 1, false},
		{"1136073600.0000000010", "1136073600.0000000010", 0, 1136073600, 1, false},
		{"1136073600.0000000001", "1136073600.0000000001", 0, 1136073600, 0, false},
		{"1136073600.0000000009", "1136073600.0000000009", 0, 1136073600, 0, false},
		{"1136073600.00000001", "1136073600.00000001", 0, 1136073600, 10, false},
		{"foo.bar", "foo.bar", 0, 0, 0, true},
		{"space", "11360 73600.00000001", 0, 0, 0, true},
		{"spaces", "11360 73600 . 000 00001", 0, 0, 0, true},
		{"whitespace", "11360	73600.000	00001", 0, 0, 0, true},
		{"1136073600.bar", "1136073600.bar", 0, 1136073600, 0, true},
		{"empty", "", -1, -1, 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualSeconds, actualNano, err := util.ParseTimestamp(tc.giveValue, tc.giveDefault)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantSeconds, actualSeconds)
			assert.Equal(t, tc.wantNano, actualNano)
		})
	}
}
