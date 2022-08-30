package cmd

import (
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/skuid/domain"
	"github.com/skuid/domain/flags"
	"github.com/skuid/domain/logging"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/common"
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
	logging.WithFields(fields).Info(color.Green.Sprint("Starting Deploy"))

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
	var auth *domain.Authorization
	if auth, err = domain.Authorize(host, username, password); err != nil {
		return
	}

	fields["authorized"] = true
	logging.WithFields(fields).Info("Authentication Successful")

	// we want the filter nil because it will be discarded without
	// initialization
	var filter *domain.NlxPlanFilter = nil

	// initialize the filter dynamically based on
	// optional filter arguments. This lets us
	// expand the pattern down the road as more things
	// are required to be build
	initFilter := func() {
		if filter == nil {
			filter = &domain.NlxPlanFilter{}
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

	logging.WithFields(fields).Info("Getting Deployment Plan")

	var deploymentPlan []byte
	if deploymentPlan, err = domain.Archive(targetDirectory, nil); err != nil {
		return
	}

	fields["deploymentBytes"] = len(deploymentPlan)
	logging.WithFields(fields).Infof("Got Deployment Plan")

	// get the plan
	var plans domain.NlxDynamicPlanMap
	if _, plans, err = domain.PrepareDeployment(auth, deploymentPlan, filter); err != nil {
		logging.Get().Warnf("Unable to prepare deployment: %v", err)
		return
	}

	fields["plans"] = len(plans)
	logging.WithFields(fields)

	logging.Get().Info("Executing Deployment Plan")

	var results []domain.NlxDeploymentResult
	if _, results, err = domain.ExecuteDeployPlan(auth, plans, targetDirectory); err != nil {
		logging.Get().Errorf("Unable to execute deployment: %v", color.Red.Sprint(err))
		return
	}

	fields["results"] = len(results)

	for _, result := range results {
		logging.Get().Tracef("result: %v", result.Url)
	}

	logging.Get().Info(color.Green.Sprint("Finished Deploy"))

	return
}
