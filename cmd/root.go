package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/skuid/skuid-cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	appClientID          string
	appClientSecret      string
	apiVersion           string
	host                 string
	appName              string
	module               string
	page                 string
	password             string
	targetDir            string
	username             string
	metadataServiceProxy string
	dataServiceProxy     string
	variablename         string
	variablevalue        string
	variabledataservice  string
	verbose              bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "skuid",
	Short:   "A CLI for interacting with Skuid APIs",
	Long:    `Deploy and retrieve Skuid metadata to / from Skuid Platform or Skuid on Salesforce`,
	Version: version.Name,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&appClientID, "client-id", os.Getenv("SKUID_CLIENT_ID"), "OAuth Client ID")
	RootCmd.PersistentFlags().StringVar(&appClientSecret, "client-secret", os.Getenv("SKUID_CLIENT_SECRET"), "OAuth Client Secret")
	RootCmd.PersistentFlags().StringVar(&apiVersion, "api-version", "", "API Version")
	RootCmd.PersistentFlags().StringVar(&host, "host", os.Getenv("SKUID_HOST"),
		"API Host base URL, e.g. https://my-site.skuidsite.com for Skuid Platform or https://my-domain.my.salesforce.com for Salesforce")
	RootCmd.PersistentFlags().StringVarP(&module, "module", "m", "", "Module name(s), separated by a comma.")
	RootCmd.PersistentFlags().StringVarP(&page, "page", "n", "", "Page name(s), separated by a comma.")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", os.Getenv("SKUID_PW"), "Skuid Platform / Salesforce Password")
	RootCmd.PersistentFlags().StringVarP(&targetDir, "dir", "d", "", "Input/output directory.")
	RootCmd.PersistentFlags().StringVarP(&username, "username", "u", os.Getenv("SKUID_UN"), "Skuid Platform / Salesforce Username")
	RootCmd.PersistentFlags().StringVarP(&metadataServiceProxy, "metadataServiceProxy", "", os.Getenv("METADATA_SERVICE_PROXY"), "Proxy used to reach the metadata service")
	RootCmd.PersistentFlags().StringVarP(&dataServiceProxy, "dataServiceProxy", "", os.Getenv("DATA_SERVICE_PROXY"), "Proxy used to reach the data service")
	RootCmd.PersistentFlags().StringVar(&variablename, "name", "", "The display name for the variable to be set")
	RootCmd.PersistentFlags().StringVar(&variablevalue, "value", "", "The value for the variable to be set")
	RootCmd.PersistentFlags().StringVar(&variabledataservice, "dataservice", "", "Optional, the name of a private data service in which the variable should be created. Leave blank to store in the default data service.")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display all possible logging info")
	RootCmd.PersistentFlags().StringVarP(&appName, "app", "a", "", "Filter retrieve/deploy to a specific Skuid App")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetConfigName(".skuid") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Ensure we have universally-required parameters
	if username == "" {
		fmt.Println("No Username provided - request could not be performed.")
		os.Exit(1)
	}

	if password == "" {
		fmt.Println("No Password provided - request could not be performed.")
		os.Exit(1)
	}

	if host == "" {
		fmt.Println("No Host provided - request could not be performed.")
		os.Exit(1)
	}

	// Graciously handle an org/site host that does not start "https://"
	if !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
}
