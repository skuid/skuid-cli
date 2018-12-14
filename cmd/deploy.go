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
	"time"

	"github.com/skuid/skuid/platform"
	"github.com/skuid/skuid/text"
	"github.com/skuid/skuid/types"
	"github.com/skuid/skuid/ziputils"
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

		plan, err := getDeployPlan(api, bufPlan)
		if err != nil {
			fmt.Println(text.PrettyError("Error getting deploy plan", err))
			os.Exit(1)
		}

		_, err = executeDeployPlan(api, plan)
		if err != nil {
			fmt.Println(text.PrettyError("Error executing deploy plan", err))
			os.Exit(1)
		}

		fmt.Println("Success! Deployed metadata to Skuid Site.")
	},
}

func getDeployPlan(api *platform.RestApi, payload io.Reader) (map[string]types.Plan, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Getting Deploy Plan"))
	}

	planStart := time.Now()
	// Get a deploy plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/deploy/plan",
		payload,
		"application/zip",
	)

	if err != nil {
		return nil, err
	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success Getting Deploy Plan", planStart))
	}

	defer (*planResult).Close()

	var plans map[string]types.Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}

	return plans, nil
}

func executeDeployPlan(api *platform.RestApi, plans map[string]types.Plan) ([]*io.ReadCloser, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Executing Deploy Plan"))
	}
	planResults := []*io.ReadCloser{}
	for _, plan := range plans {
		planResult, err := executePlanItem(api, plan)
		if err != nil {
			return nil, err
		}
		planResults = append(planResults, planResult)
	}
	return planResults, nil
}

func executePlanItem(api *platform.RestApi, plan types.Plan) (*io.ReadCloser, error) {
	// Create a buffer to write our archive to.
	var planResult *io.ReadCloser
	bufDeploy := new(bytes.Buffer)
	err := ziputils.Archive(targetDir, bufDeploy, &plan.Metadata)
	if err != nil {
		log.Print("Error creating deployment ZIP archive")
		log.Fatal(err)
	}

	deployStart := time.Now()

	if plan.Host == "" {
		if verbose {
			fmt.Println(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", plan.URL, plan.Type))
		}
		planResult, err = api.Connection.MakeRequest(
			http.MethodPost,
			plan.URL,
			bufDeploy,
			"application/zip",
		)
		if err != nil {
			return nil, err
		}
	} else {

		url := fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.URL)
		if verbose {
			fmt.Println(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", url, plan.Type))
		}
		planResult, err = api.Connection.MakeJWTRequest(
			http.MethodPost,
			url,
			bufDeploy,
			"application/zip",
		)
		if err != nil {
			return nil, err
		}

	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success Retrieving from Source", deployStart))
	}
	defer (*planResult).Close()
	return planResult, nil

}

func init() {
	RootCmd.AddCommand(deployCmd)
}
