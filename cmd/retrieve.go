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

func NewCmdRetrieve(cd *cmdutil.Factory) *cobra.Command {
	retrieveTemplate := &cmdutil.CmdTemplate{
		Use:     "retrieve",
		Short:   "Retrieve a Skuid NLX Site",
		Long:    "Retrieve Skuid metadata from a Skuid NLX Site and output it into a local directory",
		Example: "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./my-site-objects --app myapp --since 4h",
		Flags: &cmdutil.CommandFlags{
			// pages flag does not work as expected so commenting out
			// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
			// flags.Pages
			String:         []*flags.Flag[string]{flags.Username, flags.Dir, flags.App},
			RedactedString: []*flags.Flag[flags.RedactedString]{flags.Password},
			CustomString:   []*flags.Flag[flags.CustomString]{flags.Since, flags.Host},
		},
	}

	return retrieveTemplate.ToCommand(cd, nil, nil, retrieve)
}

func retrieve(factory *cmdutil.Factory, cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	start := time.Now()
	fields["process"] = "retrieve"
	fields["start"] = start

	logging.Get().Info(color.Green.Sprint("Starting Retrieve"))
	// get required arguments
	host, err := flags.GetCustomString(cmd.Flags(), flags.Host.Name)
	if err != nil {
		return
	}
	username, err := cmd.Flags().GetString(flags.Username.Name)
	if err != nil {
		return
	}
	password, err := flags.GetRedactedString(cmd.Flags(), flags.Password.Name)
	if err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username
	logging.WithFields(fields).Debug("Credentials gathered")

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

	// we want the filter nil because it will be discarded without
	// initialization
	var filter *pkg.NlxPlanFilter = nil

	// initialize the filter dynamically based on
	// optional filter arguments. This lets us
	// expand the pattern down the road as more things
	// are required to be build
	initFilter := func() {
		logging.WithFields(fields).Debug("Using filter")
		if filter == nil {
			filter = &pkg.NlxPlanFilter{}
		}
	}

	// filter by app name
	var appName string
	if appName, err = cmd.Flags().GetString(flags.App.Name); err != nil {
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
			fields["pageNames"] = pageNames
			filter.PageNames = pageNames
		}
	*/

	sinceF := cmd.Flags().Lookup(flags.Since.Name)
	if sinceF == nil {
		return fmt.Errorf("flag accessed but not defined: %s", flags.Since.Name)
	}
	sinceStr := sinceF.Value.String()
	if sinceStr != "" {
		if sec, nano, err := util.ParseTimestamp(sinceStr, 0); err != nil {
			return err
		} else {
			initFilter()
			localSince := time.Unix(sec, nano)
			filter.Since = localSince.UTC()
			sinceStr = filter.Since.Format(time.RFC3339)
			fields["sinceUTC"] = sinceStr
			logging.WithFields(fields).Info(fmt.Sprintf("retrieving metadata records updated since: %s", localSince.Format(time.RFC3339)))
		}
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

	var directory string
	if directory, err = cmd.Flags().GetString(flags.Dir.Name); err != nil {
		return
	}

	if directory, err = util.SanitizePath(directory); err != nil {
		return
	}

	fields["targetDirectory"] = directory
	logging.WithFields(fields).Infof("Target Directory is %v", color.Cyan.Sprint(directory))

	// TODO: put this behind a boolean command flag to avoid this process
	if err = pkg.ClearDirectories(directory); err != nil {
		logging.Get().Errorf("Unable to clear directory: %v", directory)
		return
	}

	fields["writeStart"] = time.Now()

	for _, v := range results {
		if err = pkg.WriteResultsToDisk(
			directory,
			pkg.WritePayload{
				PlanName: v.PlanName,
				PlanData: v.Data,
			},
		); err != nil {
			return
		}
	}

	logging.Get().Infof("Finished Writing to %v", color.Cyan.Sprint(directory))
	logging.WithFields(fields).Info(color.Green.Sprint("Finished Retrieve"))

	return
}
