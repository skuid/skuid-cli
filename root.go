package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: Comment each of these
var (
	ArgAppClientID          string
	ArgAppClientSecret      string
	ArgApiVersion           string
	ArgHost                 string
	ArgAppName              string
	ArgModule               string
	ArgPage                 string
	ArgPassword             string
	ArgTargetDir            string
	ArgUsername             string
	ArgMetadataServiceProxy string
	ArgDataServiceProxy     string
	ArgVariableName         string
	ArgVariableValue        string
	ArgVariableDataService  string
	ArgVerbose              bool
	ArgNoZip                bool

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

	// required flags
	rootCommandFlags.StringVar(&ArgHost, "host", os.Getenv(ENV_SKUID_HOST),
		"API Host base URL, e.g. https://my-site.skuidsite.com for Skuid Platform or https://my-domain.my.salesforce.com for Salesforce")
	rootCommandFlags.StringVarP(&ArgPassword, "password", "p", os.Getenv(ENV_SKUID_PASSWORD), "Skuid Platform / Salesforce Password")
	rootCommandFlags.StringVarP(&ArgUsername, "username", "u", os.Getenv(ENV_SKUID_USERNAME), "Skuid Platform / Salesforce Username")

	// these three are required
	RootCmd.MarkFlagRequired("host")
	RootCmd.MarkFlagRequired("password")
	RootCmd.MarkFlagRequired("username")

	rootCommandFlags.StringVar(&ArgAppClientID, "client-id", os.Getenv(ENV_SKUID_CLIENT_ID), "OAuth Client ID")
	rootCommandFlags.StringVar(&ArgAppClientSecret, "client-secret", os.Getenv(ENV_SKUID_CLIENT_SECRET), "OAuth Client Secret")
	rootCommandFlags.StringVar(&ArgApiVersion, "api-version", "", "API Version")
	rootCommandFlags.StringVarP(&ArgModule, "module", "m", "", "Module name(s), separated by a comma.")
	rootCommandFlags.StringVarP(&ArgPage, "page", "n", "", "Page name(s), separated by a comma.")
	rootCommandFlags.StringVarP(&ArgTargetDir, "dir", "d", "", "Input/output directory.")
	rootCommandFlags.StringVarP(&ArgMetadataServiceProxy, "metadataServiceProxy", "", os.Getenv(ENV_METADATA_SERVICE_PROXY), "Proxy used to reach the metadata service")
	rootCommandFlags.StringVarP(&ArgDataServiceProxy, "dataServiceProxy", "", os.Getenv(ENV_DATA_SERVICE_PROXY), "Proxy used to reach the data service")
	rootCommandFlags.StringVar(&ArgVariableName, "name", "", "The display name for the variable to be set")
	rootCommandFlags.StringVar(&ArgVariableValue, "value", "", "The value for the variable to be set")
	rootCommandFlags.StringVar(&ArgVariableDataService, "dataservice", "", "Optional, the name of a private data service in which the variable should be created. Leave blank to store in the default data service.")
	rootCommandFlags.BoolVarP(&ArgVerbose, "verbose", "v", false, "Display all possible logging info")
	rootCommandFlags.StringVarP(&ArgAppName, "app", "a", "", "Filter retrieve/deploy to a specific Skuid App")
	rootCommandFlags.BoolVarP(&ArgNoZip, "nozip", "z", false, "Ask for site not to be zipped")
}
