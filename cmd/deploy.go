package cmd

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/common"
	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

var deployCmd = &cobra.Command{
	SilenceErrors:     true,
	SilenceUsage:      true,
	Use:               "deploy",
	Short:             "Deploy local Skuid metadata to a Skuid NLX Site",
	Long:              "Deploy Skuid metadata stored within a local file system directory to a Skuid NLX Site",
	PersistentPreRunE: common.PrerunValidation,
	RunE:              Deploy,
}

func init() {
	TidesCmd.AddCommand(deployCmd)
	flags.AddFlags(deployCmd, flags.NLXLoginFlags...)
	flags.AddFlags(deployCmd, flags.Directory, flags.AppName, flags.ApiVersion)
	flags.AddFlags(deployCmd, flags.Pages)
}

func Deploy(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["start"] = time.Now()
	fields["process"] = "deploy"
	logging.WithFields(fields).Info("Starting Deploy")

	// get required authentication arguments
	var host, username, password string
	if host, err = cmd.Flags().GetString(flags.PlinyHost.Name); err != nil {
		return
	} else if username, err = cmd.Flags().GetString(flags.Username.Name); err != nil {
		return
	} else if password, err = cmd.Flags().GetString(flags.Password.Name); err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username
	logging.WithFields(fields).Debug("Gathered credentials")

	// auth
	var auth *pkg.Authorization
	if auth, err = pkg.Authorize(host, username, password); err != nil {
		return
	}

	fields["authorized"] = true
	logging.WithFields(fields).Debug("Successfully Authenticated")

	// we want the filter nil because it will be discarded without
	// initialization
	var filter *pkg.NlxPlanFilter = nil

	// initialize the filter dynamically based on
	// optional filter arguments. This lets us
	// expand the pattern down the road as more things
	// are required to be build
	initFilter := func() {
		if filter == nil {
			filter = &pkg.NlxPlanFilter{}
		}
	}

	// filter by app name
	var appName string
	if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
		return
	} else if appName != "" {
		initFilter()
		fields["appName"] = appName
		filter.AppName = appName
	}

	// filter by page name
	var pageNames []string
	if pageNames, err = cmd.Flags().GetStringArray(flags.Pages.Name); err != nil {
		return
	} else if len(pageNames) > 0 {
		initFilter()
		fields["pages"] = pageNames
		filter.PageNames = pageNames
	}

	// get directory argument
	var targetDirectory string
	if targetDirectory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	} else if targetDirectory != "" {
		fields["targetDirectory"] = targetDirectory
	}

	logging.WithFields(fields).Debug("Getting Deployment Plan")

	var deploymentPlan []byte
	if deploymentPlan, err = pkg.Archive(targetDirectory, nil); err != nil {
		return
	}

	fields["deploymentBytes"] = len(deploymentPlan)
	logging.WithFields(fields).Debugf("Got Deployment Plan: Size (%v)", len(deploymentPlan))

	// get the plan
	var plans pkg.NlxDynamicPlanMap
	if _, plans, err = pkg.PrepareDeployment(auth, deploymentPlan, filter); err != nil {
		return
	}

	fields["plans"] = len(plans)
	logging.WithFields(fields)

	var results []pkg.NlxDeploymentResult
	if _, results, err = pkg.ExecuteDeployPlan(auth, plans, targetDirectory); err != nil {
		return
	}

	fields["results"] = len(results)

	for _, result := range results {
		logging.Get().Debugf("result: %v\n", result)
	}

	return
}
