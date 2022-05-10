package cmd

import (

	// jsoniter. Fork of github.com/json-iterator/go
	"github.com/spf13/cobra"

	"github.com/skuid/tides/cmd/validation"
	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:      true,
		SilenceErrors:     true, // we do not want to show users raw errors
		Example:           "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval",
		Use:               "retrieve",
		Short:             "Retrieve Skuid metadata from a Site into a local directory.",
		Long:              "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
		PersistentPreRunE: validation.PrerunValidation,
		RunE:              Retrieve,
	}
)

func Retrieve(cmd *cobra.Command, _ []string) (err error) {
	// get required arguments
	var host, username, password string
	if host, err = cmd.Flags().GetString(flags.Host.Name); err != nil {
		return
	} else if username, err = cmd.Flags().GetString(flags.Username.Name); err != nil {
		return
	} else if password, err = cmd.Flags().GetString(flags.Password.Name); err != nil {
		return
	}

	var auth *nlx.Authorization
	if auth, err = nlx.Authorize(host, username, password); err != nil {
		return
	}

	// we want the filter nil because it will be discarded without
	// initialization
	var filter *nlx.NlxPlanFilter = nil

	// initialize the filter dynamically based on
	// optional filter arguments. This lets us
	// expand the pattern down the road as more things
	// are required to be build
	initFilter := func() {
		if filter == nil {
			filter = &nlx.NlxPlanFilter{}
		}
	}

	// filter by app name
	var appName string
	if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
		return
	} else if appName != "" {
		initFilter()
		filter.AppName = appName
	}

	// filter by page name
	var pageNames []string
	if pageNames, err = cmd.Flags().GetStringArray(flags.Pages.Name); err != nil {
		return
	} else if len(pageNames) > 0 {
		initFilter()
		filter.PageNames = pageNames
	}

	var plans nlx.NlxPlanPayload
	if _, plans, err = nlx.GetRetrievePlan(auth, filter); err != nil {
		return
	}

	// zip argument
	var zip bool
	if zip, err = cmd.Flags().GetBool(flags.NoZip.Name); err != nil {
		return
	}

	var results []nlx.NlxRetrievalResult
	if _, results, err = nlx.ExecuteRetrieval(auth, plans, zip); err != nil {
		return
	}

	for _, result := range results {
		logging.VerboseLn(result)
	}

	var directory string
	if directory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
		return
	}

	if err = pkg.WriteResultsToDisk3(
		directory,
		results,
		zip,
	); err != nil {
		return
	}

	return
}

func init() {
	TidesCmd.AddCommand(retrieveCmd)

	flags.AddFlags(retrieveCmd, flags.PlatformLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName, flags.ApiVersion)
	flags.AddFlags(retrieveCmd, flags.Pages)
	flags.AddFlags(retrieveCmd, flags.NoZip)
}
