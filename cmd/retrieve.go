package cmd

import (
	"time"

	// jsoniter. Fork of github.com/json-iterator/go
	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/skuid/domain"
	"github.com/skuid/domain/flags"
	"github.com/skuid/domain/logging"
	"github.com/skuid/domain/util"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/common"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:      true,
		SilenceErrors:     true, // we do not want to show users raw errors
		Example:           "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval",
		Use:               "retrieve",
		Short:             "Retrieve Skuid metadata from an Skuid NLX Site into a local directory",
		Long:              "Retrieve Skuid metadata from a Skuid NLX Site and output it into a local directory",
		PersistentPreRunE: common.PrerunValidation,
		RunE:              Retrieve,
	}
)

func Retrieve(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	start := time.Now()
	fields["process"] = "retrieve"
	fields["start"] = start

	logging.Get().Info(color.Green.Sprint("Starting Retrieve"))
	// get required arguments
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
	fields["password"] = password != ""
	logging.WithFields(fields).Debug("Credentials gathered")

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
		logging.WithFields(fields).Debug("Using filter")
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
		fields["pageNames"] = pageNames
		filter.PageNames = pageNames
	}

	logging.WithFields(fields).Info("Getting Retrieve Plan")

	var plans domain.NlxPlanPayload
	if _, plans, err = domain.GetRetrievePlan(auth, filter); err != nil {
		return
	}

	logging.WithFields(fields).Info("Got Retrieve Plan")

	var results []domain.NlxRetrievalResult
	if _, results, err = domain.ExecuteRetrieval(auth, plans); err != nil {
		return
	}

	fields["results"] = len(results)
	fields["finished"] = time.Now()
	fields["retrievalDuration"] = time.Since(start)

	logging.WithFields(fields).Debugf("Received %v Results", color.Green.Sprint(len(results)))

	var directory string
	if directory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	fields["directory"] = directory
	logging.WithFields(fields).Infof("Target Directory is %v", color.Cyan.Sprint(directory))

	domain.ClearDirectories(directory)

	fields["writeStart"] = time.Now()

	for _, v := range results {
		if err = util.WriteResultsToDisk(
			directory,
			util.WritePayload{
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

func init() {
	TidesCmd.AddCommand(retrieveCmd)

	flags.AddFlags(retrieveCmd, flags.NLXLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName, flags.ApiVersion)
	flags.AddFlags(retrieveCmd, flags.Pages)
}
