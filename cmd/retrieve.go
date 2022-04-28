package cmd

import (
	"fmt"
	"time"

	// jsoniter. Fork of github.com/json-iterator/go
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/validation"
	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:      true,
		SilenceErrors:     true, // we do not want to show users raw errors
		Example:           "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval",
		Use:               "retrieve",
		Short:             "Retrieve Skuid metadata from a Site into a local directory.",
		Long:              "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
		PersistentPreRunE: validation.PrerunValidation,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {

			logging.VerboseCommand("Retrieve Metadata")

			api, err := pkg.SkuidNlxLogin(cmd)

			retrieveStart := time.Now()

			if err != nil {
				err = fmt.Errorf("Error logging in to Skuid site: %v", err)
				return
			}

			var appName string
			if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
				return
			}

			plan, err := pkg.GetRetrievePlan(api, appName)
			if err != nil {
				err = fmt.Errorf("Error getting retrieve plan: %v", err)
				return
			}

			var noZip bool
			if noZip, err = cmd.Flags().GetBool(flags.NoZip.Name); err != nil {
				return
			}

			results, err := pkg.ExecuteRetrievePlan(api, plan, noZip)
			if err != nil {
				err = fmt.Errorf("Error executing retrieve plan: %v", err)
				return
			}

			var targetDir string
			if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
				return
			}

			err = pkg.WriteResultsToDisk(targetDir, results, pkg.WriteNewFile, pkg.CreateDirectory, pkg.ReadExistingFile, noZip)
			if err != nil {
				err = fmt.Errorf("Error writing results to disk: %v", err)
				return
			}

			successMessage := "Successfully retrieved metadata from Skuid Site"

			logging.VerboseSuccess(successMessage, retrieveStart)

			return
		},
	}
)

func init() {
	TidesCmd.AddCommand(retrieveCmd)

	flags.AddFlagFunctions(retrieveCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName)
	flags.AddFlags(retrieveCmd, flags.NoZip)
}
