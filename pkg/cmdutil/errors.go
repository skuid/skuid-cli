package cmdutil

import (
	"fmt"

	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/spf13/cobra"
)

type CommandError struct {
	// Note: not struct{error}: only *CommandError should satisfy error.
	Name string
	err  error
}

func (ce *CommandError) Error() string {
	return ce.err.Error()
}

func (ce *CommandError) Unwrap() error {
	return ce.err
}

func NewCommandError(name string, err error) error {
	return &CommandError{name, err}
}

// Detect when a panic occurs and if encountered, wrap any err with the panic
// information and return it.  This enables us to ensure that panics are
// written to the log file (if its been initialized).
// Currently only using this in the RunE of commands rather than in every function
// throughout the code base but it could be used in any function as-is and then
// the wrapped error would just bubble up.
func CheckError(cmd *cobra.Command, err error, recovered any) error {
	if recovered == nil {
		return err
	}

	msgFmt := logging.FormatPanic(recovered, logging.QuoteText(cmd.Name()))
	var msgArgs []any
	if err != nil {
		msgFmt += "\nwith error: %v"
		msgArgs = append(msgArgs, err)
	}
	return fmt.Errorf(msgFmt, msgArgs...)
}

// Wrap any returned err from a command with CommandError to allow for main.go to obtain the name of the
// command that was executed for logging purposes.  If no error, log the final success message when
// message not empty string.
func HandleCommandResult(cmd *cobra.Command, logger *logging.Logger, err error, message string) error {
	if err != nil {
		return NewCommandError(cmd.Name(), err)
	} else {
		if message != "" {
			logger.Infof("%v %v", logging.ColorSuccessIcon(), message)
		}
		return nil
	}
}
