package pkg

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for changes to local Skuid metadata, and deploy changes to a Skuid Platform Site.",
	Long:  "Watches for changes to local Skuid metadata on your file system, and automatically deploys the changed files to a Skuid Platform Site.",
	RunE: func(cmd *cobra.Command, _ []string) (err error) {

		api, err := PlatformLogin(cmd)

		if err != nil {
			err = fmt.Errorf("Error logging in to Skuid site: %v", err)
			return
		}

		var targetDirFriendly string

		// If target directory is provided,
		var back string
		back, err = os.Getwd()
		if err != nil {
			return
		}

		var targetDir string
		if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
			return
		}

		// change directory
		if targetDir != "" {
			err = os.Chdir(targetDir)
			if err != nil {
				return
			}
		}

		// switch back to directory
		defer func() {
			if targetDir != "" {
				if err := os.Chdir(back); err != nil {
					log.Fatalf("failed changing back to directory: %v", err)
				}
			}
		}()

		if targetDir == "" {
			targetDir = "."
		}

		targetDirFriendly, err = filepath.Abs(filepath.Dir(os.Args[0]))

		if err != nil {
			return
		}

		logging.Println("Watching for changes to Skuid metadata files in directory: " + targetDirFriendly)

		// Create our watcher
		w := watcher.New()

		// Only handle one file change per event cycle.
		w.SetMaxEvents(1)

		go func() {
			for {
				select {
				case event := <-w.Event:
					cleanRelativeFilePath := FromWindowsPath(strings.Split(event.Path, targetDirFriendly)[1])
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
					logging.Println("Detected change to metadata type: " + changedEntity)
					go func() {
						if err := deployModifiedFiles(api, targetDir, changedEntity); err != nil {
							w.Error <- err
						}
					}()
				case err := <-w.Error:
					log.Fatalln(err)
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

		logging.VerboseLn("** Now watching the following files for changes... **")
		for path, f := range w.WatchedFiles() {
			logging.VerboseLn(fmt.Sprintf("%s: %s", path, f.Name()))
		}
		logging.VerboseLn("Waiting for changes...")

		// Start the watching process - it'll check for changes every 100ms.
		if err = w.Start(time.Millisecond * 100); err != nil {
			return
		}

		return
	},
}

func deployModifiedFiles(api *PlatformRestApi, targetDir, modifiedFile string) (err error) {

	// Create a buffer to write our archive to.
	bufPlan := new(bytes.Buffer)
	err = ArchivePartial(targetDir, bufPlan, modifiedFile)
	if err != nil {
		err = fmt.Errorf("Error creating deployment ZIP archive: %v", err)
		return
	}

	logging.VerboseLn("Getting deploy plan...")

	plan, err := api.GetDeployPlan(bufPlan, "application/zip")
	if err != nil {
		err = fmt.Errorf("Error getting deploy plan: %v", err)
		return
	}

	logging.VerboseLn("Retrieved deploy plan. Deploying...")

	_, err = api.ExecuteDeployPlan(plan, targetDir)
	if err != nil {
		err = fmt.Errorf("Error executing deploy plan: %v", err)
		return
	}

	successMessage := "Successfully deployed metadata to Skuid Site: " + modifiedFile
	logging.Println(successMessage)

	return
}

func init() {
	RootCmd.AddCommand(watchCmd)
	flags.AddFlagFunctions(watchCmd, flags.PlatformLoginFlags...)
}
