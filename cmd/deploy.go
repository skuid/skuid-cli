package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/skuid/skuid-cli/platform"
	"github.com/skuid/skuid-cli/text"
	"github.com/skuid/skuid-cli/ziputils"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy local Skuid metadata to a Skuid Platform Site.",
	Long:  "Deploy Skuid metadata stored within a local file system directory to a Skuid Platform Site.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(text.RunCommand("Deploy Metadata"))

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			metadataServiceProxy,
			dataServiceProxy,
			verbose,
		)

		if err != nil {
			fmt.Println(text.PrettyError("Error logging in to Skuid site", err))
			os.Exit(1)
		}

		deployStart := time.Now()

		var targetDirFriendly string

		// If target directory is provided,
		// switch to that target directory and later switch back.
		if targetDir != "" {
			os.Chdir(targetDir)
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			defer os.Chdir(pwd)
		}
		targetDir = "."
		targetDirFriendly, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}

		if verbose {
			fmt.Println("Deploying site from", targetDirFriendly)
		}

		// Create a buffer to write our archive to.
		bufPlan := new(bytes.Buffer)
		err = ziputils.Archive(targetDir, bufPlan, nil)
		if err != nil {
			fmt.Println(text.PrettyError("Error creating deployment ZIP archive", err))
			os.Exit(1)
		}

		plan, err := api.GetDeployPlan(bufPlan, verbose)
		if err != nil {
			fmt.Println(text.PrettyError("Error getting deploy plan", err))
			os.Exit(1)
		}

		for _, service := range plan {
			if service.Warnings != nil {
				for _, warning := range service.Warnings {
					fmt.Println(warning)
				}
			}
		}

		_, err = api.ExecuteDeployPlan(plan, targetDir, verbose)
		if err != nil {
			fmt.Println(text.PrettyError("Error executing deploy plan", err))
			os.Exit(1)
		}

		successMessage := "Successfully deployed metadata to Skuid Site"

		if verbose {
			fmt.Println(text.SuccessWithTime(successMessage, deployStart))
		} else {
			fmt.Println(successMessage + ".")
		}

	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
