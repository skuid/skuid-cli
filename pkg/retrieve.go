package pkg

import (
	"fmt"
	"iter"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bobg/go-generics/v4/set"
	"github.com/bobg/go-generics/v4/slices"
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

type NlxRetrievalResults struct {
	MetadataService  *NlxRetrievalResult
	CloudDataService *NlxRetrievalResult
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
	// See https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	//
	// TODO: Adjust based on answer to above and/or https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	if plans.MetadataService == nil {
		return nil, fmt.Errorf("unexpected retrieval plan(s) received, expected a %v plan but did not receive one, %v", logging.QuoteText(PlanNamePliny), logging.FileAnIssueText)
	}

	// Skuid Review Required - The code in v0.6.7 would "sync" the filter.Since w/ the Since value of the
	// plan(s) retrieved (although it would not sync it correctly in all cases), however the server
	// should return the value we expect and if it doesn't, then it's an unexpected response.  See comments
	// in validateSince function for more details. See https://github.com/skuid/skuid-cli/issues/233
	//
	// TODO: eliminate the validateSince completely as it shouldn't even be necessary or adjust based on
	// answer to above
	//
	// pliny and warden are supposed to give the since value back for the retrieve, but just in case...
	if err := validateSince(filter, plans); err != nil {
		return nil, err
	}

	logger = logger.WithSuccess()
	return plans, nil
}

// TODO: This can be made private once improvements are made to address https://github.com/skuid/skuid-cli/issues/166 (e.g., test coverage, dependency injection to enable proper unit tests, etc.)
func ExecuteRetrievePlan(auth *Authorization, plans *NlxPlans) (results *NlxRetrievalResults, err error) {
	planNames := logging.CSV(slices.Values(plans.PlanNames))
	message := "Executing retrieval plan(s)"
	fields := logging.Fields{
		"planNames": planNames,
	}
	logger := logging.WithTracking("pkg.ExecuteRetrieval", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	// Skuid Review Required - v0.6.7 simply checks for != nil and skips MetadataService but will process CloudDataService if present.  Based on my testing and what I've experienced, it would
	// seem that a MetadataService should always be present, even if it is empty (e.g., a filter was applied and no results that matched it).  Given this, modifying the logic to explicitly
	// require MetadataService plan (even if empty inside) and fail if not present.  Is it correct that MetadataService should always be present in all situations?  Should a CLoudDataService
	// plan ever be processed if MetadataService isn't present? See https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	//
	// TODO: Adjust based on answer to above and/or to https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	if plans.MetadataService == nil {
		return nil, fmt.Errorf("unable to execute retrieval plan(s), expected a %v plan but did not receive one, %v", logging.QuoteText(PlanNamePliny), logging.FileAnIssueText)
	}

	allEntityPaths := logging.CSV(plans.EntityPaths.All())
	logger.WithFields(logging.Fields{
		"entities":     allEntityPaths,
		"entitiesFrom": "Execute retrieval plan(s) " + planNames,
	}).Debugf("Requesting entities %v", logging.ColorResource.Text(allEntityPaths))

	results = &NlxRetrievalResults{}
	// has to be pliny, then warden
	if result, err := executeRetrievePlan(auth, plans.MetadataService, logger); err != nil {
		return nil, err
	} else {
		results.MetadataService = result
	}

	if plans.CloudDataService != nil {
		if result, err := executeRetrievePlan(auth, plans.CloudDataService, logger); err != nil {
			return nil, err
		} else {
			results.CloudDataService = result
		}
	}

	resultPlanNames := logging.CSV(getResultPlanNames(results))
	logger = logger.WithSuccess(logging.Fields{"resultPlanNames": resultPlanNames})
	return results, nil
}

func writeRetrievePlan(targetDirectory string, results *NlxRetrievalResults) (allEntityPaths set.Of[string], err error) {
	resultPlanNames := logging.CSV(getResultPlanNames(results))
	message := "Writing retrieval result(s) to disk"
	fields := logging.Fields{
		"targetDirectory": targetDirectory,
		"resultPlanNames": resultPlanNames,
	}
	logger := logging.WithTracking("", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	allEntityPaths = set.New[string]()

	// Skuid Review Required - The code in v0.6.7 had a []NlxRetrieveResult parameter and would iterate it and process
	// results in array order.  This could lead to unexpected final state of the written files because if the order
	// of the items in the array ever changed (intentionally or unintentionally) the files would be written in a different
	// order and then for those files that are processed through CombineJSON, the final file potentially different.  To
	// avoid any intentional/unintentional outcomes, based on my understanding and the way v0.6.7 appeared to "expect"
	// things to work was that the metadata service result would always be written first, followed by the data service
	// result.  The below assumes that is the expectation and the correct way the results should be written based on the
	// way CombineJSON works.  Is this correct?  Which file should be written first?
	// See https://github.com/skuid/skuid-cli/issues/227
	//
	// TODO: Adjust below based on answer to above and/or https://github.com/skuid/skuid-cli/issues/227
	if err := writePlanResultToDisk(targetDirectory, results.MetadataService, allEntityPaths, logger); err != nil {
		return nil, err
	}

	if results.CloudDataService != nil {
		if err := writePlanResultToDisk(targetDirectory, results.CloudDataService, allEntityPaths, logger); err != nil {
			return nil, err
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

func writePlanResultToDisk(targetDirectory string, result *NlxRetrievalResult, allEntityPaths set.Of[string], logger *logging.Logger) error {
	writePayload := WritePayload{
		PlanName: result.Plan.Name,
		PlanData: result.Data,
	}
	planEntityPaths, err := WriteResultsToDisk(targetDirectory, writePayload)
	if err != nil {
		return err
	}

	loggerEntityPaths := logging.CSV(planEntityPaths.All())
	logger.WithFields(logging.Fields{
		"entities":     loggerEntityPaths,
		"entitiesFrom": "Execute retrieval plan " + logging.QuoteText(result.Plan.Name) + " response",
	}).Tracef("Received retrieval plan %v entities %v", logging.QuoteText(result.Plan.Name), logging.ColorResource.Text(loggerEntityPaths))
	allEntityPaths.AddSeq(planEntityPaths.All())
	return nil
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

func validateSince(planFilter *NlxPlanFilter, plans *NlxPlans) error {
	// Skuid Review Required - The code in v0.6.7 only "synced" the values here when `--since` was specified, however
	// if any other filter flag was specified (e.g., --app), a planFilter would be constructed and it would contain
	// a since value (the zero value of time.Time) so technically there is a "since" value provided to the server
	// even though its value didn't come from user provided since flag.  Regardless, if the response from the server
	// for "since" isn't equal to the value that we sent, that is an error/unexpected condition and we shouldn't just
	// update the "plans" returned from the server with the value that we expected to receive.  It indicates that the
	// server just simply didn't return the correct value or possibly, didn't even apply the since filter correctly.
	// Either way, we should error here, not adjust the value.  The broader question is how far do we go validating
	// every single server response?  Why did the "sync since" code even exist in v0.6.7?  The code didn't validate "appSpecific"
	// or any other field returned from the server, only "sync" so why just "since"?  At some point, the server needs to be
	// reliable which, unfortunately, as we know from many issues currently present in the repo is not the case.  Given this,
	// instead of "syncing" the value, if the value returns in the plans does not match the plan filter, an error is now returned.
	// See https://github.com/skuid/skuid-cli/issues/233
	//
	// TODO: Per above, eliminate the validation completely as it shouldn't even be necessary if the server APIs can be "trusted" or
	// adjust validation based on answers to above.  If validation remains, then other fields should be validated as well
	// (e.g., appSpecific) since there is nothing different response wise between "since" and "appSpecific" (or any other
	// field for that matter).
	validateTime := func(plan *NlxPlan, since *time.Time) error {
		if plan.Since == nil && since == nil {
			return nil
		}

		if plan.Since == nil && since != nil || plan.Since != nil && since == nil || !plan.Since.Equal(*since) {
			return fmt.Errorf("plan %v since value %v did not match plan filter since value %v, %v", logging.QuoteText(plan.Name), logging.QuoteText(logging.FormatTime(plan.Since)), logging.QuoteText(logging.FormatTime(since)), logging.FileAnIssueText)
		}
		return nil
	}
	var expectedSince *time.Time = nil
	if planFilter != nil {
		expectedSince = planFilter.Since
	}

	if err := validateTime(plans.MetadataService, expectedSince); err != nil {
		return err
	}

	if plans.CloudDataService != nil {
		if err := validateTime(plans.CloudDataService, expectedSince); err != nil {
			return err
		}
	}

	return nil
}

func getResultPlanNames(results *NlxRetrievalResults) iter.Seq[PlanName] {
	names := []PlanName{results.MetadataService.Plan.Name}
	if results.CloudDataService != nil {
		names = append(names, results.CloudDataService.Plan.Name)
	}
	return slices.Values(names)
}
