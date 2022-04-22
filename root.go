package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
		Use:     "skuid",
		Short:   "A CLI for interacting with Skuid APIs",
		Long:    `Deploy and retrieve Skuid metadata to / from Skuid Platform or Skuid on Salesforce`,
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

	fs := RootCmd.PersistentFlags()

	fs.StringVar(&ArgHost, "host", os.Getenv("SKUID_HOST"),
		"API Host base URL, e.g. https://my-site.skuidsite.com for Skuid Platform or https://my-domain.my.salesforce.com for Salesforce")
	RootCmd.MarkFlagRequired("host")

	fs.StringVarP(&ArgPassword, "password", "p", os.Getenv("SKUID_PW"), "Skuid Platform / Salesforce Password")
	RootCmd.MarkFlagRequired("password")

	fs.StringVarP(&ArgUsername, "username", "u", os.Getenv("SKUID_UN"), "Skuid Platform / Salesforce Username")
	RootCmd.MarkFlagRequired("username")

	fs.StringVar(&ArgAppClientID, "client-id", os.Getenv("SKUID_CLIENT_ID"), "OAuth Client ID")
	fs.StringVar(&ArgAppClientSecret, "client-secret", os.Getenv("SKUID_CLIENT_SECRET"), "OAuth Client Secret")
	fs.StringVar(&ArgApiVersion, "api-version", "", "API Version")
	fs.StringVarP(&ArgModule, "module", "m", "", "Module name(s), separated by a comma.")
	fs.StringVarP(&ArgPage, "page", "n", "", "Page name(s), separated by a comma.")
	fs.StringVarP(&ArgTargetDir, "dir", "d", "", "Input/output directory.")
	fs.StringVarP(&ArgMetadataServiceProxy, "metadataServiceProxy", "", os.Getenv("METADATA_SERVICE_PROXY"), "Proxy used to reach the metadata service")
	fs.StringVarP(&ArgDataServiceProxy, "dataServiceProxy", "", os.Getenv("DATA_SERVICE_PROXY"), "Proxy used to reach the data service")
	fs.StringVar(&ArgVariableName, "name", "", "The display name for the variable to be set")
	fs.StringVar(&ArgVariableValue, "value", "", "The value for the variable to be set")
	fs.StringVar(&ArgVariableDataService, "dataservice", "", "Optional, the name of a private data service in which the variable should be created. Leave blank to store in the default data service.")
	fs.BoolVarP(&ArgVerbose, "verbose", "v", false, "Display all possible logging info")
	fs.StringVarP(&ArgAppName, "app", "a", "", "Filter retrieve/deploy to a specific Skuid App")
	fs.BoolVarP(&ArgNoZip, "nozip", "z", false, "Ask for site not to be zipped")
}
