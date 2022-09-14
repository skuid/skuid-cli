package common

import (
	"github.com/skuid/domain/flags"
	"github.com/skuid/domain/logging"
	"github.com/spf13/cobra"
)

// PrerunValidation does generic validation for a function to make sure it has
// https (as was in the main function)
func PrerunValidation(cmd *cobra.Command, _ []string) error {
	// set verbosity
	if verbose, err := cmd.Flags().GetBool(flags.Verbose.Name); err != nil {
		return err
	} else if verbose {
		logging.SetVerbose()
	}

	if trace, err := cmd.Flags().GetBool(flags.Trace.Name); err != nil {
		return err
	} else if trace {
		logging.SetTrace()
	}

	if diagnostic, err := cmd.Flags().GetBool(flags.Diagnostic.Name); err != nil {
		return err
	} else if diagnostic {
		logging.SetDiagnostic()
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
