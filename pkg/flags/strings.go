package flags

import (
	"github.com/skuid/skuid-cli/pkg/constants"
)

var (
	Host = &Flag[string]{
		Name:          "host",
		Usage:         `Host URL, e.g. [ https://my.skuidsite.com | my.skuidsite.com ]`,
		LegacyEnvVars: []string{constants.ENV_PLINY_HOST},
		Required:      true,
		Parse:         ParseHost,
	}

	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	Password = &Flag[string]{
		Name:          "password",
		Shorthand:     "p",
		Usage:         "Skuid NLX Password",
		LegacyEnvVars: []string{constants.ENV_SKUID_PW},
		Required:      true,
		Redact: func(val string) string {
			if val == "" {
				return ""
			} else {
				return "***"
			}
		},
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

	Since = &Flag[string]{
		Name:          "since",
		Shorthand:     "s",
		Usage:         "Timestamp or time span specifying only updated records to retrieve. Suggested timestamp format is: \"yyyy-MM-dd HH:mm AM\" or \"HH:mm AM\". Valid time spans look like various combination of \"1y2w3d8h30m\"",
		LegacyEnvVars: []string{constants.ENV_SKUID_RETRIEVE_SINCE},
		Parse:         ParseSince,
	}
)
