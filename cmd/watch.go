package cmd

import (
	"fmt"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/cmd/common"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
)

var watchCmd = &cobra.Command{
	SilenceUsage:      true,
	Use:               "watch",
	Short:             "Watch for changes to local Skuid metadata, and deploy changes to a Skuid NLX Site",
	Long:              "Watches for changes to local Skuid metadata on your file system, and automatically deploys the changed files to a Skuid NLX Site",
	PersistentPreRunE: common.PrerunValidation,
	RunE:              Watch,
}

func init() {
	flags.AddFlags(watchCmd, flags.NLXLoginFlags...)
	flags.AddFlags(watchCmd, flags.Directory)
	AppCmd = append(AppCmd, watchCmd)
}

func Watch(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["process"] = "watch"
	// get required arguments
	host, err := cmd.Flags().GetString(flags.PlinyHost.Name)
	if err != nil {
		return
	}
	username, err := cmd.Flags().GetString(flags.Username.Name)
	if err != nil {
		return
	}
	password, err := cmd.Flags().GetString(flags.Password.Name)
	if err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username

	logging.WithFields(fields).Debug("Gathered Credentials")

	var auth *pkg.Authorization
	if auth, err = pkg.Authorize(host, username, password); err != nil {
		return
	}

	fields["authorized"] = true
	logging.WithFields(fields).Debug("Successfully Logged In")

	var targetDir string
	if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	if targetDir, err = util.SanitizePath(targetDir); err != nil {
		return
	}

	fields["targetDirectory"] = targetDir
	logging.WithFields(fields).Debug("Starting Watch")

	// Create our watcher
	w := watcher.New()

	// Only handle one file change per event cycle.
	w.SetMaxEvents(1)

	go func() {
		for {
			select {
			case event := <-w.Event:
				logging.WithFields(fields).Debug("Detected change to file: " + event.Path)
				go func() {
					if err := pkg.DeployModifiedFiles(auth, targetDir, event.Path); err != nil {
						w.Error <- err
					}
				}()
			case err := <-w.Error:
				logging.Get().Fatalf("Unable to handle file change: %v", err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch targetDir recursively for changes.
	if err = w.AddRecursive(targetDir); err != nil {
		return
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	logging.WithFields(fields).Debug("** Now watching the following files for changes... **")
	for path, f := range w.WatchedFiles() {
		logging.WithFields(fields).Debug(fmt.Sprintf("%s: %s", path, f.Name()))
	}
	logging.WithFields(fields).Debug("Waiting for changes..")

	// Start the watching process - it'll check for changes every 100ms.
	if err = w.Start(time.Millisecond * 100); err != nil {
		return
	}

	return
}
