package flags

import (
	"github.com/skuid/skuid-cli/pkg/constants"
)

var (
	Host = &Flag[CustomString]{
		Name:          "host",
		Usage:         `Host URL, e.g. [ https://my.skuidsite.com | my.skuidsite.com ]`,
		LegacyEnvVars: []string{constants.ENV_PLINY_HOST},
		Required:      true,
		Parse:         ParseHost,
	}

	Password = &Flag[RedactedString]{
		Name:          "password",
		Shorthand:     "p",
		Usage:         "Skuid NLX Password",
		LegacyEnvVars: []string{constants.ENV_SKUID_PW},
		Required:      true,
	}

	Username = &Flag[string]{
		Name:          "username",
		Shorthand:     "u",
		LegacyEnvVars: []string{constants.ENV_SKUID_UN},
		Required:      true,
		Usage:         "Skuid NLX Username",
	}

	App = &Flag[string]{
		Name:          "app",
		Shorthand:     "a",
		LegacyEnvVars: []string{constants.ENV_SKUID_APP_NAME},
		Usage:         "Scope the operation to a specific Skuid NLX App Name",
	}

	Dir = &Flag[string]{
		Name:          "dir",
		Shorthand:     "d",
		Usage:         "Target directory for this operation",
		LegacyEnvVars: []string{constants.ENV_SKUID_DEFAULT_FOLDER},
	}

	LogDirectory = &Flag[string]{
		Name:          "log-directory",
		Shorthand:     "l",
		Usage:         "Target directory for file logging",
		Default:       ".logs",
		LegacyEnvVars: []string{constants.ENV_SKUID_LOGGING_LOCATION},
		Global:        true,
	}

	Since = &Flag[CustomString]{
		Name:          "since",
		Shorthand:     "s",
		Usage:         "Timestamp or time span specifying only updated records to retrieve. Suggested timestamp format is: \"yyyy-MM-dd HH:mm AM\" or \"HH:mm AM\". Valid timespans look like various combination of \"1y2M3d8h30m\" or \"3 days\"",
		LegacyEnvVars: []string{constants.ENV_SKUID_RETRIEVE_SINCE},
		Parse:         ParseSince,
	}
)
