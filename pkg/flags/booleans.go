package flags

import "github.com/skuid/tides/pkg/constants"

var (
	Verbose = &Flag[bool]{
		argument:   &argVerbose,
		Name:       "verbose",
		Shorthand:  "v",
		Usage:      "Show additional logging",
		EnvVarName: constants.ENV_VERBOSE,
		Global:     true,
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
)
