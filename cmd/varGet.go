package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var varGetCmd = &cobra.Command{
	Use:   "variables",
	Short: "Get a list of Skuid site environment variables.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		logging.VerboseCommand("Get Variables")

		api, err := pkg.SkuidNlxLogin(cmd)

		if err != nil {
			err = fmt.Errorf("Error logging in to Skuid site: %v", err)
			return
		}

		escResult, err := pkg.GetEnvironmentSpecificConfigurations(api, true)
		if err != nil {
			err = fmt.Errorf("Error getting variables from Skuid site: %v", err)
			return
		}

		body := tablewriter.NewWriter(os.Stdout)
		body.SetHeader([]string{"Name", "Data Service"})
		for _, esc := range escResult {
			body.Append([]string{esc.Name, esc.DataServiceName})
		}

		logging.VerboseLn("Successfully retrieved variables from Skuid site")

		body.Render()

		return
	},
}

func init() {
	TidesCmd.AddCommand(varGetCmd)
	flags.AddFlags(varGetCmd, flags.PlatformLoginFlags...)
}
