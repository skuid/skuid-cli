package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"encoding/json"

	"github.com/skuid/tides/force"
	"github.com/skuid/tides/types"
	"github.com/spf13/cobra"
)

var noModule bool

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull Skuid Pages from Salesforce into a local directory.",
	Long:  `Pull your Skuid Pages from your Salesforce org to a local directory.`,
	Run: func(cmd *cobra.Command, args []string) {

		//login to the Force API
		api, err := force.Login(
			appClientID,
			appClientSecret,
			host,
			username,
			password,
			apiVersion,
		)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Provide a default for targetDir
		if targetDir == "" {
			targetDir = "skuidpages"
		}

		//build the module and name query paramaters
		query := url.Values{}

		if noModule {
			query.Add("nomodule", strconv.FormatBool(noModule))
		}

		if module != "" {
			query.Add("module", module)
		}

		if page != "" {
			query.Add("page", page)
		}

		//query the API for all pages in the requested module
		result, err := api.Connection.Get("/skuid/api/v1/pages", query)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		var pages map[string]types.PullResponse

		//you have to unquote the string because what comes back
		//is escaped json
		unquoted, _ := strconv.Unquote(string(result))

		//unmarshal all pages into the type PullResponse
		err = json.Unmarshal([]byte(unquoted), &pages)

		if err != nil {
			fmt.Println(string(result))
			fmt.Println(err.Error())
			os.Exit(1)
		}

		for _, pageRecord := range pages {

			//write the page in the at rest format
			err = pageRecord.WriteAtRest(targetDir)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}

		fmt.Printf("Pages written to %s\n", targetDir)
	},
}

func init() {
	pullCmd.Flags().BoolVarP(&noModule, "no-module", "", false, "Retrieve only those pages that do not have a module")
	RootCmd.AddCommand(pullCmd)

}
