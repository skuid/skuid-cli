package flags

import (
	"github.com/skuid/skuid-cli/pkg/constants"
)

var (
	PlinyHost = &Flag[string]{
		Name:        "host",
		Usage:       `Host URL, e.g. [ https://my.skuidsite.com | my.skuidsite.com ]`,
		EnvVarNames: []string{constants.ENV_SKUID_HOST, constants.ENV_PLINY_HOST},
		Required:    true,
	}

	Password = &Flag[string]{
		argument:    &argPassword,
		Name:        "password",
		Shorthand:   "p",
		Usage:       "Skuid NLX Password",
		EnvVarNames: []string{constants.ENV_SKUID_PASSWORD},
		Required:    true,
	}

	Username = &Flag[string]{
		Name:        "username",
		Shorthand:   "u",
		EnvVarNames: []string{constants.ENV_SKUID_USERNAME},
		Required:    true,
		Usage:       "Skuid NLX Username",
	}

	AppName = &Flag[string]{
		Name:      "app",
		Shorthand: "a",
		Usage:     "Scope the operation to a specific Skuid NLX App (name)",
	}

	Directory = &Flag[string]{
		Name:        "dir",
		Shorthand:   "d",
		Usage:       "Target directory for this operation",
		EnvVarNames: []string{constants.ENV_SKUID_DEFAULT_FOLDER},
	}

	FileLoggingDirectory = &Flag[string]{
		Name:        "log-directory",
		Shorthand:   "l",
		Usage:       "Target directory for file logging",
		Default:     ".logs",
		EnvVarNames: []string{constants.ENV_SKUID_LOGGING_LOCATION},
		Global:      true,
	}
)
