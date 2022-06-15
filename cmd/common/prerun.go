package common

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

// PrerunValidation does generic validation for a function to make sure it has
// https (as was in the main function)
func PrerunValidation(cmd *cobra.Command, _ []string) error {
	if host, err := cmd.Flags().GetString(flags.PlinyHost.Name); err != nil {
		return err
	} else {
		// host validation: it must have https:// to start
		if !strings.HasPrefix("https://", host) {
			// do nothing
			host = fmt.Sprintf("https://%v", host)
		} else if err := cmd.Flags().Set(flags.PlinyHost.Name, host); err != nil {
			return err
		}
	}

	// set verbosity
	if verbose, err := cmd.Flags().GetBool(flags.Verbose.Name); err != nil {
		return err
	} else if verbose {
		logging.Logger.SetLevel(logrus.TraceLevel)
	}

	if err := LoggingValidation(cmd, []string{}); err != nil {
		return err
	}

	return nil
}

func LoggingValidation(cmd *cobra.Command, _ []string) (err error) {
	var fileLoggingEnabled bool
	if fileLoggingEnabled, err = cmd.Flags().GetBool(flags.FileLogging.Name); err != nil {
		return
	}

	var loggingDirectory string
	if loggingDirectory, err = cmd.Flags().GetString(flags.FileLoggingDirectory.Name); err != nil {
		return
	}

	// try to open a file for this run off this
	if fileLoggingEnabled {
		logging.SetFileLogging(loggingDirectory)
	}

	return

}
