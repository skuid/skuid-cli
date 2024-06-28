package cmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dlclark/regexp2"
	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/cmd/common"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var (
	retrieveCmd = &cobra.Command{
		SilenceUsage:      true,
		Example:           "retrieve -u myUser -p myPassword --host my-site.skuidsite.com --dir ./retrieval --since 4h",
		Use:               "retrieve",
		Short:             "Retrieve a Skuid NLX Site",
		Long:              "Retrieve Skuid metadata from a Skuid NLX Site and output it into a local directory",
		PersistentPreRunE: common.PrerunValidation,
		RunE:              Retrieve,
	}
)

// stringclean makes sure string contains only letters, digits, or "."
func stringClean(str string) string {
	return strings.Map(func(r rune) rune {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.') {
			return -1
		}
		return r
	}, str)
}

func regexp2FindAllMatch(re *regexp2.Regexp, s string) []*regexp2.Match {
	var matches []*regexp2.Match
	m, _ := re.FindStringMatch(s)
	for m != nil {
		matches = append(matches, m)
		m, _ = re.FindNextMatch(m)
	}
	return matches
}

func Retrieve(cmd *cobra.Command, _ []string) (err error) {
	fields := make(logrus.Fields)
	start := time.Now()
	fields["process"] = "retrieve"
	fields["start"] = start

	logging.Get().Info(color.Green.Sprint("Starting Retrieve"))
	// get required arguments
	host, err := cmd.Flags().GetString(flags.PlinyHost.Name)
	if err != nil {
		return
	}
	username, err := cmd.Flags().GetString(flags.Username.Name)
	if err != nil {
		return
	}
	password, err := cmd.Flags().GetString(flags.Password.Name)
	if err != nil {
		return
	}

	fields["host"] = host
	fields["username"] = username
	logging.WithFields(fields).Debug("Credentials gathered")

	var auth *pkg.Authorization
	if auth, err = pkg.Authorize(host, username, password); err != nil {
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

	var sinceStr string
	since := time.Now()
	hasSince := false
	if sinceStr, err = cmd.Flags().GetString(flags.Since.Name); err != nil {
		return err
	} else if len(sinceStr) > 0 {
		// First try to parse something like "01/02 03:04:05PM '06 -0700"
		if parseTry, err := time.ParseInLocation(constants.DefaultTimeFormat, sinceStr, time.Local); err == nil {
			hasSince = true
			since = parseTry
		}
		// Next try to parse with the other layout constants in time
		if !hasSince {
			for _, layout := range constants.TimeFormatStrings {
				//sinceStrHyphen := strings.ReplaceAll(sinceStr, "/", "-")
				if parseTry, err := time.ParseInLocation(layout, sinceStr, time.Local); err == nil {
					hasSince = true
					if layout == time.Stamp { // this layout includes month and day, but year will be "0000"
						parseTry = parseTry.AddDate(since.Year(), 0, 0)
					}
					since = parseTry
					break
				}
			}
		}
		// Next try to parse as a timespan like "2days3hours" or "2d3h"
		if !hasSince {
			// First deal with capital 'M' month
			for _, alias := range constants.TimeUnits["M"] {
				if strings.Contains(sinceStr, alias) {
					sinceStr = strings.ReplaceAll(sinceStr, alias, "M")
				}
			}
			// remove everything but digits, letters, and '.'
			lsinceStr := stringClean(sinceStr)
			for k, aliases := range constants.TimeUnits {
				for _, alias := range aliases {
					if strings.Contains(lsinceStr, alias) {
						sinceStr = strings.ReplaceAll(lsinceStr, alias, k)
					}
				}
			}
			spanr, err := regexp2.Compile(`(?P<timeunit>\d+(?:\.\d+)?[smhdMy])(?=[\d\s]+|$)`, regexp2.RE2)
			if err != nil {
				return err
			}
			//for _, match := range spanr.FindAllString(sinceStr, -1) {
			for _, match := range regexp2FindAllMatch(spanr, sinceStr) {
				submatch := match.String()
				lc := len(submatch) - 1
				timeQuant, err := strconv.ParseFloat(submatch[:lc], 64)
				if err != nil {
					continue
				}
				hasSince = true
				timeInt := int(math.Abs(math.Round(timeQuant))) * -1
				switch submatch[lc:] {
				case "s", "m", "h":
					timeDur, err := time.ParseDuration(submatch)
					if err != nil {
						continue
					}
					since = since.Add(-1.0 * timeDur)
				case "d":
					since = since.AddDate(0, 0, timeInt)
				case "M":
					since = since.AddDate(0, timeInt, 0)
				case "y":
					since = since.AddDate(timeInt, 0, 0)
				}
			}
		}
		if !hasSince {
			logging.WithFields(fields).Fatal("A --since option was provided but was not parsable as a time or duration")
			return
		}
	}
	if hasSince {
		now := time.Now()
		// If the user specifies just 14:30:05 then the date is 0000-01-01, but no site is almost one year older than Jesus.
		if since.Year() == 0 {
			since = since.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		}
		if since.Compare(now) > 0 {
			logging.WithFields(fields).Fatal("A --since option in the future was specified. This is probably not what you mean.")
			return
		}

		initFilter()
		since = since.UTC()
		filter.Since = since
		sinceStr = since.Format(time.RFC3339)
		logging.WithFields(fields).Info(fmt.Sprintf("retrieving metadata records updated since: %s", sinceStr))
	}

	logging.WithFields(fields).Info("Getting Retrieve Plan")

	var plans pkg.NlxPlanPayload
	if _, plans, err = pkg.GetRetrievePlan(auth, filter); err != nil {
		return
	}

	logging.WithFields(fields).Info("Got Retrieve Plan")

	// pliny and warden are supposed to give the since value back for the retrieve, but just in case...
	if hasSince {
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
	if directory, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
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

func init() {
	flags.AddFlags(retrieveCmd, flags.NLXLoginFlags...)
	flags.AddFlags(retrieveCmd, flags.Directory, flags.AppName)
	flags.AddFlags(retrieveCmd, flags.Pages)
	flags.AddFlags(retrieveCmd, flags.Since)
	AppCmd = append(AppCmd, retrieveCmd)
}
