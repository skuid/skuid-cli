package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gookit/color"
	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

type watchCommander struct {
	factory  *cmdutil.Factory
	authOpts pkg.AuthorizeOptions
	dir      string
}

func (c *watchCommander) GetCommand() *cobra.Command {
	template := &cmdutil.CmdTemplate{
		Use:     "watch",
		Short:   "Watch for changes to local Skuid metadata, and deploy changes to a Skuid NLX Site",
		Long:    "Watches for changes to local Skuid metadata on your file system, and automatically deploys the changed files to a Skuid NLX Site",
		Example: "watch -u myUser -p myPassword --host my-site.skuidsite.com --dir ./my-site-objects",
	}
	cmd := template.ToCommand(c.watch)

	cmdutil.AddAuthFlags(cmd, &c.authOpts)
	cmdutil.AddStringFlag(cmd, &c.dir, flags.Dir)

	return cmd
}

func NewCmdWatch(factory *cmdutil.Factory) *cobra.Command {
	commander := new(watchCommander)
	commander.factory = factory
	return commander.GetCommand()
}

func (c *watchCommander) watch(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["start"] = time.Now()
	fields["process"] = "watch"
	logging.WithFields(fields).Info(color.Green.Sprint("Starting Watch"))

	fields["host"] = c.authOpts.Host
	fields["username"] = c.authOpts.Username
	logging.WithFields(fields).Debug("Gathered Credentials")

	auth, err := pkg.Authorize(&c.authOpts)
	// we don't need it anymore - very inelegant approach but at least it is something for now
	// Clearing it here instead of in auth package which is the only place its accessed because the tests that exist
	// for auth rely on package global variables so clearing in there would break those tests as they currently exist.
	//
	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	// intentionally ignoring error since there is nothing we can do and we should fail entirely as a result
	_ = c.authOpts.Password.Set("")
	if err != nil {
		return
	}

	fields["authorized"] = true
	logging.WithFields(fields).Info("Authentication Successful")

	var targetDirectory string
	if targetDirectory, err = filepath.Abs(c.dir); err != nil {
		return
	}

	fields["targetDirectory"] = targetDirectory
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

	// setup event filter (e.g., ignore directory operations, non-metadata directory file operations)
	w.AddFilterHook(filterEvents(targetDirectory))

	go func() {
		for {
			select {
			case event := <-w.Event:
				logging.WithFields(fields).Debugf("Detected %v operation to file: %q", event.Op, event.Path)
				go func() {
					if err := handleEvent(auth, targetDirectory, event); err != nil {
						logging.Get().Errorf("unable to handle %v operation to file %q: %v", event.Op, event.Path, err)
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
	if err = w.AddRecursive(targetDirectory); err != nil {
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

func handleEvent(auth *pkg.Authorization, targetDirectory string, event watcher.Event) error {
	relativeFilePath, err := filepath.Rel(targetDirectory, event.Path)
	if err != nil {
		return err
	}
	entity, err := metadata.NewMetadataEntityFile(relativeFilePath)
	if err != nil {
		return err
	}
	entitiesToArchive := []metadata.MetadataEntity{entity.Entity}
	archiveFilter := pkg.MetadataEntityArchiveFilter(entitiesToArchive)
	return pkg.Deploy(auth, targetDirectory, archiveFilter, entitiesToArchive, nil)
}

func filterEvents(targetDirectory string) watcher.FilterFileHookFunc {
	return func(fileInfo os.FileInfo, fullPath string) error {
		// ignore any operations on directories
		if fileInfo.IsDir() {
			return watcher.ErrSkip
		}

		// ignore any operations on files in non-metadata directories - We only want to monitor metadata directories, however
		// we must configure AddRecursive to monitor the entire target directory instead of separate AddRecursive calls for
		// each metadata directory because Add/AddRecursive requires that directories exist when called and we need to handle
		// the situation where a metadata directory is created while watch is running. When evaluating events for filtering,
		// we only evaluate that the event is coming from a metadata folder, we do not evaluate the validity of the file itself
		// here, leaving that instead to the processing of the event because we want to notify the user of success or failure
		// of all events coming from a metaadata folder. For example, if the target directory is skuid-objects and the event is
		// for skuid-objects/README.md, we filter it since its not in a metadata directory.  However, if the event is for
		// skuid-objects/pages/README.md, we do not filter it since its in a known metadata directory even though the file itself
		// is not valid for the metadata type.
		if relPath, err := filepath.Rel(targetDirectory, fullPath); err != nil {
			return err
		} else if !metadata.IsMetadataTypePath(relPath) {
			return watcher.ErrSkip
		}

		return nil
	}
}
