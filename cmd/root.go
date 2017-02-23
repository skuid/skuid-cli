package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sfdcUsername string
var sfdcPassword string
var appClientID string
var appClientSecret string
var apiVersion string
var sfdcHost string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "skuid",
	Short: "A CLI for interating with Skuid Pages",
	Long:  `Push and Pull Skuid pages to and from Skuid running on the Salesforce Platform`,
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

	RootCmd.PersistentFlags().StringVar(&sfdcUsername, "username", "", "Salesforce Username")
	RootCmd.PersistentFlags().StringVar(&sfdcPassword, "password", "", "Salesforce Password")
	RootCmd.PersistentFlags().StringVar(&appClientID, "client-id", "", "Connected App Client ID")
	RootCmd.PersistentFlags().StringVar(&appClientSecret, "client-secret", "", "Connected App Client Secret")
	RootCmd.PersistentFlags().StringVar(&apiVersion, "api-version", "36.0", "Salesforce API Version")
	RootCmd.PersistentFlags().StringVar(&sfdcHost, "host", "https://login.salesforce.com", "Salesforce Login Url")

	if sfdcUsername == "" {
		sfdcUsername = os.Getenv("SKUID_UN")
	}

	if sfdcPassword == "" {
		sfdcPassword = os.Getenv("SKUID_PW")
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
