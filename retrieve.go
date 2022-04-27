package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	// jsoniter. Fork of github.com/json-iterator/go
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true, // we do not want to show users raw errors
		Example:       "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval",
		Use:           "retrieve",
		Short:         "Retrieve Skuid metadata from a Site into a local directory.",
		Long:          "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {

			VerboseCommand("Retrieve Metadata")

			api, err := PlatformLogin(cmd)

			retrieveStart := time.Now()

			if err != nil {
				err = fmt.Errorf("Error logging in to Skuid site: %v", err)
				return
			}

			plan, err := GetRetrievePlan(api, ArgAppName)
			if err != nil {
				err = fmt.Errorf("Error getting retrieve plan: %v", err)
				return
			}

			results, err := executeRetrievePlan(api, plan)
			if err != nil {
				PrintError("Error executing retrieve plan", err)
				return
			}

			err = WriteResultsToDisk(results, writeNewFile, createDirectory, readExistingFile)
			if err != nil {
				PrintError("Error writing results to disk", err)
				return
			}

			successMessage := "Successfully retrieved metadata from Skuid Site"

			VerboseSuccess(successMessage, retrieveStart)

			return
		},
	}
)

func init() {
	RootCmd.AddCommand(retrieveCmd)

	AddFlags(retrieveCmd, PlatformLoginFlags...)
}

func GetRetrievePlan(api *PlatformRestApi, appName string) (map[string]Plan, error) {

	VerboseSection("Getting Retrieve Plan")

	var postBody io.Reader
	if appName != "" {
		retFilter, err := json.Marshal(RetrieveFilter{
			AppName: appName,
		})
		if err != nil {
			return nil, err
		}
		postBody = bytes.NewReader(retFilter)
	}

	planStart := time.Now()
	// Get a retrieve plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/retrieve/plan",
		postBody,
		"application/json",
	)

	if err != nil {
		return nil, err
	}

	VerboseSuccess("Success Getting Retrieve Plan", planStart)

	defer (*planResult).Close()

	var plans map[string]Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func executeRetrievePlan(api *PlatformRestApi, plans map[string]Plan) ([]*io.ReadCloser, error) {

	VerboseSection("Executing Retrieve Plan")

	planResults := []*io.ReadCloser{}
	for _, plan := range plans {
		metadataBytes, err := json.Marshal(RetrieveRequest{
			Metadata: plan.Metadata,
			DoZip:    !ArgNoZip,
		})
		if err != nil {
			return nil, err
		}
		retrieveStart := time.Now()
		if plan.Host == "" {

			VerboseLn(fmt.Sprintf("Making Retrieve Request: URL: [%s] Type: [%s]", plan.URL, plan.Type))

			planResult, err := api.Connection.MakeRequest(
				http.MethodPost,
				plan.URL,
				bytes.NewReader(metadataBytes),
				"application/json",
			)
			if err != nil {
				return nil, err
			}
			planResults = append(planResults, planResult)
		} else {
			url := fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.URL)

			VerboseLn(fmt.Sprintf("Making Retrieve Request: URL: [%s] Type: [%s]", url, plan.Type))

			planResult, err := api.Connection.MakeJWTRequest(
				http.MethodPost,
				url,
				bytes.NewReader(metadataBytes),
				"application/json",
			)
			if err != nil {
				return nil, err
			}
			planResults = append(planResults, planResult)
		}

		VerboseSuccess("Success Retrieving from Source", retrieveStart)

	}
	return planResults, nil
}
