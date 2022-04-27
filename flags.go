package main

import (
	"log"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

// Flag Names
const (
	FlagNameHost                 = "host"          // Flag name for ArgHost
	FlagNamePassword             = "password"      // Flag name for ArgPassword
	FlagNameAppClientId          = "client-id"     // Flag name for ArgClientId
	FlagNameVerbose              = "verbose"       // Flag Name for GlobalArgVerbose
	FlagNameUsername             = "username"      // Flag name for ArgUsername
	FlagNameClientSecret         = "client-secret" // Flag name for ArgAppClientSecret
	FlagNameApiVersion           = "api-version"
	FlagNameModules              = "module"
	FlagNamePages                = "page"
	FlagNameDirectory            = "dir"
	FlagNameMetadataServiceProxy = "metadataServiceProxy"
	FlagNameDataserviceProxy     = "dataServiceProxy"
	FlagNameDisplayName          = "name"
	FlagNameVariableValue        = "value"
	FlagNameVariableDataService  = "dataservice"
	FlagNameAppName              = "app"
	FlagNameNoZip                = "nozip"
)

var (
	// PlatformLoginFlags adds the required necessary flags to a command
	// for the function PlatformLogin
	PlatformLoginFlags = []func(cmd *cobra.Command) error{
		AddFlagHost, AddFlagUsername, AddFlagPassword, AddFlagApiVersion, AddFlagMetadataServiceProxy, AddFlagDataServiceProxy,
	}
)

// AddFlags is a function that takes a command and adds a series of flags
// and if there's an issue adding the flags (usually just a null pointer)
// we'll log fatal
func AddFlags(cmd *cobra.Command, adds ...func(*cobra.Command) error) {
	for _, addFlag := range adds {
		if err := addFlag(cmd); err != nil {
			log.Fatalf("Unable to set up %v: %v", cmd.Name(), err.Error())
		}
	}
}

// If we find an environment variable indicate that in the usage text so that there
// is some additional information given
func EnvironmentVariableFound(environmentVariableName, usageText string) string {
	environmentVariableValue := os.Getenv(environmentVariableName)
	usageText = color.Gray.Sprint(usageText)
	if environmentVariableValue != "" {
		usageText = usageText + "\n" +
			color.Green.Sprintf(`(Environment Variable: "$%v" FOUND)`, environmentVariableName)
	} else {
		usageText = usageText + "\n" +
			color.Yellow.Sprintf(`(Environment Variable: "$%v")`, environmentVariableName)
	}
	return usageText
}

// AddFlagHost adds the flag --host "host" to
// the provided command. Gray helptext, yellow
// text provided if the environment variable ENV_SKUID_HOST
// is set.
func AddFlagHost(cmd *cobra.Command) error {
	envVar := os.Getenv(ENV_SKUID_HOST)

	usageText := EnvironmentVariableFound(
		ENV_SKUID_HOST,
		`Host URL, e.g. [ https://my.skuidsite.com | my.skuidsite.com ]`,
	)

	cmd.Flags().StringVar(
		&ArgHost,
		FlagNameHost,
		envVar,
		usageText,
	)

	if envVar != "" {
		return nil
	} else {
		return cmd.MarkFlagRequired(FlagNameHost)
	}
}

// AddFlagPassword adds "password" and "-p" flag
// to a command. If the environment variable is found
// we don't mark it required. if it isn't, we do.
func AddFlagPassword(cmd *cobra.Command) error {
	envVar := os.Getenv(ENV_SKUID_PASSWORD)

	usageText := EnvironmentVariableFound(
		ENV_SKUID_PASSWORD,
		"Skuid Platform / Salesforce Password",
	)

	cmd.Flags().StringVarP(
		&ArgPassword,
		FlagNamePassword,
		string(FlagNamePassword[0]),
		envVar,
		usageText,
	)

	// if we set a default value,
	// we shouldn't mark it as required.
	// this will cause it to fail if the environment
	// variables are there.
	if envVar != "" {
		return nil
	} else {
		return cmd.MarkFlagRequired(FlagNamePassword)
	}
}

func AddFlagUsername(cmd *cobra.Command) error {
	envVar := os.Getenv(ENV_SKUID_USERNAME)

	usageText := EnvironmentVariableFound(
		ENV_SKUID_USERNAME,
		"Skuid Platform / Salesforce Username",
	)

	cmd.Flags().StringVarP(
		&ArgUsername,
		FlagNameUsername,
		string(FlagNameUsername[0]),
		envVar,
		usageText,
	)

	if envVar != "" {
		return nil
	} else {
		return cmd.MarkFlagRequired(FlagNameUsername)
	}
}

func AddFlagClientId(cmd *cobra.Command) error {
	envVar := os.Getenv(ENV_SKUID_CLIENT_ID)

	usageText := EnvironmentVariableFound(ENV_SKUID_CLIENT_ID, "OAuth Client ID")

	cmd.Flags().StringVar(
		&ArgClientId,
		FlagNameAppClientId,
		envVar,
		usageText,
	)

	if envVar != "" {
		return nil
	} else {
		return cmd.MarkFlagRequired(FlagNameAppClientId)
	}
}

func AddFlagOutputFilename(cmd *cobra.Command) error {
	cmd.Flags().StringVarP(
		&ArgOutputFileName,
		"output",
		"o",
		"",
		"Filename of output Page Pack file",
	)

	return nil // no required check
}

func AddFlagApiVersion(cmd *cobra.Command) error {

	cmd.Flags().StringVar(
		&ArgApiVersion,
		FlagNameApiVersion,
		"",
		"API Version",
	)

	return nil
}

func AddFlagMetadataServiceProxy(cmd *cobra.Command) error {
	envVar := os.Getenv(ENV_METADATA_SERVICE_PROXY)

	usageText := EnvironmentVariableFound(ENV_METADATA_SERVICE_PROXY, "Proxy used to reach the metadata service")

	cmd.Flags().StringVarP(
		&ArgMetadataServiceProxy,
		FlagNameMetadataServiceProxy,
		"",
		envVar,
		usageText,
	)

	return nil
}

func AddFlagDataServiceProxy(cmd *cobra.Command) error {
	envVar := os.Getenv(ENV_DATA_SERVICE_PROXY)

	usageText := EnvironmentVariableFound(ENV_DATA_SERVICE_PROXY, "Proxy used to reach the data service")

	cmd.Flags().StringVarP(
		&ArgDataServiceProxy,
		FlagNameDataserviceProxy,
		"",
		envVar,
		usageText,
	)

	return nil
}
