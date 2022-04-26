package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	ArgPushFile string

	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push Skuid Pages from a directory to Skuid.",
		Long:  "Push Skuid Pages from a directory to Skuid.",
		RunE: func(_ *cobra.Command, _ []string) (err error) {

			pageDefinitions, err := ReadFiles(ArgTargetDir, ArgModule, ArgPushFile)
			if err != nil {
				return err
			}

			pagePost := &SkuidPagePost{Changes: pageDefinitions}

			Printf("Pushing %d pages.\n", len(pagePost.Changes))

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
				return
			}

			result, err := api.Connection.Post("/skuid/api/v1/pages", pagePost)

			if err != nil {
				Println(err.Error())
				return
			}

			unquoted, _ := strconv.Unquote(string(result))

			response := &SkuidPagePostResult{}

			_ = json.Unmarshal([]byte(unquoted), response)

			if response.Success == false && len(response.Errors) > 0 {
				Println("There were errors pushing Skuid Pages...")
				for _, err := range response.Errors {
					Println(err)
				}
				return
			}

			Println(fmt.Sprintf("Pages successfully pushed to org: %s.", response.OrgName))

			return
		},
	}
)

func init() {
	RootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&ArgPushFile, "file", "f", "", "Skuid Page file(s) to push. Supports file globs.")
}
