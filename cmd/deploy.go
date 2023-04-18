package cmd

import (
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/cmd/common"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var deployCmd = &cobra.Command{
	SilenceUsage:      true,
	Use:               "deploy",
	Short:             "Deploy local Skuid metadata to a Skuid NLX Site",
	Long:              "Deploy Skuid metadata stored within a local file system directory to a Skuid NLX Site",
	PersistentPreRunE: common.PrerunValidation,
	RunE:              Deploy,
}

func init() {
	flags.AddFlags(deployCmd, flags.NLXLoginFlags...)
	flags.AddFlags(deployCmd, flags.Directory, flags.AppName)
	flags.AddFlags(deployCmd, flags.IgnoreSkuidDb)
	flags.AddFlags(deployCmd, flags.Pages)
	AppCmd = append(AppCmd, deployCmd)
}

func Deploy(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["start"] = time.Now()
	fields["process"] = "deploy"
	logging.WithFields(fields).Info(color.Green.Sprint("Starting Deploy"))

	// get required authentication arguments
	host, err := cmd.Flags().GetString(flags.PlinyHost.Name)
	if err != nil {
		return
	}
	username, err := cmd.Flags().GetString(flags.Username.Name)
	if err != nil {
		return
	}
	password, err := cmd.Flags().GetString(flags.Password.Name)
	if err != nil {
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
	logging.WithFields(fields).Info("Authentication Successful")

	// only create the filter struct if it hasn't been created yet
	var filter *pkg.NlxPlanFilter = nil
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

	// ignore skuiddb
	var ignoreSkuidDb bool

	if ignoreSkuidDb, err = cmd.Flags().GetBool(flags.IgnoreSkuidDb.Name); err != nil {
		return
	} else {
		initFilter()
		fields["ignoreSkuidDb"] = ignoreSkuidDb
		filter.IgnoreSkuidDb = ignoreSkuidDb
	}

	// get directory argument
	var targetDirectory string
	if targetDirectory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	} else if targetDirectory == "" {
		targetDirectory = "."
	}

	fields["targetDirectory"] = targetDirectory
	logging.WithFields(fields).Info("Getting Deployment Payload")

	var deploymentPlan []byte
	if deploymentPlan, err = pkg.Archive(targetDirectory, nil); err != nil {
		return
	}

	fields["deploymentBytes"] = len(deploymentPlan)
	logging.WithFields(fields).Info("Got Deployment Payload")

	// get the plan
	logging.WithFields(fields).Info("Getting Deployment Plan")
	var plans pkg.NlxDynamicPlanMap
	if _, plans, err = pkg.GetDeployPlan(auth, deploymentPlan, filter); err != nil {
		logging.Get().Errorf("Unable to prepare deployment: %v", err)
		return
	}
	logging.WithFields(fields).Info("Got Deployment Plan")

	fields["plans"] = len(plans)

	logging.WithFields(fields).Info("Executing Deployment Plan")

	var results []pkg.NlxDeploymentResult
	if _, results, err = pkg.ExecuteDeployPlan(auth, plans, targetDirectory); err != nil {
		logging.Get().Errorf("Unable to execute deployment: %v", err)
		return
	}

	fields["results"] = len(results)

	for _, result := range results {
		logging.Get().Tracef("result: %v", result.Url)
	}

	logging.Get().Info(color.Green.Sprint("Finished Deploy"))

	return
}
