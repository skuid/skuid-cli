package constants

import "time"

var (
	DefaultTimeFormat = "2006-01-02 03:04 PM"
	// TimeFormatStrings defines the format strings we will use for time.Parse
	// Note we have excluded some from `time` that are 1) intended for network timestamps, which we do not expect users
	// to want, or 2) those that contain time zone abbreviations; go currently disregards them, even with time/tzdata.
	// Anything other than the "time.Local" location (e.g. "EST") ends up assumed to be UTC, it's a bit aggravating.
	TimeFormatStrings = []string{
		time.ANSIC,    // "Mon Jan _2 15:04:05 2006"
		time.RubyDate, // "Mon Jan 02 15:04:05 -0700 2006"
		time.RFC822Z,  // "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
		time.Kitchen,  // "3:04PM"
		time.Stamp,    // "Jan _2 15:04:05"
		time.DateTime, // "2006-01-02 15:04:05"
		"2006/01/02 15:04:05",
		"2006-01-02 03:04:05 PM",
		"2006/01/02 03:04:05 PM",
		"2006-01-02 03:04:05PM",
		"2006/01/02 03:04:05PM",
		"2006-01-02 15:04",
		"2006/01/02 15:04",
		"2006-01-02 03:04 PM",
		"2006/01/02 03:04 PM",
		"2006-01-02 03:04PM",
		"2006/01/02 03:04PM",
		time.DateOnly, // "2006-01-02"
		"2006/01/02",
		time.TimeOnly, // "15:04:05"
		"15:04",
		"3:04 PM",
		"03:04 PM",
		"03:04PM",
		//time.Layout,      // "01/02 03:04:05PM '06 -0700"
		//time.UnixDate,    // "Mon Jan _2 15:04:05 MST 2006"
		//time.RFC822,      // "02 Jan 06 15:04 MST"
		//time.RFC850,      // "Monday, 02-Jan-06 15:04:05 MST"
		//time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
		//time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
		//time.StampMilli,  // "Jan _2 15:04:05.000"
		//time.StampMicro,  // "Jan _2 15:04:05.000000"
		//time.StampNano,   // "Jan _2 15:04:05.000000000"
	}
	TimeUnits = map[string][]string{
		"s": {
			"s",
			"ss",
			"sec",
			"secs",
			"second",
			"seconds",
		},
		"m": {
			"m",
			"mm",
			"min",
			"mins",
			"minute",
			"minutes",
		},
		"h": {
			"h",
			"hh",
			"hr",
			"hrs",
			"hour",
			"hours",
		},
		"d": {
			"d",
			"day",
			"days",
		},
		"M": {
			"MM",
			"mo",
			"Mo",
			"mos",
			"Mos",
			"mon",
			"month",
			"months",
		},
		"y": {
			"yy",
			"yyyy",
			"yr",
			"yrs",
			"year",
			"years",
		},
	}
)
