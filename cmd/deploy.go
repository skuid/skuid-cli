package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/validation"
	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var deployCmd = &cobra.Command{
	SilenceErrors:     true,
	SilenceUsage:      true,
	Use:               "deploy",
	Short:             "Deploy local Skuid metadata to a Skuid Platform Site.",
	Long:              "Deploy Skuid metadata stored within a local file system directory to a Skuid Platform Site.",
	PersistentPreRunE: validation.PrerunValidation,
	RunE: func(cmd *cobra.Command, _ []string) (err error) {

		logging.VerboseCommand("Deploy Metadata")

		api, err := pkg.SkuidNlxLogin(cmd)

		if err != nil {
			logging.PrintError("Error logging in to Skuid site", err)
			return
		}

		deployStart := time.Now()

		var currDir string

		currentDirectory, err := os.Getwd()
		if err != nil {
			logging.PrintError("Unable to get working directory", err)
			return
		}

		defer func() {
			err := os.Chdir(currentDirectory)
			if err != nil {
				logging.PrintError("Unable to change directory", err)
				panic(err)
			}
		}()

		var targetDir string
		if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
			return
		}

		// If target directory is provided,
		// switch to that target directory and later switch back.
		if targetDir != "" {
			err = os.Chdir(targetDir)
			if err != nil {
				logging.PrintError("Unable to change working directory", err)
				return
			}
		}

		dotDir := "."
		currDir, err = filepath.Abs(filepath.Dir(dotDir))
		if err != nil {
			logging.PrintError("Unable to form filepath", err)
			return
		}

		logging.VerboseLn("Deploying site from", currDir)

		// Create a buffer to write our archive to.
		bufPlan := new(bytes.Buffer)
		err = pkg.Archive(currDir, bufPlan, nil)
		if err != nil {
			logging.PrintError("Error creating deployment ZIP archive", err)
			return
		}

		var deployPlan io.Reader
		mimeType := "application/zip"

		var appName string
		if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
			return
		}

		if appName != "" {
			filter := pkg.DeployFilter{
				AppName: appName,
				Plan:    bufPlan.Bytes(),
			}
			deployBytes, err := json.Marshal(filter)
			if err != nil {
				logging.PrintError("Error creating deployment plan payload", err)
				return err
			}
			deployPlan = bytes.NewReader(deployBytes)
			mimeType = "application/json"
		} else {
			deployPlan = bufPlan
		}

		plan, err := api.GetDeployPlan(deployPlan, mimeType)
		if err != nil {
			logging.PrintError("Error getting deploy plan", err)
			return
		}

		for _, service := range plan {
			if service.Warnings != nil {
				for _, warning := range service.Warnings {
					logging.Println(warning)
				}
			}
		}

		_, err = api.ExecuteDeployPlan(plan, dotDir)
		if err != nil {
			logging.PrintError("Error executing deploy plan", err)
			return
		}

		successMessage := "Successfully deployed metadata to Skuid Site"

		logging.SuccessWithTime(successMessage, deployStart)

		return
	},
}

func init() {
	TidesCmd.AddCommand(deployCmd)
	flags.AddFlags(deployCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(deployCmd, flags.Directory, flags.AppName, flags.ApiVersion)
}
