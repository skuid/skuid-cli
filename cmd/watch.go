package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/common"
	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/util"
)

var watchCmd = &cobra.Command{
	SilenceErrors:     true,
	SilenceUsage:      true,
	Use:               "watch",
	Short:             "Watch for changes to local Skuid metadata, and deploy changes to a Skuid NLX Site.",
	Long:              "Watches for changes to local Skuid metadata on your file system, and automatically deploys the changed files to a Skuid NLX Site.",
	PersistentPreRunE: common.PrerunValidation,
	RunE:              Watch,
}

func init() {
	TidesCmd.AddCommand(watchCmd)
	flags.AddFlags(watchCmd, flags.NLXLoginFlags...)
	flags.AddFlags(watchCmd, flags.Directory)
}

func Watch(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["process"] = "watch"
	// get required arguments
	var host, username, password string
	if host, err = cmd.Flags().GetString(flags.Host.Name); err != nil {
		return
	} else if username, err = cmd.Flags().GetString(flags.Username.Name); err != nil {
		return
	} else if password, err = cmd.Flags().GetString(flags.Password.Name); err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username

	logging.Logger.WithFields(fields).Debug("Gathered Credentials.")

	var auth *pkg.Authorization
	if auth, err = pkg.Authorize(host, username, password); err != nil {
		return
	}

	fields["authorized"] = true
	logging.Logger.WithFields(fields).Debug("Successfully Logged In.")

	var targetDir string
	if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	// If target directory is provided,
	var back string
	back, err = os.Getwd()
	if err != nil {
		return
	}

	// switch back to directory
	defer func() {
		if targetDir != "" {
			if err := os.Chdir(back); err != nil {
				logging.Logger.WithFields(fields).WithError(err).Fatal("Failed changing back to directory: %v")
			}
		}
	}()

	if targetDir == "" {
		targetDir = "."
	}

	var friendly string
	if friendly, err = util.SanitizePath(targetDir); err != nil {
		return
	}

	// Create our watcher
	w := watcher.New()

	// Only handle one file change per event cycle.
	w.SetMaxEvents(1)

	fields["targetDir"] = targetDir
	logging.Logger.WithFields(fields).Debug("Starting Watch.")

	go func() {
		for {
			select {
			case event := <-w.Event:
				logging.Logger.WithFields(fields).Debug("Event Detected.")
				cleanRelativeFilePath := util.FromWindowsPath(strings.Split(event.Path, friendly)[1])
				dirSplit := strings.Split(cleanRelativeFilePath, string(filepath.Separator))
				metadataType, remainder := dirSplit[1], dirSplit[2]
				var changedEntity string
				if metadataType == "componentpacks" {
					changedEntity = filepath.Join(metadataType, remainder)
				} else if metadataType == "site" {
					changedEntity = "site"
				} else {
					changedEntity = filepath.Join(metadataType, strings.Split(remainder, ".")[0])
				}
				logging.Logger.WithFields(fields).Debug("Detected change to metadata type: " + changedEntity)
				go func() {
					if err := pkg.DeployModifiedFiles(auth, targetDir, changedEntity); err != nil {
						w.Error <- err
					}
				}()
			case err := <-w.Error:
				logging.Logger.WithError(err).Fatal("Unable to handle file change.")
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch targetDir recursively for changes.
	if err = w.AddRecursive("."); err != nil {
		return
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	logging.Logger.WithFields(fields).Debug("** Now watching the following files for changes... **")
	for path, f := range w.WatchedFiles() {
		logging.Logger.WithFields(fields).Debug(fmt.Sprintf("%s: %s", path, f.Name()))
	}
	logging.Logger.WithFields(fields).Debug("Waiting for changes...")

	// Start the watching process - it'll check for changes every 100ms.
	if err = w.Start(time.Millisecond * 100); err != nil {
		return
	}

	return
}
