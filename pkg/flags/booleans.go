package flags

import "github.com/skuid/tides/pkg/constants"

var (
	Verbose = &Flag[bool]{
		argument:    &argVerbose,
		Name:        "verbose",
		Shorthand:   "v",
		Usage:       "Show additional logging",
		EnvVarNames: []string{constants.ENV_VERBOSE},
		Global:      true,
	}

	NoZip = &Flag[bool]{
		argument:  &argNoZip,
		Name:      "no-zip",
		Shorthand: "x",
		Usage:     "Ask for site not to be zipped",
	}

	NoModule = &Flag[bool]{
		argument: &argNoModule,
		Name:     "no-module",
		Usage:    "Retrieve only those pages that do not have a module",
	}

	FileLogging = &Flag[bool]{
		argument:    &argFileLogging,
		Name:        "file-logging",
		Usage:       "Log diagnostic information to files",
		EnvVarNames: []string{constants.ENV_TIDES_FILE_LOGGING},
		Default:     true,
		Global:      true,
	}
)
