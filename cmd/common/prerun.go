package common

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var (
	fileStringFormat = func() (ret string) {
		ret = time.RFC3339
		ret = strings.ReplaceAll(ret, " ", "")
		ret = strings.ReplaceAll(ret, ":", "")
		return
	}()
	file *os.File
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

	if err := initLog(cmd, []string{}); err != nil {
		return err
	}

	return nil
}

func initLog(cmd *cobra.Command, _ []string) (err error) {
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
		var wd string
		if wd, err = os.Getwd(); err != nil {
			return
		}

		var dir string
		if strings.Contains(loggingDirectory, wd) {
			dir = loggingDirectory
		} else {
			dir = path.Join(wd, loggingDirectory)
		}

		if stat, e := os.Stat(dir); e != nil {
			if err = os.MkdirAll(dir, 0777); err != nil {
				return err
			}
		} else if !stat.IsDir() {
			err = fmt.Errorf("Directory required at loc: %v", dir)
			return
		}

		logFileName := time.Now().Format(fileStringFormat) + ".log"

		if _, err = os.Create(path.Join(dir, logFileName)); err != nil {
			return
		}

		if file, err = os.OpenFile(path.Join(dir, logFileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
			return
		}

		logging.Logger.SetOutput(file)
	}

	return

}
