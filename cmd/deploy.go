package cmd

import (
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/cmd/common"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
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
	flags.AddFlags(deployCmd, flags.PlinyHost, flags.Username)
	flags.AddFlags(deployCmd, flags.Password)
	flags.AddFlags(deployCmd, flags.Directory, flags.AppName)
	// TODO: SkipDataSources can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	flags.AddFlags(deployCmd, flags.IgnoreSkuidDb, flags.SkipDataSources)
	// pages flag does not work as expected so commenting out
	// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	//flags.AddFlags(deployCmd, flags.Pages)

	// do not allow ignoring skuid db errors when skipping datasources or vice-versa as errors can't occur if we're skipping all data sources
	flags.MarkFlagsMutuallyExclusive(deployCmd, [][]string{{flags.IgnoreSkuidDb.Name, flags.SkipDataSources.Name}})
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
	password, err := flags.GetPassword(cmd.Flags())
	if err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username
	logging.WithFields(fields).Debug("Gathered credentials")

	auth, err := pkg.Authorize(host, username, password)
	// we don't need it anymore - very inelegant approach but at least it is something for now
	// Clearing it here instead of in auth package which is the only place its accessed because the tests that exist
	// for auth rely on package global variables so clearing in there would break those tests as they currently exist.
	//
	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	password.Set("")
	if err != nil {
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
	// pages flag does not work as expected so commenting out
	// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	/*
		var pageNames []string
		if pageNames, err = cmd.Flags().GetStringArray(flags.Pages.Name); err != nil {
			return
		} else if len(pageNames) > 0 {
			initFilter()
			fields["pages"] = pageNames
			filter.PageNames = pageNames
		}
	*/

	// ignore skuiddb
	var ignoreSkuidDb bool

	if ignoreSkuidDb, err = cmd.Flags().GetBool(flags.IgnoreSkuidDb.Name); err != nil {
		return
	} else {
		initFilter()
		fields["ignoreSkuidDb"] = ignoreSkuidDb
		filter.IgnoreSkuidDb = ignoreSkuidDb
	}

	// skip datasources
	// TODO: This can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	var skipDataSources bool
	if skipDataSources, err = cmd.Flags().GetBool(flags.SkipDataSources.Name); err != nil {
		return
	}
	fields["skipDataSources"] = skipDataSources
	var excludedMetadataDirs []string
	if skipDataSources {
		logging.WithFields(fields).Info("Skipping deployment of all DataSources")
		var mdDirName string
		if mdDirName, err = pkg.GetMetadataTypeDirName("DataSources"); err != nil {
			return
		}
		excludedMetadataDirs = append(excludedMetadataDirs, mdDirName)
	}

	// get directory argument
	var targetDirectory string
	if targetDirectory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	if targetDirectory, err = util.SanitizePath(targetDirectory); err != nil {
		return
	}

	fields["targetDirectory"] = targetDirectory
	logging.WithFields(fields).Info("Getting Deployment Payload")

	var deploymentPlan []byte
	if deploymentPlan, _, err = pkg.Archive(os.DirFS(targetDirectory), util.NewFileUtil(), pkg.MetadataDirArchiveFilter(excludedMetadataDirs)); err != nil {
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
		// Error will be logged via main.go
		return
	}

	fields["results"] = len(results)

	for _, result := range results {
		logging.Get().Tracef("result: %v", result.Url)
	}

	logging.Get().Info(color.Green.Sprint("Finished Deploy"))

	return
}
