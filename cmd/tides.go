package cmd

import (
	"github.com/skuid/domain/constants"
	"github.com/skuid/domain/flags"
	"github.com/skuid/domain/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/skuid/tides/cmd/common"
)

var (
	GlobalArgVerbose bool
)

var (
	// TidesCmd represents the base command when called without any subcommands
	TidesCmd = &cobra.Command{
		Use:   constants.PROJECT_NAME,
		Short: "Tides: A CLI for interacting with Skuid APIs",
		Long:  `Tides: Deploy and retrieve Skuid metadata to / from Skuid domain.`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version:           constants.VERSION_NAME,
		PersistentPreRunE: common.PrerunValidation,
	}
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigName(".skuid") // name of config file (without extension)
		viper.AddConfigPath("$HOME")  // adding home directory as first search path

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			logging.Get().Debug("Using config file:", viper.ConfigFileUsed())
		}
	})

	flags.AddFlags(TidesCmd, flags.Verbose, flags.Trace, flags.FileLogging, flags.Diagnostic)
	flags.AddFlags(TidesCmd, flags.FileLoggingDirectory)
}
