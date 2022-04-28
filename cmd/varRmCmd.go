package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var rmvarCmd = &cobra.Command{
	Use:   "rm-variable",
	Short: "Delete a Skuid site environment variable",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		logging.VerboseCommand("Delete Variable")

		api, err := pkg.SkuidNlxLogin(cmd)

		if err != nil {
			err = fmt.Errorf("Error logging in to Skuid site: %v", err)
			return
		}

		var variableName, variableDataService string

		variableName, err = cmd.Flags().GetString(flags.VariableName.Name)
		if err != nil {
			return
		}

		variableDataService, err = cmd.Flags().GetString(flags.VariableDataService.Name)
		if err != nil {
			return
		}

		variableStart := time.Now()
		err = pkg.RemoveEnvironmentSpecificConfigurations(api, variableName, variableDataService)
		if err != nil {
			err = fmt.Errorf("Error deleting variable in Skuid site: %v", err)
			return
		}

		successMessage := "Successfully deleted variable in Skuid Site"

		logging.VerboseSuccess(successMessage, variableStart)

		return
	},
}
