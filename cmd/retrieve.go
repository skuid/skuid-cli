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

			logging.VerboseCommand("Retrieve Skuid NLX Metadata")

			retrieveStart := time.Now()

			// var host, username, password string

			// if host, err = cmd.Flags().GetString(flags.Host.Name); err != nil {
			// 	return
			// }

			// if username, err = cmd.Flags().GetString(flags.Username.Name); err != nil {
			// 	return
			// }

			// if password, err = cmd.Flags().GetString(flags.Password.Name); err != nil {
			// 	return
			// }

			// var accessToken string
			// if accessToken, err = nlx.GetAccessToken(host, username, password); err != nil {
			// 	return
			// }

			// var authToken string
			// if authToken, err = nlx.GetAuthorizationToken(accessToken ); err != nil {
			// 	return
			// }

			var appName string
			if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
				return
			}

			api, err := pkg.SkuidNlxLogin(cmd)
			if err != nil {
				return
			}

			plan, err := pkg.GetRetrievePlan(api, appName)
			if err != nil {
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

			logging.VerboseSuccess("Skuid NLX Metadata Retrieved", retrieveStart)

			return
		},
	}
)

func init() {
	TidesCmd.AddCommand(retrieveCmd)

	flags.AddFlags(retrieveCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName)
	flags.AddFlags(retrieveCmd, flags.NoZip)
}
