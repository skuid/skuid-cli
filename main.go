package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/logging"

	"github.com/skuid/skuid-cli/cmd"
)

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

// VersionName has to be here instead of in the constants package
// because embed does not allow relative paths (parent directories)

//go:embed .version
var VersionName string

func init() {
	VersionName = strings.TrimSpace(VersionName)
	constants.VersionName = VersionName
}

func main() {
	code := Run()
	os.Exit(int(code))
}

// Run is a function so that TestMain can execute it
func Run() exitCode {
	start := time.Now()
	factory := cmdutil.NewFactory(VersionName)
	defer Teardown(factory)
	if err := cmd.NewCmdRoot(factory).Execute(); err != nil {
		finished := time.Now()
		duration := finished.Sub(start)
		msgFormat := "%v %v\n"
		// note that colors may be disabled if --file-logging was specified
		msgArgs := []any{logging.ColorFailureIcon(), logging.ColorFailure.Sprint(err)}
		// logging isn't initialized until root command PersistentPreRunE, however error may have occurred
		// prior to it running (e.g., Cobra error due to invalid flag) so if we have initialized logging,
		// output there, otherwise output to stderr.
		// TODO: Logging and handling output needs to be reworked - see https://github.com/skuid/skuid-cli/issues?q=is%3Aissue+is%3Aopen+logging
		//       In short, most of the issues with Logging are resolved in https://github.com/skuid/skuid-cli/pull/205, however the two
		//       things left are:
		//          1. eliminate the global singleton and use dependency injection of a Logger which will aid with testability and use of the logger
		//          2. the command line flags can be pre-parsed and logging initialized prior to Cobra even starting its pipeline allowing full
		//             control over stdin/out/err across both Cobra and Logging.
		if factory.LogConfig.IsInitialized() {
			var cmdError *cmdutil.CommandError
			cmdName := "unknown"
			if errors.As(err, &cmdError) {
				cmdName = cmdError.Name
			}
			fields := logging.TrackingFields(start, finished, duration, false, err, logging.Fields{
				logging.CommandNameKey: cmdName,
			})
			logging.WithName("main.Run", fields).Errorf(msgFormat, msgArgs...)
		} else {
			// if logging isn't initialized, output directly to stderr and enable colors
			// intentionally ignoring return values
			_, _ = fmt.Fprintf(os.Stderr, msgFormat, msgArgs...)
		}

		return exitError
	}

	return exitOK
}

func Teardown(factory *cmdutil.Factory) {
	// We teardown logging (e.g., closing file) here instead of in the command
	// for two reasons:
	//   1. The Cobra *PostRun(E) functions are not called when RunE returns an error
	//   2. Cobra.OnFinalize is global hook so during tests, registering a finalizer
	//      will result in it being called every time a cobra command is executing and
	//      there is no way to "clear/reset" finalizers.  In short, while OnFinalize
	//      approach would work during normal execution as it would run only once, it
	//      impacts testability
	factory.LogConfig.Teardown()
}
