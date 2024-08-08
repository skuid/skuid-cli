package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
)

type deployCommander struct {
	factory         *cmdutil.Factory
	authOpts        pkg.AuthorizeOptions
	dir             string
	app             string
	ignoreSkuidDb   bool
	skipDataSources bool
}

func (c *deployCommander) GetCommand() *cobra.Command {
	template := cmdutil.CmdTemplate{
		Use:     "deploy",
		Short:   "Deploy local Skuid metadata to a Skuid NLX Site",
		Long:    "Deploy Skuid metadata stored within a local file system directory to a Skuid NLX Site",
		Example: "deploy -u myUser -p myPassword --host my-site.skuidsite.com --dir ./my-site-objects --app myapp",
	}
	cmd := template.ToCommand(c.deploy)

	cmdutil.AddAuthFlags(cmd, &c.authOpts)
	cmdutil.AddStringFlag(cmd, &c.dir, flags.Dir)
	cmdutil.AddStringFlag(cmd, &c.app, flags.App)
	cmdutil.AddBoolFlag(cmd, &c.ignoreSkuidDb, flags.IgnoreSkuidDb)
	// TODO: SkipDataSources can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	cmdutil.AddBoolFlag(cmd, &c.skipDataSources, flags.SkipDataSources)
	// TODO: Pages does not work as expected - remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	//cmdutil.StringSliceFlag(cmd, &c.pages, flags.Pages)

	cmd.MarkFlagsMutuallyExclusive(flags.IgnoreSkuidDb.Name, flags.SkipDataSources.Name)

	return cmd
}

func NewCmdDeploy(factory *cmdutil.Factory) *cobra.Command {
	commander := new(deployCommander)
	commander.factory = factory
	return commander.GetCommand()
}

func (c *deployCommander) deploy(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	fields["start"] = time.Now()
	fields["process"] = "deploy"
	logging.WithFields(fields).Info(color.Green.Sprint("Starting Deploy"))

	fields["host"] = c.authOpts.Host
	fields["username"] = c.authOpts.Username
	logging.WithFields(fields).Debug("Gathered credentials")

	auth, err := pkg.Authorize(&c.authOpts)
	// we don't need it anymore - very inelegant approach but at least it is something for now
	// Clearing it here instead of in auth package which is the only place its accessed because the tests that exist
	// for auth rely on package global variables so clearing in there would break those tests as they currently exist.
	//
	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	// intentionally ignoring error since there is nothing we can do and we should fail entirely as a result
	_ = c.authOpts.Password.Set("")
	if err != nil {
		return
	}

	fields["authorized"] = true
	logging.WithFields(fields).Info("Authentication Successful")

	var filter *pkg.NlxPlanFilter = &pkg.NlxPlanFilter{}

	// filter by app name
	if c.app != "" {
		fields["appName"] = c.app
		filter.AppName = c.app
	}

	// filter by page name
	// TODO: pages flag does not work as expected - remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	/*
		if len(c.pages) > 0 {
			fields["pages"] = c.pages
			filter.PageNames = c.pages
		}
	*/

	// ignore skuiddb
	fields["ignoreSkuidDb"] = c.ignoreSkuidDb
	filter.IgnoreSkuidDb = c.ignoreSkuidDb

	// skip datasources
	// TODO: This can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	fields["skipDataSources"] = c.skipDataSources
	var excludedMetadataDirs []pkg.MetadataType
	if c.skipDataSources {
		logging.WithFields(fields).Info("Skipping deployment of all DataSources")
		excludedMetadataDirs = append(excludedMetadataDirs, pkg.MetadataTypeDataSources)
	}

	// get directory argument
	var targetDirectory string
	if targetDirectory, err = filepath.Abs(c.dir); err != nil {
		return
	}

	fields["targetDirectory"] = targetDirectory
	logging.WithFields(fields).Info("Getting Deployment Payload")

	var deploymentPlan []byte
	if deploymentPlan, _, err = pkg.Archive(os.DirFS(targetDirectory), util.NewFileUtil(), pkg.MetadataTypeArchiveFilter(excludedMetadataDirs)); err != nil {
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
