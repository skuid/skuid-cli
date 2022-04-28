package validation

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

// PrerunValidation does generic validation for a function to make sure it has
// https (as was in the main function)
func PrerunValidation(cmd *cobra.Command, _ []string) error {
	if host, err := cmd.Flags().GetString(flags.Host.Name); err != nil {
		return err
	} else {
		// host validation: it must have https:// to start
		if strings.HasPrefix("https://", host) {
			// do nothing
		} else if err := cmd.Flags().Set(flags.Host.Name, fmt.Sprintf("https://%v", host)); err != nil {
			return err
		}
	}

	// set verbosity
	if verbose, err := cmd.Flags().GetBool(flags.Verbose.Name); err != nil {
		return err
	} else {
		logging.SetVerbose(verbose)
	}

	return nil
}
