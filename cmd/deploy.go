package cmd

import (
	"fmt"

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
	cmdutil.AddValueFlag(cmd, &c.dir, flags.Dir)
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

func (c *deployCommander) deploy(cmd *cobra.Command, args []string) (err error) {
	message := fmt.Sprintf("Executing command %v entities to site %v from directory %v", logging.QuoteText(cmd.Name()), logging.ColorResource.Text(c.authOpts.Host), logging.ColorResource.QuoteText(c.dir))
	fields := logging.Fields{
		logging.CommandNameKey: cmd.Name(),
		"host":                 c.authOpts.Host,
		"username":             c.authOpts.Username,
		"sourceDirectory":      c.dir,
		"app":                  c.app,
		"ignoreSkuidDb":        c.ignoreSkuidDb,
		"skipDataSources":      c.skipDataSources,
		"entities":             logging.CSV(metadata.MetadataEntityPaths(c.entities).All()),
		"entitiesFrom":         fmt.Sprintf("--%v flag", flags.Entities.Name),
	}
	logger := logging.WithTracking("cmd.deploy", message, fields).StartTracking()
	defer func() {
		err = cmdutil.CheckError(cmd, err, recover())
		logger = logger.FinishTracking(err)
		err = cmdutil.HandleCommandResult(cmd, logger, err, fmt.Sprintf("Deployed site %v from %v", logging.ColorResource.Text(c.authOpts.Host), logging.ColorResource.QuoteText(c.dir)))
	}()

	auth, err := pkg.AuthorizeOnce(&c.authOpts)
	if err != nil {
		return err
	}
	archiveFilter, entitiesToArchive := c.getArchiveFilter(logger)
	options := pkg.DeployOptions{
		ArchiveFilter:     archiveFilter,
		Auth:              auth,
		EntitiesToArchive: entitiesToArchive,
		PlanFilter:        c.getPlanFilter(logger),
		SourceDirectory:   c.dir,
	}
	if err := pkg.Deploy(options); err != nil {
		return err
	}

	logger = logger.WithSuccess()
	return nil
}

func (c *deployCommander) getPlanFilter(logger *logging.Logger) *pkg.NlxPlanFilter {
	var filter *pkg.NlxPlanFilter
	initFilter := func() {
		if filter == nil {
			filter = &pkg.NlxPlanFilter{}
		}
	}

	// filter by app name
	if c.app != "" {
		initFilter()
		filter.AppName = c.app
		logger.Debugf("Filtering deployment to app %v", logging.ColorFilter.QuoteText(c.app))
	}

	// filter by page name
	// TODO: pages flag does not work as expected - remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	/*
		if len(c.pages) > 0 {
			logEntry.Data["pages"] = c.pages
			logEntry.Infof("Filtering retrieval to pages: %q", c.pages)
			filter.PageNames = c.pages
		}
	*/

	// ignore skuiddb
	if c.ignoreSkuidDb {
		initFilter()
		filter.IgnoreSkuidDb = c.ignoreSkuidDb
		logger.Debug(logging.ColorFilter.Text("Ignoring any problematic Skuid Database metadata"))
	}

	return filter
}

// Resolve the pkg.ArchiveFilter to apply to the deployment
// Flags affecting the ArchiveFilter (e.g., Entities, SkipDataSources) should always be marked MutuallyExclusive
// so only one is eligible to be applied based on flags specified
func (c *deployCommander) getArchiveFilter(logger *logging.Logger) (pkg.ArchiveFilter, []metadata.MetadataEntity) {
	if len(c.entities) > 0 {
		// dedupe in case input contains same entity multiple times
		uniqueEntities := metadata.UniqueEntities(c.entities)
		// paths := logging.CSV(seqs.Map(slices.Values(uniqueEntities), func(me metadata.MetadataEntity) string {
		// 	return me.Path
		// }))
		paths := logging.CSV(metadata.MetadataEntityPaths(uniqueEntities).All())
		logger.WithField("entityPaths", paths).Debugf("Filtering deployment to entities %v", logging.ColorFilter.Text(paths))
		return pkg.MetadataEntityArchiveFilter(uniqueEntities), uniqueEntities
	}

	// skip datasources
	// TODO: This can be removed once https://github.com/skuid/skuid-cli/issues/150 is resolved
	if c.skipDataSources {
		logger.Debugf("Skipping deployment of all %v", logging.ColorFilter.QuoteText("Data Sources"))
		return pkg.MetadataTypeArchiveFilter([]metadata.MetadataType{metadata.MetadataTypeDataSources}), nil
	}

	return nil, nil
}
