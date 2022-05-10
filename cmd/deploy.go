package cmd

import (
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/validation"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
)

var deployCmd = &cobra.Command{
	SilenceErrors:     true,
	SilenceUsage:      true,
	Use:               "deploy",
	Short:             "Deploy local Skuid metadata to a Skuid Platform Site.",
	Long:              "Deploy Skuid metadata stored within a local file system directory to a Skuid Platform Site.",
	PersistentPreRunE: validation.PrerunValidation,
	RunE:              Deploy,
}

func init() {
	TidesCmd.AddCommand(deployCmd)
	flags.AddFlags(deployCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(deployCmd, flags.Directory, flags.AppName, flags.ApiVersion)
	flags.AddFlags(deployCmd, flags.Pages)
}

func Deploy(cmd *cobra.Command, _ []string) (err error) {
	// get required authentication arguments
	var host, username, password string
	if host, err = cmd.Flags().GetString(flags.Host.Name); err != nil {
		return
	} else if username, err = cmd.Flags().GetString(flags.Username.Name); err != nil {
		return
	} else if password, err = cmd.Flags().GetString(flags.Password.Name); err != nil {
		return
	}

	// auth
	var auth *nlx.Authorization
	if auth, err = nlx.Authorize(host, username, password); err != nil {
		return
	}

	// we want the filter nil because it will be discarded without
	// initialization
	var filter *nlx.NlxPlanFilter = nil

	// initialize the filter dynamically based on
	// optional filter arguments. This lets us
	// expand the pattern down the road as more things
	// are required to be build
	initFilter := func() {
		if filter == nil {
			filter = &nlx.NlxPlanFilter{}
		}
	}

	// filter by app name
	var appName string
	if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
		return
	} else if appName != "" {
		initFilter()
		filter.AppName = appName
	}

	// filter by page name
	var pageNames []string
	if pageNames, err = cmd.Flags().GetStringArray(flags.Pages.Name); err != nil {
		return
	} else if len(pageNames) > 0 {
		initFilter()
		filter.PageNames = pageNames
	}

	// get directory argument
	var targetDirectory string
	if targetDirectory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	var deploymentPlan []byte
	if deploymentPlan, err = nlx.Archive(targetDirectory, nil); err != nil {
		return
	}

	// get the plan
	var plans nlx.NlxDynamicPlanMap
	if _, plans, err = nlx.PrepareDeployment(auth, deploymentPlan, filter); err != nil {
		return
	}

	var results []nlx.NlxDeploymentResult
	if _, results, err = nlx.ExecuteDeployPlan(auth, plans, targetDirectory); err != nil {
		return
	}

	for _, result := range results {
		logging.VerboseF("result: %v\n", result)
	}

	return
}
