package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/skuid/tides/cmd/common"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/ui"
)

var (
	GlobalArgVerbose bool
)

var (
	// TidesCmd represents the base command when called without any subcommands
	TidesCmd = &cobra.Command{
		Use:   constants.PROJECT_NAME,
		Short: "Tides: A CLI for interacting with Skuid APIs",
		Long:  `Tides: Deploy and retrieve Skuid metadata to / from Skuid pkg.`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: constants.VERSION_NAME,
		PreRunE: common.LoggingValidation,
		Run: func(cmd *cobra.Command, _ []string) {

			// want to hide the logger if we're not file logging
			if fileLogging, err := cmd.Flags().GetBool(flags.FileLogging.Name); err != nil {
				logging.Logger.WithError(err).Panic("we need to know if we're file logging")
			} else if !fileLogging {
				// disable the logger
				fmt.Println("NO NO LOG")
				logging.Logger.Out = nil
				logging.Logger.SetLevel(logrus.PanicLevel)
			}

			p := tea.NewProgram(ui.Main(cmd))
			if err := p.Start(); err != nil {
				logging.Logger.WithError(err).Error("Unable to Start User Interface.")
			}
		},
	}
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigName(".skuid") // name of config file (without extension)
		viper.AddConfigPath("$HOME")  // adding home directory as first search path

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			logging.Logger.Println("Using config file:", viper.ConfigFileUsed())
		}
	})

	if err := flags.Add(flags.Verbose)(TidesCmd); err != nil {
		logging.Logger.WithError(err).Fatal("Unable to assign verbose flag to command")
	}

	if err := flags.Add(flags.FileLogging)(TidesCmd); err != nil {
		logging.Logger.WithError(err).Fatal("Unable to assign file logging flag to command")
	}

	if err := flags.Add(flags.FileLoggingDirectory)(TidesCmd); err != nil {
		logging.Logger.WithError(err).Fatal("Unable to assign file logging directory flag to command")
	}

}
