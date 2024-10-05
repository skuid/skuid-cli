package cmd

import (
	"fmt"
	"time"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/spf13/cobra"
)

type retrieveCommander struct {
	factory  *cmdutil.Factory
	authOpts pkg.AuthorizeOptions
	dir      string
	app      string
	since    *time.Time
	noClean  bool
}

func (c *retrieveCommander) GetCommand() *cobra.Command {
	template := &cmdutil.CmdTemplate{
		Use:     "retrieve",
		Short:   "Retrieve a Skuid NLX Site",
		Long:    "Retrieve Skuid metadata from a Skuid NLX Site and output it into a local directory",
		Example: "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./my-site-objects --app myapp --since 4h",
	}
	cmd := template.ToCommand(c.retrieve)

	cmdutil.AddAuthFlags(cmd, &c.authOpts)
	cmdutil.AddValueFlag(cmd, &c.dir, flags.Dir)
	cmdutil.AddStringFlag(cmd, &c.app, flags.App)
	cmdutil.AddValueFlag(cmd, &c.since, flags.Since)
	cmdutil.AddBoolFlag(cmd, &c.noClean, flags.NoClean)
	// TODO: Pages does not work as expected - remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	//cmdutil.StringSliceFlag(cmd, &c.pages, flags.Pages)

	return cmd
}

func NewCmdRetrieve(factory *cmdutil.Factory) *cobra.Command {
	commander := new(retrieveCommander)
	commander.factory = factory
	return commander.GetCommand()
}

func (c *retrieveCommander) retrieve(cmd *cobra.Command, _ []string) (err error) {
	message := fmt.Sprintf("Executing command %v entities from site %v to directory %v", logging.QuoteText(cmd.Name()), logging.ColorResource.Text(c.authOpts.Host), logging.ColorResource.QuoteText(c.dir))
	fields := logging.Fields{
		logging.CommandNameKey: cmd.Name(),
		"host":                 c.authOpts.Host,
		"username":             c.authOpts.Username,
		"targetDirectory":      c.dir,
		"app":                  c.app,
		"since":                c.since,
		"noClean":              c.noClean,
	}
	logger := logging.WithTracking("cmd.retrieve", message, fields).StartTracking()
	defer func() {
		err = cmdutil.CheckError(cmd, err, recover())
		logger = logger.FinishTracking(err)
		err = cmdutil.HandleCommandResult(cmd, logger, err, fmt.Sprintf("Retrieved site %v to %v", logging.ColorResource.Text(c.authOpts.Host), logging.ColorResource.QuoteText(c.dir)))
	}()

	auth, err := pkg.AuthorizeOnce(&c.authOpts)
	if err != nil {
		return err
	}
	options := pkg.RetrieveOptions{
		Auth:            auth,
		NoClean:         c.noClean,
		PlanFilter:      c.getPlanFilter(logger),
		Since:           c.since,
		TargetDirectory: c.dir,
	}
	if err := pkg.Retrieve(options); err != nil {
		return err
	}

	logger = logger.WithSuccess()
	return nil
}

func (c *retrieveCommander) getPlanFilter(logger *logging.Logger) *pkg.NlxPlanFilter {
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
		logger.Debugf("Filtering retrieval to app %v", logging.ColorFilter.QuoteText(c.app))
	}

	// filter by page name
	// TODO: pages flag does not work as expected - remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	/*
		if len(c.pages) > 0 {
			initFilter()
			logger.Infof("Filtering retrieval to pages: %q", c.pages)
			filter.PageNames = c.pages
		}
	*/

	if c.since != nil {
		initFilter()
		filter.Since = c.since.UTC()
		logger.Debugf("Filtering retrieval to metadata records updated since %v", logging.ColorFilter.QuoteText(flags.FormatSince(c.since)))
	}

	return filter
}
