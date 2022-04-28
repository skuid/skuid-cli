package pkg

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push Skuid Pages from a directory to Skuid.",
		Long:  "Push Skuid Pages from a directory to Skuid.",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			var targetDir, pushFile string
			var modules []string

			if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
				return
			}

			if modules, err = cmd.Flags().GetStringArray(flags.Modules.Name); err != nil {
				return
			}

			if pushFile, err = cmd.Flags().GetString(flags.PushFile.Name); err != nil {
				return
			}

			pageDefinitions, err := ReadFiles(targetDir, modules, pushFile)
			if err != nil {
				return
			}

			pagePost := &SkuidPagePost{Changes: pageDefinitions}

			logging.VerboseF("Pushing %d pages.\n", len(pagePost.Changes))

			api, err := SalesforceLogin(cmd)

			if err != nil {
				return
			}

			result, err := api.Connection.Post("/skuid/api/v1/pages", pagePost)

			if err != nil {
				return
			}

			unquoted, _ := strconv.Unquote(string(result))

			response := &SkuidPagePostResult{}

			_ = json.Unmarshal([]byte(unquoted), response)

			if response.Success == false && len(response.Errors) > 0 {
				logging.VerboseLn("There were errors pushing Skuid Pages...")
				var errors []string
				for _, err := range response.Errors {
					errors = append(errors, err)
				}
				return fmt.Errorf("Error(s) encountered: %v", strings.Join(errors, "\n"))
			}

			logging.Printf("Pages successfully pushed to org: %s.\n", response.OrgName)

			return
		},
	}
)

func init() {
	RootCmd.AddCommand(pushCmd)
	flags.AddFlagFunctions(pushCmd, flags.SalesforceLoginFlags...)
	flags.AddFlags(pushCmd, flags.Modules)
	flags.AddFlags(pushCmd, flags.Directory, flags.PushFile)
}
