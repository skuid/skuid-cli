package common

import (
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
)

// PrerunValidation sets up logging according to command flags
func PrerunValidation(cmd *cobra.Command, _ []string) error {
	logging.Get().Infof("Skuid CLI Version %v", constants.VERSION_NAME)

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

	fileLoggingEnabled, err := cmd.Flags().GetBool(flags.FileLogging.Name)
	if err != nil {
		return err
	}

	loggingDirectory, err := cmd.Flags().GetString(flags.FileLoggingDirectory.Name)
	if err != nil {
		return err
	}

	if fileLoggingEnabled {
		logging.SetFileLogging(loggingDirectory)
	}
	return nil
}
