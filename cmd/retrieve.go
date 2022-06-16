package cmd

import (
	"time"

	// jsoniter. Fork of github.com/json-iterator/go
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/common"
	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/util"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:      true,
		SilenceErrors:     true, // we do not want to show users raw errors
		Example:           "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval",
		Use:               "retrieve",
		Short:             "Retrieve Skuid metadata from an Skuid NLX Site into a local directory.",
		Long:              "Retrieve Skuid metadata from a Skuid NLX Site and output it into a local directory.",
		PersistentPreRunE: common.PrerunValidation,
		RunE:              Retrieve,
	}
)

func Retrieve(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	start := time.Now()
	fields["process"] = "retrieve"
	fields["start"] = start

	logging.Get().Info("Starting retrieve")
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
	logging.Get().WithFields(fields).Debug("Credentials gathered.")

	var auth *pkg.Authorization
	if auth, err = pkg.Authorize(host, username, password); err != nil {
		return
	}

	fields["authorized"] = true
	logging.Get().WithFields(fields).Debug("Authentication successful")

	// we want the filter nil because it will be discarded without
	// initialization
	var filter *pkg.NlxPlanFilter = nil

	// initialize the filter dynamically based on
	// optional filter arguments. This lets us
	// expand the pattern down the road as more things
	// are required to be build
	initFilter := func() {
		logging.Get().WithFields(fields).Debug("Using filter.")
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
		fields["pageNames"] = pageNames
		filter.PageNames = pageNames
	}

	logging.Get().WithFields(fields).Debug("Getting Retrieve Plan.")

	var plans pkg.NlxPlanPayload
	if _, plans, err = pkg.GetRetrievePlan(auth, filter); err != nil {
		return
	}

	logging.Get().WithFields(fields).Debug("Got Retrieve Plan.")

	// noZip argument
	var noZip bool
	if noZip, err = cmd.Flags().GetBool(flags.NoZip.Name); err != nil {
		return
	}

	fields["NoZip"] = noZip

	logging.Get().WithFields(fields).Debugf("Zipping? %v", noZip)

	var results []pkg.NlxRetrievalResult
	if _, results, err = pkg.ExecuteRetrieval(auth, plans, noZip); err != nil {
		return
	}

	fields["results"] = len(results)
	fields["finished"] = time.Now()
	fields["retrievalDuration"] = time.Since(start)

	logging.Get().WithFields(fields).Debugf("Received %v Results.", len(results))

	var directory string
	if directory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	fields["directory"] = directory
	logging.Get().WithFields(fields).Debugf("Target directory is %v.", directory)

	var resultBytes [][]byte = make([][]byte, 0)
	for _, result := range results {
		resultBytes = append(resultBytes, result.Data)
	}

	logging.Get().WithFields(fields).Debug("Clearing Directory first!")

	if err = util.DeleteDirectories(directory, pkg.GetMetadataTypeDirNames()); err != nil {
		return
	}

	logging.Get().WithFields(fields).Debug("Directory Cleared.")
	fields["writeStart"] = time.Now()

	if err = util.WriteResultsToDisk(
		directory,
		resultBytes,
		noZip,
	); err != nil {
		return
	}

	logging.Get().WithFields(fields).Debug("Finished writing to disk.")

	return
}

func init() {
	TidesCmd.AddCommand(retrieveCmd)

	flags.AddFlags(retrieveCmd, flags.NLXLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName, flags.ApiVersion)
	flags.AddFlags(retrieveCmd, flags.Pages)
	flags.AddFlags(retrieveCmd, flags.NoZip)
}
