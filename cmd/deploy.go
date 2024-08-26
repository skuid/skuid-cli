package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

type deployCommander struct {
	factory         *cmdutil.Factory
	authOpts        pkg.AuthorizeOptions
	dir             string
	app             string
	ignoreSkuidDb   bool
	skipDataSources bool
	entities        []metadata.MetadataEntity
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
	cmdutil.AddSliceValueFlag(cmd, &c.entities, flags.Entities)
	// TODO: SkipDataSources can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	cmdutil.AddBoolFlag(cmd, &c.skipDataSources, flags.SkipDataSources)
	// TODO: Pages does not work as expected - remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	//cmdutil.AddStringSliceFlag(cmd, &c.pages, flags.Pages)

	cmd.MarkFlagsMutuallyExclusive(flags.IgnoreSkuidDb.Name, flags.SkipDataSources.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.Entities.Name, flags.App.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.Entities.Name, flags.IgnoreSkuidDb.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.Entities.Name, flags.SkipDataSources.Name)
	// TODO: Pages does not work as expected - once https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148 are addressed, if
	// Pages flag is maintained, uncomment below, else remove below
	//cmd.MarkFlagsMutuallyExclusive(flags.Entities.Name, flags.Pages.Name)

	return cmd
}

func NewCmdDeploy(factory *cmdutil.Factory) *cobra.Command {
	commander := new(deployCommander)
	commander.factory = factory
	return commander.GetCommand()
}

func (c *deployCommander) deploy(cmd *cobra.Command, _ []string) error {
	processStart := time.Now()
	fields := make(logrus.Fields)
	fields["processStart"] = processStart
	fields["processName"] = "deploy"
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
		return err
	}

	fields["authorized"] = true
	logging.WithFields(fields).Info("Authentication Successful")

	var targetDirectory string
	if targetDirectory, err = filepath.Abs(c.dir); err != nil {
		return err
	}
	fields["targetDirectory"] = targetDirectory

	planFilter := c.getPlanFilter(fields)
	archiveFilter, entitiesToArchive := c.getArchiveFilter(fields)
	if err = pkg.Deploy(auth, targetDirectory, archiveFilter, entitiesToArchive, planFilter); err != nil {
		return err
	}

	processFinish := time.Now()
	fields["processFinish"] = processFinish
	fields["processDuration"] = processFinish.Sub(processStart)
	logging.WithFields(fields).Info(color.Green.Sprint("Finished Deploy"))

	return nil
}

func (c *deployCommander) getPlanFilter(fields logrus.Fields) *pkg.NlxPlanFilter {
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

	return filter
}

// Resolve the pkg.ArchiveFilter to apply to the deployment
// Flags affecting the ArchiveFilter (e.g., Entities, SkipDataSources) should always be marked MutuallyExclusive
// so only one is eligible to be applied based on flags specified
func (c *deployCommander) getArchiveFilter(fields logrus.Fields) (pkg.ArchiveFilter, []metadata.MetadataEntity) {
	if len(c.entities) > 0 {
		// dedupe in case input contains same entity multiple times
		uniqueEntities := metadata.UniqueEntities(c.entities)

		var paths []string
		for _, e := range uniqueEntities {
			paths = append(paths, fmt.Sprintf("%q", e.Path))
		}
		fields["entities"] = paths
		logging.WithFields(fields).Infof("Deploying entities: %v", paths)
		return pkg.MetadataEntityArchiveFilter(c.entities), uniqueEntities
	}

	// skip datasources
	// TODO: This can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	if c.skipDataSources {
		fields["skipDataSources"] = c.skipDataSources
		logging.WithFields(fields).Info("Skipping deployment of all DataSources")
		return pkg.MetadataTypeArchiveFilter([]metadata.MetadataType{metadata.MetadataTypeDataSources}), nil
	}

	return nil, nil
}
