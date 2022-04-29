package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var (
	GlobalArgVerbose bool
)

var (
	// TidesCmd represents the base command when called without any subcommands
	TidesCmd = &cobra.Command{
		Use:     constants.PROJECT_NAME,
		Short:   "Tides: A CLI for interacting with Skuid APIs",
		Long:    `Tides: Deploy and retrieve Skuid metadata to / from Skuid NLX.`,
		Version: constants.VERSION_NAME,
	}
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigName(".skuid") // name of config file (without extension)
		viper.AddConfigPath("$HOME")  // adding home directory as first search path

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			logging.Println("Using config file:", viper.ConfigFileUsed())
		}
	})

	if err := flags.Add(flags.Verbose)(TidesCmd); err != nil {
		logging.Fatal(err)
	}

}