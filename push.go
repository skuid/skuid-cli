package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var f string

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push Skuid Pages from a directory to Skuid.",
	Long:  "Push Skuid Pages from a directory to Skuid.",
	Run: func(cmd *cobra.Command, args []string) {

		pageDefinitions, _ := ReadFiles(ArgTargetDir, ArgModule, f)

		pagePost := &PagePost{Changes: pageDefinitions}

		fmt.Println(fmt.Sprintf("Pushing %d pages.", len(pagePost.Changes)))

		api, err := Login(
			ArgAppClientID,
			ArgAppClientSecret,
			ArgHost,
			ArgUsername,
			ArgPassword,
			ArgApiVersion,
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

		response := &PagePostResult{}

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
	pushCmd.Flags().StringVarP(&f, "file", "f", "", "Skuid Page file(s) to push. Supports file globs.")
}
