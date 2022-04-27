package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy local Skuid metadata to a Skuid Platform Site.",
	Long:  "Deploy Skuid metadata stored within a local file system directory to a Skuid Platform Site.",
	RunE: func(cmd *cobra.Command, _ []string) (err error) {

		VerboseCommand("Deploy Metadata")

		var verbose bool
		if verbose, err = cmd.Flags().GetBool(FlagNameVerbose); err != nil {
			return
		}

		api, err := PlatformLogin(cmd)

		if err != nil {
			PrintError("Error logging in to Skuid site", err)
			return
		}

		deployStart := time.Now()

		var currDir string

		currentDirectory, err := os.Getwd()
		if err != nil {
			PrintError("Unable to get working directory", err)
			return
		}

		defer func() {
			err := os.Chdir(currentDirectory)
			if err != nil {
				PrintError("Unable to change directory", err)
				log.Fatal(err)
			}
		}()

		// If target directory is provided,
		// switch to that target directory and later switch back.
		if ArgTargetDir != "" {
			err = os.Chdir(ArgTargetDir)
			if err != nil {
				PrintError("Unable to change working directory", err)
				return
			}
		}

		dotDir := "."
		currDir, err = filepath.Abs(filepath.Dir(dotDir))
		if err != nil {
			PrintError("Unable to form filepath", err)
			return
		}

		VerboseLn("Deploying site from", currDir)

		// Create a buffer to write our archive to.
		bufPlan := new(bytes.Buffer)
		err = Archive(".", bufPlan, nil)
		if err != nil {
			PrintError("Error creating deployment ZIP archive", err)
			return
		}

		var deployPlan io.Reader
		mimeType := "application/zip"
		if ArgAppName != "" {
			filter := DeployFilter{
				AppName: ArgAppName,
				Plan:    bufPlan.Bytes(),
			}
			deployBytes, err := json.Marshal(filter)
			if err != nil {
				PrintError("Error creating deployment plan payload", err)
				return err
			}
			deployPlan = bytes.NewReader(deployBytes)
			mimeType = "application/json"
		} else {
			deployPlan = bufPlan
		}

		plan, err := api.GetDeployPlan(deployPlan, mimeType, verbose)
		if err != nil {
			PrintError("Error getting deploy plan", err)
			return
		}

		for _, service := range plan {
			if service.Warnings != nil {
				for _, warning := range service.Warnings {
					Println(warning)
				}
			}
		}

		_, err = api.ExecuteDeployPlan(plan, dotDir, verbose)
		if err != nil {
			PrintError("Error executing deploy plan", err)
			return
		}

		successMessage := "Successfully deployed metadata to Skuid Site"

		SuccessWithTime(successMessage, deployStart)

		return
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	AddFlags(deployCmd, PlatformLoginFlags...)
}
