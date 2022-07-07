package flags

import "github.com/skuid/tides/pkg/constants"

var (
	Verbose = &Flag[bool]{
		Name:        "verbose",
		Shorthand:   "v",
		Usage:       "Show additional logging",
		EnvVarNames: []string{constants.ENV_VERBOSE},
		Global:      true,
	}

	Trace = &Flag[bool]{
		Name:    "trace",
		Usage:   "Show incredibly verbose logging",
		Default: false,
		Global:  true,
	}

	NoModule = &Flag[bool]{
		Name:  "no-module",
		Usage: "Retrieve only those pages that do not have a module",
	}

	FileLogging = &Flag[bool]{
		Name:        "file-logging",
		Usage:       "Log diagnostic information to files",
		EnvVarNames: []string{constants.ENV_TIDES_FILE_LOGGING},
		Global:      true,
	}

	FieldLogging = &Flag[bool]{
		Name:        "field-logging",
		Usage:       "Log additioal fields with each log statement",
		EnvVarNames: []string{constants.ENV_TIDES_FIELD_LOGGING},
		Global:      true,
	}
)
