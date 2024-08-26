package flags

import "github.com/skuid/skuid-cli/pkg/constants"

var (
	Verbose = &Flag[bool, bool]{
		Name:      "verbose",
		Shorthand: "v",
		Usage:     "Show additional logging",
		Global:    true,
	}

	Trace = &Flag[bool, bool]{
		Name:    "trace",
		Usage:   "Show incredibly verbose logging",
		Default: false,
		Global:  true,
	}

	Debug = &Flag[bool, bool]{
		Name:   "debug",
		Usage:  "Show additional information in log messages",
		Global: true,
	}

	// Not used anywhere, so commenting out for now
	// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/149
	/*
		NoModule = &Flag[bool, bool]{
			Name:  "no-module",
			Usage: "Retrieve only those pages that do not have a module",
		}
	*/

	FileLogging = &Flag[bool, bool]{
		Name:   "file-logging",
		Usage:  "Log diagnostic information to files",
		Global: true,
	}

	Diagnostic = &Flag[bool, bool]{
		Name:          "diagnostic",
		Usage:         "Log additional fields with each log statement. Used for diagnostic purposes.",
		LegacyEnvVars: []string{constants.ENV_SKUID_FIELD_LOGGING},
		Global:        true,
	}

	IgnoreSkuidDb = &Flag[bool, bool]{
		Name:          "ignore-skuid-db",
		Shorthand:     "i",
		Usage:         "Force deployment by ignoring skuid db errors",
		LegacyEnvVars: []string{constants.ENV_SKUID_IGNORE_SKUIDDB},
	}

	NoConsoleLogging = &Flag[bool, bool]{
		Name:   "no-console-logging",
		Usage:  "Disable console logging",
		Global: true,
	}

	// TODO: This can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	SkipDataSources = &Flag[bool, bool]{
		Name:  "skip-data-sources",
		Usage: "Skip all DataSources",
	}
)
