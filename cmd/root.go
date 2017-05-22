package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appClientID string
var appClientSecret string
var apiVersion string
var host string
var module string
var password string
var targetDir string
var username string
var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "skuid",
	Short: "A CLI for interating with Skuid APIs",
	Long:  `Deploy and retrieve Skuid metadata to / from Skuid Platform or Skuid on Salesforce`,
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

	RootCmd.PersistentFlags().StringVar(&appClientID, "client-id", "", "OAuth Client ID")
	RootCmd.PersistentFlags().StringVar(&appClientSecret, "client-secret", "", "OAuth Client Secret")
	RootCmd.PersistentFlags().StringVar(&apiVersion, "api-version", "", "API Version")
	RootCmd.PersistentFlags().StringVar(&host, "host", "", 
		"API Host base URL, e.g. my-site.skuidsite.com for Skuid Platform or my-domain.my.salesforce.com for Salesforce")
	RootCmd.PersistentFlags().StringVarP(&module, "module", "m", "", "Module name(s), separated by a comma.")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Skuid Platform / Salesforce Password")
	RootCmd.PersistentFlags().StringVarP(&targetDir, "dir", "d", "", "Input/output directory.")
	RootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Skuid Platform / Salesforce Username")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display all possible logging info")
	
	if username == "" {
		username = os.Getenv("SKUID_UN")
	}

	if password == "" {
		password = os.Getenv("SKUID_PW")
	}

	if appClientID == "" {
		appClientID = os.Getenv("SKUID_CLIENT_ID")
	}

	if appClientSecret == "" {
		appClientSecret = os.Getenv("SKUID_CLIENT_SECRET")
	}

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetConfigName(".skuid") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
