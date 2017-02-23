package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var sfdcUsername string
var sfdcPassword string
var appClientID string
var appClientSecret string
var apiVersion string
var sfdcHost string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "skuid",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.skuid.yaml)")
	RootCmd.PersistentFlags().StringVar(&sfdcUsername, "username", os.Getenv("SKUID_UN"), "SFDC Username")
	RootCmd.PersistentFlags().StringVar(&sfdcPassword, "password", os.Getenv("SKUID_PW"), "SFDC Password")
	RootCmd.PersistentFlags().StringVar(&appClientID, "client-id", os.Getenv("SKUID_CLIENT_ID"), "Connected App Client ID")
	RootCmd.PersistentFlags().StringVar(&appClientSecret, "client-secret", os.Getenv("SKUID_CLIENT_SECRET"), "Connected App Client Secret")
	RootCmd.PersistentFlags().StringVar(&apiVersion, "api-version", "36.0", "SFDC API Version (default: 36.0)")
	RootCmd.PersistentFlags().StringVar(&sfdcHost, "host", "https://login.salesforce.com", "SDFC Login Url")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".skuid") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
