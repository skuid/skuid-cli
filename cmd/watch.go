package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
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
	cmdutil.AddValueFlag(cmd, &c.dir, flags.Dir)

	return cmd
}

func NewCmdWatch(factory *cmdutil.Factory) *cobra.Command {
	commander := new(watchCommander)
	commander.factory = factory
	return commander.GetCommand()
}

func (c *watchCommander) watch(cmd *cobra.Command, args []string) (err error) {
	message := fmt.Sprintf("Executing command %v entities for site %v in directory %v", logging.QuoteText(cmd.Name()), logging.ColorResource.Text(c.authOpts.Host), logging.ColorResource.QuoteText(c.dir))
	fields := logging.Fields{
		logging.CommandNameKey: cmd.Name(),
		"host":                 c.authOpts.Host,
		"username":             c.authOpts.Username,
		"sourceDirectory":      c.dir,
	}
	logger := logging.WithTracking("cmd.watch", message, fields).StartTracking()
	defer func() {
		err = cmdutil.CheckError(cmd, err, recover())
		logger = logger.FinishTracking(err)
		err = cmdutil.HandleCommandResult(cmd, logger, err, fmt.Sprintf("%v command completed", logging.QuoteText(cmd.Name())))
	}()

	auth, err := pkg.AuthorizeOnce(&c.authOpts)
	if err != nil {
		return err
	}

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

	// Skuid Review Required - Should we ignore Rename & Move?  Depending on answers to open
	// issues in the repro around how "name" field in definition file is used and if it must
	// match the filename for CLI to work (or the CLI needs to change to accomodate), renaming/
	// moving may be impossible to manage anyway unless modifying the name of the json to match
	// the new file name/location (which carries further issues if it changes metadata dir
	// locations).
	// see https://github.com/skuid/skuid-cli/issues/184 & https://github.com/skuid/skuid-cli/issues/203
	//
	// ignore watcher.Remove & watcher.Chmod
	w.FilterOps(watcher.Create, watcher.Write, watcher.Rename, watcher.Move)

	// setup event filter (e.g., ignore directory operations, non-metadata directory file operations)
	w.AddFilterHook(filterEvents(c.dir))

	// watch source directory recursively for changes.
	if err = w.AddRecursive(c.dir); err != nil {
		return err
	}

	// Print a list of all of the files and folders currently being watched and their
	// paths noting that this is just a current snapshot.  Watcher polls based on
	// polling interval specified so the files being watch can change if the user
	// adds/removes files.  The key is that its looking for any change under the
	// sourceDirectory that is within a metadata directory so we won't miss anything
	// new that's added (we ignore deletes) in the filter we pass to watcher).
	logger.Debugf("Current files within directory %v", logging.ColorResource.QuoteText(c.dir))
	for path := range w.WatchedFiles() {
		logger.Debug(fmt.Sprintf("%v", logging.ColorResource.QuoteText(path)))
	}

	// if w.Start() fails, the listeners will be running so need to notify them to exit
	abort := make(chan struct{}, 1)
	defer close(abort)
	// setup listeners receiving channel for notification when listener has exited
	watcherDone := listenForEvents(cmd, w, abort, auth, c.dir, logger)

	// Start the watching process - it'll check for changes every 100ms.
	logger.Infof("Watching directory %v", logging.ColorResource.QuoteText(c.dir))
	logger.Info("Ctrl+C to cancel")
	if err := w.Start(time.Millisecond * 100); err != nil {
		return err
	}

	if err := <-watcherDone; err != nil {
		return err
	}

	logger = logger.WithSuccess()
	return nil
}

func listenForEvents(cmd *cobra.Command, w *watcher.Watcher, abort <-chan struct{}, auth *pkg.Authorization, sourceDirectory string, logger *logging.Logger) <-chan error {
	done := make(chan error, 1)       // notification to send back to calling function
	term := make(chan error, 1)       // notification from watcher polling go routine
	cancel := make(chan os.Signal, 1) // signal notification (e.g., Ctrl+C, Ctrl+Z)
	quit := make(chan struct{}, 1)    // notification that w.Close() has completed
	signal.Notify(cancel, os.Interrupt, syscall.SIGTERM, syscall.SIGTSTP)

	go func() {
		defer func() {
			// nothing further we can do to clean-up gracefully
			logger.FatalGoRoutineCondition(recover(), logging.QuoteText(cmd.Name()))
			close(quit)
			close(cancel)
			close(term)
			close(done)
		}()

		// Must drain the queue in a different go routine than the one calling Close() due
		// to watcher deadlock bug (see https://github.com/radovskyb/watcher/issues/72).
		// We drain here instead of in the polling go routine below because in a panic
		// situation our listening loop will have exited unexpectedly so handle both
		// expected and unexpected scenarios in one place
		drainQueue := func() {
			go func() {
				defer func() { logger.FatalGoRoutineCondition(recover(), "draining watcher queue") }()
				for {
					select {
					case wErr := <-w.Error:
						logger.Tracef("Skipping processing of watcher error because in the processing of closing: %v", wErr)
					case event := <-w.Event:
						logger.Tracef("Skipping processing of watcher %v event for file %v because in the processing of closing", event.Op, logging.QuoteText(event.Path))
					case <-w.Closed:
						logger.Tracef("Skipping processing of watcher closed because in the processing of closing")
					case <-quit:
						return
					}
				}
			}()
		}

		// wait for watcher polling to complete
		wErr := <-term
		drainQueue()
		w.Close()
		quit <- struct{}{}
		done <- wErr
	}()

	go func() {
		term <- func() (err error) {
			defer func() {
				err = cmdutil.CheckError(cmd, err, recover())
			}()

			logWaitMsg := func() { logger.Info("Waiting for changes...") }
			logWaitMsg()

			for {
				select {
				case event := <-w.Event:
					go func(event watcher.Event) {
						defer func() {
							logger.FatalGoRoutineConditionf(recover(), "deploying entity for file %v", logging.QuoteText(event.Path))
						}()
						if _, err := handleEvent(auth, sourceDirectory, event, logger); err != nil {
							logger.WithError(err).Error(logging.ColorFailure.Sprintf("%v deploying entity for file %v: %v", logging.FailureIcon, logging.QuoteText(event.Path), err))
						}
						logWaitMsg()
					}(event)
				case wErr := <-w.Error:
					return wErr
				case <-w.Closed: // w.Close() was called
					return nil
				case <-cancel: // Signal was sent
					return nil
				case <-abort: // calling function wants us to quit
					return nil
				}
			}
		}()
	}()

	return done
}

func handleEvent(auth *pkg.Authorization, sourceDirectory string, event watcher.Event, logger *logging.Logger) (mdEntityFile *metadata.MetadataEntityFile, err error) {
	message := fmt.Sprintf("Handling %v operation to file: %v", event.Op, logging.ColorResource.QuoteText(event.Path))
	fields := logging.Fields{}
	logger = logger.WithTraceTracking("cmd.handleEvent", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	relativeFilePath, err := filepath.Rel(sourceDirectory, event.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve path to file %v: %w", logging.QuoteText(event.Path), err)
	}
	ef, err := metadata.NewMetadataEntityFile(relativeFilePath)
	if err != nil {
		return nil, err
	}
	entitiesToArchive := []metadata.MetadataEntity{ef.Entity}
	archiveFilter := pkg.MetadataEntityArchiveFilter(entitiesToArchive)
	options := pkg.DeployOptions{
		ArchiveFilter:     archiveFilter,
		Auth:              auth,
		EntitiesToArchive: entitiesToArchive,
		SourceDirectory:   sourceDirectory,
		PlanFilter:        nil,
	}
	err = pkg.Deploy(options)
	if err != nil {
		return nil, err
	}

	logger = logger.WithSuccess()
	return ef, nil
}

func filterEvents(sourceDirectory string) watcher.FilterFileHookFunc {
	return func(fileInfo os.FileInfo, fullPath string) error {
		// ignore any operations on directories
		if fileInfo.IsDir() {
			return watcher.ErrSkip
		}

		// ignore any operations on files in non-metadata directories - We only want to monitor metadata directories, however
		// we must configure AddRecursive to monitor the entire source directory instead of separate AddRecursive calls for
		// each metadata directory because Add/AddRecursive requires that directories exist when called and we need to handle
		// the situation where a metadata directory is created while watch is running. When evaluating events for filtering,
		// we only evaluate that the event is coming from a metadata folder, we do not evaluate the validity of the file itself
		// here, leaving that instead to the processing of the event because we want to notify the user of success or failure
		// of all events coming from a metaadata folder. For example, if the source directory is skuid-objects and the event is
		// for skuid-objects/README.md, we filter it since its not in a metadata directory.  However, if the event is for
		// skuid-objects/pages/README.md, we do not filter it since its in a known metadata directory even though the file itself
		// is not valid for the metadata type.
		if relPath, err := filepath.Rel(sourceDirectory, fullPath); err != nil {
			return err
		} else if !metadata.IsMetadataTypePath(relPath) {
			return watcher.ErrSkip
		}

		return nil
	}
}
