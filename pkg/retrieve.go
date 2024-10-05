package pkg

import (
	"fmt"
	"iter"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bobg/go-generics/v4/set"
	"github.com/bobg/go-generics/v4/slices"
	"github.com/bobg/seqs"
	"github.com/goccy/go-json"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var (
	RetrievePlanRoute = fmt.Sprintf("api/%v/metadata/retrieve/plan", DefaultApiVersion)
)

type NlxRetrievalResult struct {
	Plan *NlxPlan
	Data []byte
}

type RetrieveOptions struct {
	Auth            *Authorization
	NoClean         bool
	PlanFilter      *NlxPlanFilter
	Since           *time.Time
	TargetDirectory string
}

func Retrieve(options RetrieveOptions) (err error) {
	message := fmt.Sprintf("Retrieving site %v to directory %v", logging.ColorResource.Text(options.Auth.Host), logging.ColorResource.QuoteText(options.TargetDirectory))
	fields := logging.Fields{
		"host":            options.Auth.Host,
		"targetDirectory": options.TargetDirectory,
		"noClean":         options.NoClean,
		"since":           options.Since,
		"planFilter":      fmt.Sprintf("%+v", options.PlanFilter),
	}
	logger := logging.WithTracking("pkg.Retrieve", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if !filepath.IsAbs(options.TargetDirectory) {
		return fmt.Errorf("targetDirectory %v must be an absolute path", logging.QuoteText(options.TargetDirectory))
	}

	plans, err := GetRetrievePlan(options.Auth, options.PlanFilter)
	if err != nil {
		return err
	}

	results, err := ExecuteRetrievePlan(options.Auth, plans)
	if err != nil {
		return err
	}

	if !options.NoClean {
		if err := ClearDirectories(options.TargetDirectory); err != nil {
			return err
		}
	}

	entityPaths, err := writeRetrievePlan(options.TargetDirectory, results)
	if err != nil {
		return err
	}

	entitiesWritten := len(entityPaths)
	logger = logger.WithField("entityPathsLen", entitiesWritten)
	if entitiesWritten == 0 {
		logger.Warn(logging.ColorWarning.Sprintf("No entities retrieved, please check any filter(s) that may have been specified, %v", flags.UseDebugMessage()))
	}

	logger = logger.WithSuccess()
	return nil
}

// TODO: This can be made private once improvements are made to address https://github.com/skuid/skuid-cli/issues/166 (e.g., test coverage, dependency injection to enable proper unit tests, etc.)
func GetRetrievePlan(auth *Authorization, filter *NlxPlanFilter) (plans *NlxPlans, err error) {
	message := "Getting retrieval plan(s)"
	fields := logging.Fields{
		"filter": fmt.Sprintf("%+v", filter),
	}
	logger := logging.WithTracking("pkg.GetRetrievePlan", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	var body []byte
	if filter != nil {
		logger.Tracef("Including %v in get retrieval plan(s) request", logging.ColorFilter.Text("plan filter"))
		if body, err = json.Marshal(filter); err != nil {
			logger.WithError(err).Debugf("Error marshalling plan filter: %+v", filter)
			return nil, fmt.Errorf("unable to convert plan filter to JSON bytes: %w", err)
		}
	}

	// this is a pliny request, so we provide the access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)

	// no matter what we want to pass application/json
	// because the application/zip is discarded by pliny
	// and warden will throw an error
	headers[HeaderContentType] = JsonContentType

	plans, err = RequestNlxPlans(fmt.Sprintf("%s/%s", auth.Host, RetrievePlanRoute), headers, body, PlanModeRetrieve)
	if err != nil {
		return nil, err
	}
	planNames := logging.CSV(slices.Values(plans.PlanNames))
	logger = logger.WithField("planNames", planNames)
	logger.Debugf("Received retrieval plan(s) %v", planNames)

	// Skuid Review Required - Based on testing, it appears that a Metadata service plan should always be present, even if there is no metadata in it.  Modifying behavior
	// to perform a validation to ensure we have a metadata service plan.  Is there any situation where a metadata service plan would not be present?  Should a Cloud data
	// service that is present be processed if there isn't a metadata service plan?
	// See https://github.com/skuid/skuid-cli/issues/225 & https://github.com/skuid/skuid-cli/issues/226.
	//
	// TODO: Adjust based on answer to above and/or https://github.com/skuid/skuid-cli/issues/225 & https://github.com/skuid/skuid-cli/issues/226
	if plans.MetadataService == nil {
		return nil, fmt.Errorf("unexpected retrieval plan(s) received, expected a %v plan but did not receive one, %v", logging.QuoteText(PlanNamePliny), logging.FileAnIssueText)
	}

	// pliny and warden are supposed to give the since value back for the retrieve, but just in case...
	syncSince(filter, plans)

	logger = logger.WithSuccess()
	return plans, nil
}

// TODO: This can be made private once improvements are made to address https://github.com/skuid/skuid-cli/issues/166 (e.g., test coverage, dependency injection to enable proper unit tests, etc.)
func ExecuteRetrievePlan(auth *Authorization, plans *NlxPlans) (results []NlxRetrievalResult, err error) {
	planNames := logging.CSV(slices.Values(plans.PlanNames))
	message := "Executing retrieval plan(s)"
	fields := logging.Fields{
		"planNames": planNames,
	}
	logger := logging.WithTracking("pkg.ExecuteRetrieval", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	allEntityPaths := logging.CSV(plans.EntityPaths.All())
	logger.WithFields(logging.Fields{
		"entities":     allEntityPaths,
		"entitiesFrom": "Execute retrieval plan(s) " + planNames,
	}).Debugf("Requesting entities %v", logging.ColorResource.Text(allEntityPaths))

	// Skuid Review Required - v0.6.7 simply checks for != nil and skips MetadataService but will process CloudDataService if present.  Based on my testing and what I've experienced, it would
	// seem that a MetadataService should always be present, even if it is empty (e.g., a filter was applied and no results that matched it).  Given this, modifying the logic to explicitly
	// require MetadataService plan (even if empty inside) and fail if not present.  Is it correct that MetadataService should always be present in all situations?  Should a CLoudDataService
	// plan ever be processed if MetadataService isn't present? See https://github.com/skuid/skuid-cli/issues/225
	//
	// TODO: Adjust based on answer to above and/or to https://github.com/skuid/skuid-cli/issues/225
	if plans.MetadataService == nil {
		return nil, fmt.Errorf("unable to execute retrieval plan(s), expected a %v plan but did not receive one, %v", logging.QuoteText(PlanNamePliny), logging.FileAnIssueText)
	}

	// has to be pliny, then warden
	if result, err := executeRetrievePlan(auth, plans.MetadataService, logger); err != nil {
		return nil, err
	} else {
		results = append(results, *result)
	}

	if plans.CloudDataService != nil {
		if result, err := executeRetrievePlan(auth, plans.CloudDataService, logger); err != nil {
			return nil, err
		} else {
			results = append(results, *result)
		}
	}

	resultPlanNames := logging.CSV(getResultPlanNames(results))
	logger = logger.WithSuccess(logging.Fields{"resultPlanNames": resultPlanNames})
	return results, nil
}

func writeRetrievePlan(targetDirectory string, results []NlxRetrievalResult) (allEntityPaths set.Of[string], err error) {
	resultPlanNames := logging.CSV(getResultPlanNames(results))
	message := "Writing retrieval result(s) to disk"
	fields := logging.Fields{
		"targetDirectory": targetDirectory,
		"resultPlanNames": resultPlanNames,
	}
	logger := logging.WithTracking("", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	allEntityPaths = set.New[string]()
	for _, v := range results {
		writePayload := WritePayload{
			PlanName: v.Plan.Name,
			PlanData: v.Data,
		}
		if planEntityPaths, err := WriteResultsToDisk(targetDirectory, writePayload); err != nil {
			return nil, err
		} else {
			loggerEntityPaths := logging.CSV(planEntityPaths.All())
			logger.WithFields(logging.Fields{
				"entities":     loggerEntityPaths,
				"entitiesFrom": "Execute retrieval plan " + logging.QuoteText(v.Plan.Name) + " response",
			}).Tracef("Received retrieval plan %v entities %v", logging.QuoteText(v.Plan.Name), logging.ColorResource.Text(loggerEntityPaths))
			allEntityPaths.AddSeq(planEntityPaths.All())
		}
	}
	loggerEntityPaths := logging.CSV(allEntityPaths.All())
	logger.WithFields(logging.Fields{
		"entities":     loggerEntityPaths,
		"entitiesFrom": "Execute retrieval plan(s) " + resultPlanNames + " result",
	}).Debugf("Received entities %v", logging.ColorResource.Text(loggerEntityPaths))

	logger = logger.WithSuccess()
	return allEntityPaths, nil
}

// this function generically handles a plan based on name / stuff
func executeRetrievePlan(auth *Authorization, plan *NlxPlan, logger *logging.Logger) (result *NlxRetrievalResult, err error) {
	url := GeneratePlanRoute(auth, plan.Name, plan.Host, plan.Port, plan.Endpoint)
	message := fmt.Sprintf("Executing retrieval plan %v", logging.QuoteText(plan.Name))
	fields := logging.Fields{
		"planName": plan.Name,
		"planType": plan.Type,
		"url":      url,
	}
	logger = logger.WithTraceTracking("executeRetrievePlan", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	headers := GeneratePlanHeaders(auth, plan.Name, plan.Host)
	headers[HeaderContentType] = JsonContentType

	reqBody, err := NewRetrievalRequestBody(plan.Metadata, plan.Since, plan.AppSpecific)
	if err != nil {
		return nil, err
	}

	planEntityPaths := logging.CSV(plan.EntityPaths.All())
	logger.WithFields(logging.Fields{
		"entities":     planEntityPaths,
		"entitiesFrom": "Execute retrieval plan " + logging.QuoteText(plan.Name) + " request",
	}).Tracef("Requesting retrieval of plan %v for entities %v", logging.QuoteText(plan.Name), logging.ColorResource.Text(planEntityPaths))
	response, err := Request(url, http.MethodPost, reqBody, headers)
	if err != nil {
		return nil, err
	}

	result = &NlxRetrievalResult{
		Plan: plan,
		Data: response,
	}
	logger = logger.WithSuccess(logging.Fields{"responseLen": len(response)})
	return result, nil
}

func syncSince(planFilter *NlxPlanFilter, plans *NlxPlans) {
	// Skuid Review Required - The prior code only "synced" the values here when `--since` was specified, however
	// if any other filter flag was specified (e.g., --app), a planFilter would be constructed and it would contain
	// a since value (the zero value of time.Time) so technically there is a "since" value provided to the server
	// even though its value didn't come from user provided since flag.  Given that, shouldn't this code sync the
	// values for since if we have a non-nil filter instead of only when --since was specified by user via flag
	// (e.g., planFilter.Since.IsZero() != true)?  In short, if any filter flag is specified, a "since" value will
	// be sent to server in the GetRetrievePlans request that will either contain Zero value of time.Time (if no
	// since was specified, or non-zero value of time if since was specified).  Since we sent a time, shouldn't we
	// always sync?  Or possibly we change NlxPlanFilter to be *time.Time to avoid writing something unless there
	// was a flag?
	//
	// TODO: Based on answer to above, adjust condition below and remove Since from RetrieveOptions as it isn't needed
	// for anything other than to ensure consistent logic with v0.6.7
	if planFilter == nil || planFilter.Since.IsZero() {
		return
	}

	sinceStr := flags.FormatSince(&planFilter.Since)
	if plans.MetadataService.Since == "" {
		plans.MetadataService.Since = sinceStr
	}
	if plans.CloudDataService != nil {
		if plans.CloudDataService.Since == "" {
			plans.CloudDataService.Since = sinceStr
		}
	}
}

func getResultPlanNames(results []NlxRetrievalResult) iter.Seq[PlanName] {
	return seqs.Map(slices.Values(results), func(r NlxRetrievalResult) PlanName {
		return r.Plan.Name
	})
}
