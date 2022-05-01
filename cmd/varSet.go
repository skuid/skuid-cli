package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

// setvarCmd represents the setvariable command
var varSetCmd = &cobra.Command{
	Use:   "set-variable",
	Short: "Set a Skuid site environment variable",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		logging.VerboseCommand("Set Variable")

		api, err := pkg.SkuidNlxLogin(cmd)

		if err != nil {
			err = fmt.Errorf("Error logging in to Skuid site: %v", err)
			return
		}

		var variableName, variableValue, variableDataService string

		if variableName, err = cmd.Flags().GetString(flags.VariableName.Name); err != nil {
			return
		}

		if variableValue, err = cmd.Flags().GetString(flags.VariableValue.Name); err != nil {
			return
		}

		if variableDataService, err = cmd.Flags().GetString(flags.VariableDataService.Name); err != nil {
			return
		}

		variableStart := time.Now()
		err = pkg.SetEnvironmentSpecificConfiguration(api, variableName, variableValue, variableDataService)
		if err != nil {
			err = fmt.Errorf("Error setting variable in Skuid site: %v", err)
			return
		}

		successMessage := "Successfully set variable in Skuid Site"

		logging.VerboseSuccess(successMessage, variableStart)

		return
	},
}

func init() {
	TidesCmd.AddCommand(varSetCmd)
	flags.AddFlags(varSetCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(varSetCmd, flags.VariableName, flags.VariableValue, flags.VariableDataService)
}
