package flags

import (
	"github.com/skuid/tides/pkg/constants"
)

var (
	Host = &Flag[string]{
		argument:   &argHost,
		Name:       "host",
		Usage:      `Host URL, e.g. [ https://my.skuidsite.com | my.skuidsite.com ]`,
		EnvVarName: constants.ENV_SKUID_HOST,
		Required:   true,
	}

	Password = &Flag[string]{
		argument:   &argPassword,
		Name:       "password",
		Shorthand:  "p",
		Usage:      "Skuid Platform / Salesforce Password",
		EnvVarName: constants.ENV_SKUID_PASSWORD,
		Required:   true,
	}

	Username = &Flag[string]{
		argument:   &argUsername,
		Name:       "username",
		Shorthand:  "u",
		EnvVarName: constants.ENV_SKUID_USERNAME,
		Required:   true,
		Usage:      "Skuid Platform / Salesforce Username",
	}

	ClientId = &Flag[string]{
		argument:   &argClientId,
		Name:       "client-id",
		EnvVarName: constants.ENV_SKUID_CLIENT_ID,
		Required:   false,
		Usage:      "OAuth Client ID",
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
		Default:  constants.DEFAULT_SALESFORCE_API_VERSION,
	}

	MetadataServiceProxy = &Flag[string]{
		argument:   &argMetadataServiceProxy,
		Name:       "metadata-service-proxy",
		EnvVarName: constants.ENV_METADATA_SERVICE_PROXY,
		Usage:      "Proxy used to reach the metadata service",
	}

	DataServiceProxy = &Flag[string]{
		argument:   &argDataServiceProxy,
		Name:       "data-service-proxy",
		EnvVarName: constants.ENV_DATA_SERVICE_PROXY,
		Usage:      "Proxy used to reach the data service",
	}

	ClientSecret = &Flag[string]{
		argument:   &argClientSecret,
		Name:       "client-secret",
		EnvVarName: constants.ENV_SKUID_CLIENT_SECRET,
		Usage:      "OAuth Client Secret",
	}

	AppName = &Flag[string]{
		argument:  &argAppName,
		Name:      "app-name",
		Shorthand: "a",
		Usage:     "Scope the operation to a specific Skuid App (name).",
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
		argument:  &argTargetDir,
		Name:      "dir",
		Aliases:   []string{"directory"},
		Shorthand: "d",
		Usage:     "Target directory for this operation.",
	}

	PushFile = &Flag[string]{
		argument:  &argPushFile,
		Name:      "file",
		Shorthand: "f",
		Usage:     "Skuid Page file(s) to push. Supports file globs.",
	}

	Module = &Flag[string]{
		argument:  &argModule,
		Name:      "module",
		Shorthand: "m",
		Usage:     "Module to pack",
	}
)
