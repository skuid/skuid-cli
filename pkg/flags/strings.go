package flags

import (
	"github.com/skuid/tides/pkg/constants"
)

var (
	Host = &Flag[string]{
		argument: &argHost,
		Name:     "host",
		// Aliases:    []string{"hostname"},
		Usage:      `Host URL, e.g. [ https://my.skuidsite.com | my.skuidsite.com ]`,
		EnvVarName: constants.ENV_SKUID_HOST,
		Required:   true,
	}

	Password = &Flag[string]{
		argument:   &argPassword,
		Name:       "password",
		Shorthand:  "p",
		Usage:      "Skuid NLX Password",
		EnvVarName: constants.ENV_SKUID_PASSWORD,
		Required:   true,
	}

	Username = &Flag[string]{
		argument:   &argUsername,
		Name:       "username",
		Shorthand:  "u",
		EnvVarName: constants.ENV_SKUID_USERNAME,
		Required:   true,
		Usage:      "Skuid NLX Username",
	}

	OutputFile = &Flag[string]{
		argument:  &argOutputFilename,
		Name:      "output",
		Shorthand: "o",
		Usage:     "Filename of output file",
		Required:  true,
	}

	ApiVersion = &Flag[string]{
		argument: &argApiVersion,
		Name:     "api-version",
		Usage:    "API Version",
		Default:  "2",
	}

	AppName = &Flag[string]{
		argument:  &argAppName,
		Name:      "app-name",
		Shorthand: "a",
		Usage:     "Scope the operation to a specific Skuid NLX App (name).",
	}

	VariableDataService = &Flag[string]{
		argument: &argVariableDataservice,
		Name:     "dataservice",
		Usage:    "Optional, the name of a private data service in which the variable should be created. Leave blank to store in the default data service.",
	}

	VariableValue = &Flag[string]{
		argument: &argVariableValue,
		Name:     "value",
		Usage:    "The value for the variable to be set",
	}

	VariableName = &Flag[string]{
		argument: &argVariableName,
		Name:     "name",
		Usage:    "The display name for the variable to be set",
	}

	Directory = &Flag[string]{
		argument: &argTargetDir,
		Name:     "directory",
		// Aliases:   []string{"directory"},
		Shorthand: "d",
		Usage:     "Target directory for this operation.",
	}
)
