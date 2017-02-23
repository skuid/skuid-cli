package cmd

import (
	"fmt"
	"os"
	"strconv"

	"encoding/json"

	"github.com/skuid/skuid/force"
	"github.com/skuid/skuid/types"
	"github.com/spf13/cobra"
)

var inputDir string
var desiredModule string
var f string

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push Skuid Pages from a directory to Skuid.",
	Long:  "Push Skuid Pages from a directory to Skuid.",
	Run: func(cmd *cobra.Command, args []string) {

		pageDefinitions, _ := types.ReadFiles(inputDir, desiredModule, f)

		pagePost := &types.PagePost{Changes: pageDefinitions}

		fmt.Println(fmt.Sprintf("Pushing %d pages.", len(pagePost.Changes)))

		api, err := force.Login(
			appClientID,
			appClientSecret,
			sfdcHost,
			sfdcUsername,
			sfdcPassword,
			apiVersion,
		)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		result, err := api.Connection.Post("/skuid/api/v1/pages", pagePost)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		unquoted, _ := strconv.Unquote(string(result))

		response := &types.PagePostResult{}

		_ = json.Unmarshal([]byte(unquoted), response)

		if response.Success == false && len(response.Errors) > 0 {
			fmt.Println("There were errors pushing Skuid Pages...")
			for _, err := range response.Errors {
				fmt.Println(err)
			}
			os.Exit(1)
		}

		fmt.Println(fmt.Sprintf("Pages successfully pushed to org: %s.", response.OrgName))

	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&inputDir, "dir", "d", "skuidpages", "Directory where Skuid Pages live.")
	pushCmd.Flags().StringVarP(&desiredModule, "module", "m", "", "Module to push. (Only supports single module)")
	pushCmd.Flags().StringVarP(&f, "file", "f", "", "Skuid Page to Push. Supports file globs.")
}
