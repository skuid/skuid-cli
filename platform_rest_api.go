package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type PlatformRestApi struct {
	Connection *PlatformRestConnection
}

// GetDeployPlan fetches a deploymnent plan from Skuid Platform API
func (api *PlatformRestApi) GetDeployPlan(payload io.Reader, mimeType string, verbose bool) (map[string]Plan, error) {
	if verbose {
		VerboseSection("Getting Deploy Plan")
	}
	if mimeType == "" {
		mimeType = "application/zip"
	}

	planStart := time.Now()
	// Get a deploy plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/deploy/plan",
		payload,
		mimeType,
	)

	if err != nil {
		return nil, err
	}

	if verbose {
		SuccessWithTime("Success Getting Deploy Plan", planStart)
	}

	defer (*planResult).Close()

	var plans map[string]Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}

	return plans, nil
}

// ExecuteDeployPlan executes a map of plan items in a deployment plan
func (api *PlatformRestApi) ExecuteDeployPlan(plans map[string]Plan, targetDir string, verbose bool) ([]*io.ReadCloser, error) {
	if verbose {
		VerboseSection("Executing Deploy Plan")
	}
	planResults := []*io.ReadCloser{}
	for _, plan := range plans {
		planResult, err := api.ExecutePlanItem(plan, targetDir, verbose)
		if err != nil {
			return nil, err
		}
		planResults = append(planResults, planResult)
	}
	return planResults, nil
}

// ExecutePlanItem executes a particular item in a deployment plan
func (api *PlatformRestApi) ExecutePlanItem(plan Plan, targetDir string, verbose bool) (*io.ReadCloser, error) {
	// Create a buffer to write our archive to.
	var planResult *io.ReadCloser
	bufDeploy := new(bytes.Buffer)
	err := Archive(targetDir, bufDeploy, &plan.Metadata)
	if err != nil {
		log.Print("Error creating deployment ZIP archive")
		log.Fatal(err)
	}

	deployStart := time.Now()

	if plan.Host == "" {
		if verbose {
			Println(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", plan.URL, plan.Type))
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
			Println(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", url, plan.Type))
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
		SuccessWithTime("Success Deploying to Source", deployStart)
	}
	defer (*planResult).Close()
	return planResult, nil

}