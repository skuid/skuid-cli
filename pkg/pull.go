package pkg

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull Skuid Pages from Salesforce into a local directory.",
	Long:  `Pull your Skuid Pages from your Salesforce org to a local directory.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		//login to the Force API
		api, err := SalesforceLogin(cmd)

		if err != nil {
			return err
		}

		var targetDir string
		if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
			return
		}

		// Provide a default for targetDir
		if targetDir == "" {
			targetDir = "skuidpages"
		}

		//build the module and name query paramaters
		query := url.Values{}

		var noModule bool
		if noModule, err = cmd.Flags().GetBool(flags.NoModule.Name); err != nil {
			return
		}

		if noModule {
			query.Add("nomodule", strconv.FormatBool(noModule))
		}

		var modules, pages []string
		if modules, err = cmd.Flags().GetStringArray(flags.Modules.Name); err != nil {
			return
		}

		if len(modules) > 0 {
			for _, module := range modules {
				query.Add("module", module)
			}
		}

		if pages, err = cmd.Flags().GetStringArray(flags.Pages.Name); err != nil {
			return
		}

		if len(pages) > 0 {
			for _, page := range pages {
				query.Add("page", page)
			}
		}

		//query the API for all pages in the requested module
		result, err := api.Connection.Get("/skuid/api/v1/pages", query)

		if err != nil {
			return err
		}

		var pulledPages map[string]SkuidPullResponse

		//you have to unquote the string because what comes back
		//is escaped json
		unquoted, _ := strconv.Unquote(string(result))

		//unmarshal all pages into the type PullResponse
		err = json.Unmarshal([]byte(unquoted), &pulledPages)

		if err != nil {
			return err
		}

		for _, pageRecord := range pulledPages {
			//write the page in the at rest format
			err = pageRecord.WriteAtRest(targetDir)
			if err != nil {
				return err
			}
		}

		logging.Printf("Pages written to %s\n", targetDir)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
	flags.AddFlagFunctions(pullCmd, flags.SalesforceLoginFlags...)
	flags.AddFlags(pullCmd, flags.Directory)            // strings
	flags.AddFlags(pullCmd, flags.Modules, flags.Pages) // string arrays (this is so annoying)
	flags.AddFlags(pullCmd, flags.NoModule)
}
