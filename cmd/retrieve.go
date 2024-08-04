package cmd

import (
	"fmt"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/spf13/cobra"
)

type retrieveCommander struct {
	factory  *cmdutil.Factory
	authOpts pkg.AuthorizeOptions
	dir      string
	app      string
	since    *time.Time
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
	cmdutil.AddStringFlag(cmd, &c.dir, flags.Dir)
	cmdutil.AddStringFlag(cmd, &c.app, flags.App)
	cmdutil.AddValueFlag(cmd, &c.since, flags.Since)
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
	fields := make(logrus.Fields)
	start := time.Now()
	fields["process"] = "retrieve"
	fields["start"] = start
	logging.WithFields(fields).Info(color.Green.Sprint("Starting Retrieve"))

	fields["host"] = c.authOpts.Host
	fields["username"] = c.authOpts.Username
	logging.WithFields(fields).Debug("Credentials gathered")

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

	// we want the filter nil because it will be discarded without
	// initialization
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

	var sinceStr string
	if c.since != nil {
		filter.Since = c.since.UTC()
		sinceStr = filter.Since.Format(time.RFC3339)
		fields["sinceUTC"] = sinceStr
		logging.WithFields(fields).Info(fmt.Sprintf("retrieving metadata records updated since: %s", c.since.Format(time.RFC3339)))
	}

	logging.WithFields(fields).Info("Getting Retrieve Plan")

	var plans pkg.NlxPlanPayload
	if _, plans, err = pkg.GetRetrievePlan(auth, filter); err != nil {
		return
	}

	logging.WithFields(fields).Info("Got Retrieve Plan")

	// pliny and warden are supposed to give the since value back for the retrieve, but just in case...
	if sinceStr != "" {
		if plans.MetadataService.Since == "" {
			plans.MetadataService.Since = sinceStr
		}
		if plans.CloudDataService != nil {
			if plans.CloudDataService.Since == "" {
				plans.CloudDataService.Since = sinceStr
			}
		}
	}

	var results []pkg.NlxRetrievalResult
	if _, results, err = pkg.ExecuteRetrieval(auth, plans); err != nil {
		return
	}

	fields["results"] = len(results)
	fields["finished"] = time.Now()
	fields["retrievalDuration"] = time.Since(start)

	logging.WithFields(fields).Debugf("Received %v Results", color.Green.Sprint(len(results)))

	var targetDirectory string
	if targetDirectory, err = util.SanitizePath(c.dir); err != nil {
		return
	}

	fields["targetDirectory"] = targetDirectory
	logging.WithFields(fields).Infof("Target Directory is %v", color.Cyan.Sprint(targetDirectory))

	// TODO: put this behind a boolean command flag to avoid this process
	if err = pkg.ClearDirectories(targetDirectory); err != nil {
		logging.Get().Errorf("Unable to clear directory: %v", targetDirectory)
		return
	}

	fields["writeStart"] = time.Now()

	for _, v := range results {
		if err = pkg.WriteResultsToDisk(
			targetDirectory,
			pkg.WritePayload{
				PlanName: v.PlanName,
				PlanData: v.Data,
			},
		); err != nil {
			return
		}
	}

	logging.Get().Infof("Finished Writing to %v", color.Cyan.Sprint(targetDirectory))
	logging.WithFields(fields).Info(color.Green.Sprint("Finished Retrieve"))

	return
}
