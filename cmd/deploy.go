package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skuid/skuid/platform"
	"github.com/skuid/skuid/types"
	"github.com/skuid/skuid/ziputils"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy local Skuid metadata to a Skuid Platform Site.",
	Long:  "Deploy Skuid metadata stored within a local file system directory to a Skuid Platform Site.",
	Run: func(cmd *cobra.Command, args []string) {

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			verbose,
		)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

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
			log.Print("Error creating deployment ZIP archive")
			log.Fatal(err)
		}

		plan, err := getDeployPlan(api, bufPlan)
		if err != nil {
			fmt.Println("Error getting deploy plan: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Deploying metadata...")

		_, err = executeDeployPlan(api, plan)
		if err != nil {
			fmt.Println("Error executing deploy plan: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully deployed metadata to Skuid Site.")
	},
}

func getDeployPlan(api *platform.RestApi, payload io.Reader) (map[string]types.Plan, error) {
	// Get a retrieve plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/deploy/plan",
		payload,
		"application/zip",
	)
	defer planResult.Close()

	if err != nil {
		return nil, err
	}

	var plans map[string]types.Plan
	if err := json.NewDecoder(planResult).Decode(&plans); err != nil {
		return nil, err
	}

	return plans, nil
}

func executeDeployPlan(api *platform.RestApi, plans map[string]types.Plan) ([]io.ReadCloser, error) {
	planResults := []io.ReadCloser{}
	for _, plan := range plans {
		// Create a buffer to write our archive to.
		bufDeploy := new(bytes.Buffer)
		err := ziputils.Archive(targetDir, bufDeploy, &plan.Metadata)
		if err != nil {
			log.Print("Error creating deployment ZIP archive")
			log.Fatal(err)
		}

		if plan.Host == "" {
			planResult, err := api.Connection.MakeRequest(
				http.MethodPost,
				plan.URL,
				bufDeploy,
				"application/zip",
			)
			if err != nil {
				return nil, err
			}
			defer planResult.Close()
			planResults = append(planResults, planResult)
		} else {

			url := fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.URL)
			planResult, err := api.Connection.MakeJWTRequest(
				http.MethodPost,
				url,
				bufDeploy,
				"application/zip",
			)
			if err != nil {
				return nil, err
			}
			defer planResult.Close()
			planResults = append(planResults, planResult)

		}
	}
	return planResults, nil
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
