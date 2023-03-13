package constants

import "time"

var (
	TimeFormatStrings = []string{
		//time.Layout,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		time.DateTime,
		time.DateOnly,
		time.TimeOnly,
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
