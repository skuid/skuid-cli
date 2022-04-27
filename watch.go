package main

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
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for changes to local Skuid metadata, and deploy changes to a Skuid Platform Site.",
	Long:  "Watches for changes to local Skuid metadata on your file system, and automatically deploys the changed files to a Skuid Platform Site.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		api, err := PlatformLogin(cmd)

		if err != nil {
			PrintError("Error logging in to Skuid site", err)
			return
		}

		var targetDirFriendly string

		// If target directory is provided,
		// switch to that target directory and later switch back.
		if ArgTargetDir != "" {
			os.Chdir(ArgTargetDir)
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			defer os.Chdir(pwd)
		}
		ArgTargetDir = "."
		targetDirFriendly, err = filepath.Abs(filepath.Dir(os.Args[0]))

		if err != nil {
			return
		}

		Println("Watching for changes to Skuid metadata files in directory: " + targetDirFriendly)

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
					Println("Detected change to metadata type: " + changedEntity)
					go deployModifiedFiles(api, changedEntity)
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		// Watch targetDir recursively for changes.
		if err = w.AddRecursive(ArgTargetDir); err != nil {
			return
		}

		// Print a list of all of the files and folders currently
		// being watched and their paths.

		VerboseLn("** Now watching the following files for changes... **")
		for path, f := range w.WatchedFiles() {
			VerboseLn(fmt.Sprintf("%s: %s", path, f.Name()))
		}
		VerboseLn("Waiting for changes...")

		// Start the watching process - it'll check for changes every 100ms.
		if err = w.Start(time.Millisecond * 100); err != nil {
			return
		}

		return
	},
}

func deployModifiedFiles(api *PlatformRestApi, modifiedFile string) {

	// Create a buffer to write our archive to.
	bufPlan := new(bytes.Buffer)
	err := ArchivePartial(ArgTargetDir, bufPlan, modifiedFile)
	if err != nil {
		PrintError("Error creating deployment ZIP archive", err)
		os.Exit(1)
	}

	VerboseLn("Getting deploy plan...")

	plan, err := api.GetDeployPlan(bufPlan, "application/zip", GlobalArgVerbose)
	if err != nil {
		PrintError("Error getting deploy plan", err)
		os.Exit(1)
	}

	VerboseLn("Retrieved deploy plan. Deploying...")

	_, err = api.ExecuteDeployPlan(plan, ArgTargetDir, GlobalArgVerbose)
	if err != nil {
		PrintError("Error executing deploy plan", err)
		os.Exit(1)
	}

	successMessage := "Successfully deployed metadata to Skuid Site: " + modifiedFile
	Println(successMessage)
}

func init() {
	RootCmd.AddCommand(watchCmd)
	AddFlags(watchCmd, PlatformLoginFlags...)
}
