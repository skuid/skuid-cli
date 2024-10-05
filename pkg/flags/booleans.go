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

	// Equivalent of a "verbose" mode that will include any logging.Fields in log messages. Should
	// not be confused with logging level which is controlled via --log-level.  The --verbose
	// flag existed in v0.6.7 and has been deprecated, however for backwards compat, it could not
	// be re-used so --diag was added and chosen in favor of a --debug to avoid confusion with
	// the logging level.
	Diag = &Flag[bool, bool]{
		Name:   "diag",
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
		LegacyEnvVars: []string{constants.EnvSkuidFieldLogging},
		Global:        true,
	}

	IgnoreSkuidDb = &Flag[bool, bool]{
		Name:          "ignore-skuid-db",
		Shorthand:     "i",
		Usage:         "Force deployment by ignoring skuid db errors",
		LegacyEnvVars: []string{constants.EnvSkuidIgnoreSkuidDb},
	}

	NoConsoleLogging = &Flag[bool, bool]{
		Name:   "no-console-logging",
		Usage:  "Disable console logging",
		Global: true,
	}

	NoClean = &Flag[bool, bool]{
		Name:  "no-clean",
		Usage: "Do not clean the target directory",
	}

	// TODO: This can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	SkipDataSources = &Flag[bool, bool]{
		Name:  "skip-data-sources",
		Usage: "Skip all DataSources",
	}
)
