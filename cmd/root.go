package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var VERSION_NAME string

var (
	// SkuidCmd represents the base command when called without any subcommands
	SkuidCmd = &cobra.Command{
		Use:   constants.PROJECT_NAME,
		Short: "A CLI for Skuid APIs",
		Long:  `A command-line interface used to retrieve and deploy Skuid NLX sites.`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: VERSION_NAME,
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

	flags.AddFlags(SkuidCmd, flags.Verbose, flags.Trace, flags.FileLogging, flags.Diagnostic)
	flags.AddFlags(SkuidCmd, flags.FileLoggingDirectory)
}
