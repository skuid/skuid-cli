package main

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var noModule bool

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull Skuid Pages from Salesforce into a local directory.",
	Long:  `Pull your Skuid Pages from your Salesforce org to a local directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		//login to the Force API
		api, err := SalesforceLogin(
			ArgAppClientID,
			ArgAppClientSecret,
			ArgHost,
			ArgUsername,
			ArgPassword,
			ArgApiVersion,
		)

		if err != nil {
			Println(err.Error())
			return err
		}

		// Provide a default for targetDir
		if ArgTargetDir == "" {
			ArgTargetDir = "skuidpages"
		}

		//build the module and name query paramaters
		query := url.Values{}

		if noModule {
			query.Add("nomodule", strconv.FormatBool(noModule))
		}

		if ArgModule != "" {
			query.Add("module", ArgModule)
		}

		if ArgPage != "" {
			query.Add("page", ArgPage)
		}

		//query the API for all pages in the requested module
		result, err := api.Connection.Get("/skuid/api/v1/pages", query)

		if err != nil {
			Println(err.Error())
			return err
		}

		var pages map[string]SkuidPullResponse

		//you have to unquote the string because what comes back
		//is escaped json
		unquoted, _ := strconv.Unquote(string(result))

		//unmarshal all pages into the type PullResponse
		err = json.Unmarshal([]byte(unquoted), &pages)

		if err != nil {
			Println(string(result))
			Println(err.Error())
			return err
		}

		for _, pageRecord := range pages {

			//write the page in the at rest format
			err = pageRecord.WriteAtRest(ArgTargetDir)
			if err != nil {
				Println(err.Error())
				return err
			}
		}

		Printf("Pages written to %s\n", ArgTargetDir)

		return nil
	},
}

func init() {
	pullCmd.Flags().BoolVarP(&noModule, "no-module", "", false, "Retrieve only those pages that do not have a module")
	RootCmd.AddCommand(pullCmd)

}