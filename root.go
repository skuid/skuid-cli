package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ArgHost                 string
	ArgPassword             string
	ArgUsername             string
	ArgClientId             string
	GlobalArgVerbose        bool
	ArgAppClientSecret      string
	ArgApiVersion           string
	ArgAppName              string
	ArgModules              []string
	ArgPages                []string
	ArgTargetDir            string
	ArgMetadataServiceProxy string
	ArgDataServiceProxy     string
	ArgVariableName         string
	ArgVariableValue        string
	ArgVariableDataService  string
	ArgNoZip                bool
)

var (
	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:     PROJECT_NAME,
		Short:   "Tides: A CLI for interacting with Skuid APIs",
		Long:    `Tides: Deploy and retrieve Skuid metadata to / from Skuid Platform or Skuid on Salesforce`,
		Version: VERSION_NAME,
	}
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigName(".skuid") // name of config file (without extension)
		viper.AddConfigPath("$HOME")  // adding home directory as first search path

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			Println("Using config file:", viper.ConfigFileUsed())
		}

		// Graciously handle an org/site host that does not start "https://"
		if !strings.HasPrefix(ArgHost, "https://") {
			ArgHost = "https://" + ArgHost
		}
	})

	rootCommandFlags := RootCmd.PersistentFlags()

	// globals
	rootCommandFlags.BoolVarP(
		&GlobalArgVerbose,
		FlagNameVerbose,
		"v",
		false,
		"Display all possible logging info",
	)

	rootCommandFlags.BoolVarP(
		&ArgNoZip,
		FlagNameNoZip,
		"z",
		false,
		"Ask for site not to be zipped",
	)

	rootCommandFlags.StringArrayVarP(
		&ArgModules,
		FlagNameModules,
		"m",
		[]string{},
		"Module name(s), separated by a comma.",
	)

	rootCommandFlags.StringArrayVarP(
		&ArgPages,
		FlagNamePages,
		"n",
		[]string{},
		"Page name(s), separated by a comma.",
	)

	rootCommandFlags.StringVarP(
		&ArgTargetDir,
		FlagNameDirectory,
		"d",
		"",
		"Input/output directory.",
	)

	rootCommandFlags.StringVar(
		&ArgVariableName,
		FlagNameDisplayName,
		"",
		"The display name for the variable to be set",
	)

	rootCommandFlags.StringVar(
		&ArgVariableValue,
		FlagNameVariableValue,
		"",
		"The value for the variable to be set",
	)

	rootCommandFlags.StringVar(
		&ArgVariableDataService,
		FlagNameVariableDataService,
		"",
		"Optional, the name of a private data service in which the variable should be created. Leave blank to store in the default data service.",
	)

	rootCommandFlags.StringVarP(
		&ArgAppName,
		FlagNameAppName,
		"a",
		"",
		"Filter retrieve/deploy to a specific Skuid App",
	)

	rootCommandFlags.StringVar(
		&ArgAppClientSecret,
		FlagNameClientSecret,
		os.Getenv(ENV_SKUID_CLIENT_SECRET),
		"OAuth Client Secret",
	)

}
