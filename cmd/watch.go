package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
)

func NewCmdWatch(factory *cmdutil.Factory) *cobra.Command {
	watchTemplate := &cmdutil.CmdTemplate{
		Use:     "watch",
		Short:   "Watch for changes to local Skuid metadata, and deploy changes to a Skuid NLX Site",
		Long:    "Watches for changes to local Skuid metadata on your file system, and automatically deploys the changed files to a Skuid NLX Site",
		Example: "watch -u myUser -p myPassword --host my-site.skuidsite.com --dir ./my-site-objects",
		Flags: &cmdutil.CommandFlags{
			String:         []*flags.Flag[string]{flags.Username, flags.Dir},
			RedactedString: []*flags.Flag[flags.RedactedString]{flags.Password},
			CustomString:   []*flags.Flag[flags.CustomString]{flags.Host},
		},
	}

	return watchTemplate.ToCommand(factory, nil, nil, watch)
}

func watch(factory *cmdutil.Factory, cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["process"] = "watch"
	// get required arguments
	host, err := flags.GetCustomString(cmd.Flags(), flags.Host.Name)
	if err != nil {
		return
	}
	username, err := cmd.Flags().GetString(flags.Username.Name)
	if err != nil {
		return
	}
	password, err := flags.GetRedactedString(cmd.Flags(), flags.Password.Name)
	if err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username

	logging.WithFields(fields).Debug("Gathered Credentials")

	auth, err := pkg.Authorize(host, username, password)
	// we don't need it anymore - very inelegant approach but at least it is something for now
	// Clearing it here instead of in auth package which is the only place its accessed because the tests that exist
	// for auth rely on package global variables so clearing in there would break those tests as they currently exist.
	//
	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	password.Set("")
	if err != nil {
		return
	}

	fields["authorized"] = true
	logging.WithFields(fields).Debug("Successfully Logged In")

	var targetDir string
	if targetDir, err = cmd.Flags().GetString(flags.Dir.Name); err != nil {
		return
	}

	if targetDir, err = util.SanitizePath(targetDir); err != nil {
		return
	}

	fields["targetDirectory"] = targetDir
	logging.WithFields(fields).Debug("Starting Watch")

	// Create our watcher
	w := watcher.New()

	// TODO: Limiting the events could result in file changes being missed.  Unclear why the original
	// code had limit of 1 event for cycle, possibly for perf reasons and/or intentionally only
	// supporting single "file" changes (as opposed to a set of files changing simultaneously with
	// a directory move operation for example).  For now, given the original code had a limit of
	// one (1) event per cycle and its not a common use case to modify multiple files simultaneously,
	// sticking with the limit.  However, if/when Skuid officially documents what is supported
	// functionality & behavior for watch across all possible situations, the limit can be removed/
	// adjusted and thorough testing of handling multiple events performed prior to finalizing.
	w.SetMaxEvents(1)

	// ignore watcher.Remove & watcher.Chmod
	w.FilterOps(watcher.Create, watcher.Write, watcher.Rename, watcher.Move)

	// setup event filter (e.g., ignore directory operations)
	w.AddFilterHook(filterEvents)

	go func() {
		for {
			select {
			case event := <-w.Event:
				logging.WithFields(fields).Debugf("Detected %v operation to file: %v", event.Op, event.Path)
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

func filterEvents(fileInfo os.FileInfo, fullPath string) error {
	// ignore any operations on directories
	if fileInfo.IsDir() {
		return watcher.ErrSkip
	}

	return nil
}
