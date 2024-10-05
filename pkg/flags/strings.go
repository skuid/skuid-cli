package flags

import (
	"github.com/skuid/skuid-cli/pkg/constants"
)

var (
	Host = &Flag[string, string]{
		Name:          "host",
		Usage:         "Skuid NLX Host, e.g., https://my.skuidsite.com, my.skuidsite.com",
		LegacyEnvVars: []string{constants.EnvPlinyHost},
		Required:      true,
		Parse:         ParseHost,
	}

	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	Password = &Flag[string, string]{
		Name:          "password",
		Shorthand:     "p",
		Usage:         "Skuid NLX Password",
		LegacyEnvVars: []string{constants.EnvSkuidPw},
		Required:      true,
		Parse:         ParseString(ParseStringOptions{AllowBlank: false}),
		Redact: func(val string) string {
			if val == "" {
				return ""
			} else {
				return "***"
			}
		},
	}

	Username = &Flag[string, string]{
		Name:          "username",
		Shorthand:     "u",
		LegacyEnvVars: []string{constants.EnvSkuidUn},
		Required:      true,
		Usage:         "Skuid NLX Username",
		Parse:         ParseString(ParseStringOptions{AllowBlank: false}),
	}

	App = &Flag[string, string]{
		Name:          "app",
		Shorthand:     "a",
		LegacyEnvVars: []string{constants.EnvSkuidAppName},
		Usage:         "Scope the operation to a specific Skuid NLX App Name",
	}

	Dir = &Flag[string, string]{
		Name:          "dir",
		Shorthand:     "d",
		Usage:         "Target directory for this operation",
		LegacyEnvVars: []string{constants.EnvSkuidDefaultFolder},
		Parse:         ParseDirectory,
	}

	LogDirectory = &Flag[string, string]{
		Name:          "log-directory",
		Shorthand:     "l",
		Usage:         "Target directory for file logging",
		Default:       ".logs",
		LegacyEnvVars: []string{constants.EnvSkuidLoggingLocation},
		Global:        true,
	}
)
