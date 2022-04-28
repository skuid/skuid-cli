package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	// jsoniter. Fork of github.com/json-iterator/go
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:      true,
		SilenceErrors:     true, // we do not want to show users raw errors
		Example:           "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval",
		Use:               "retrieve",
		Short:             "Retrieve Skuid metadata from a Site into a local directory.",
		Long:              "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
		PersistentPreRunE: PrerunValidation,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {

			logging.VerboseCommand("Retrieve Metadata")

			api, err := SkuidNlxLogin(cmd)

			retrieveStart := time.Now()

			if err != nil {
				err = fmt.Errorf("Error logging in to Skuid site: %v", err)
				return
			}

			var appName string
			if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
				return
			}

			plan, err := GetRetrievePlan(api, appName)
			if err != nil {
				err = fmt.Errorf("Error getting retrieve plan: %v", err)
				return
			}

			var noZip bool
			if noZip, err = cmd.Flags().GetBool(flags.NoZip.Name); err != nil {
				return
			}

			results, err := executeRetrievePlan(api, plan, noZip)
			if err != nil {
				err = fmt.Errorf("Error executing retrieve plan: %v", err)
				return
			}

			var targetDir string
			if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
				return
			}

			err = WriteResultsToDisk(targetDir, results, writeNewFile, createDirectory, readExistingFile, noZip)
			if err != nil {
				err = fmt.Errorf("Error writing results to disk: %v", err)
				return
			}

			successMessage := "Successfully retrieved metadata from Skuid Site"

			logging.VerboseSuccess(successMessage, retrieveStart)

			return
		},
	}
)

func init() {
	RootCmd.AddCommand(retrieveCmd)

	flags.AddFlagFunctions(retrieveCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName)
	flags.AddFlags(retrieveCmd, flags.NoZip)
}

func GetRetrievePlan(api *NlxApi, appName string) (map[string]Plan, error) {

	logging.VerboseSection("Getting Retrieve Plan")

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

	logging.VerboseSuccess("Success Getting Retrieve Plan", planStart)

	defer (*planResult).Close()

	var plans map[string]Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func executeRetrievePlan(api *NlxApi, plans map[string]Plan, noZip bool) (planResults []*io.ReadCloser, err error) {

	logging.VerboseSection("Executing Retrieve Plan")

	for _, plan := range plans {
		metadataBytes, err := json.Marshal(RetrieveRequest{
			Metadata: plan.Metadata,
			DoZip:    !noZip,
		})
		if err != nil {
			return nil, err
		}
		retrieveStart := time.Now()
		if plan.Host == "" {

			logging.VerboseF("Making Retrieve Request: URL: [%s] Type: [%s]\n", plan.URL, plan.Type)

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

			logging.VerboseF("Making Retrieve Request: URL: [%s] Type: [%s]\n", url, plan.Type)

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

		logging.VerboseSuccess("Success Retrieving from Source", retrieveStart)

	}
	return planResults, nil
}
