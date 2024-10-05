package pkg

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bobg/go-generics/v4/set"
	"github.com/bobg/go-generics/v4/slices"
	"github.com/bobg/seqs"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	DeployPlanRoute   = fmt.Sprintf("api/%v/metadata/deploy/plan", DefaultApiVersion)
	ErrArchiveNoFiles = fmt.Errorf("unable to create archive, no files were found, %v", flags.UseDebugMessage())
)

type DeployOptions struct {
	ArchiveFilter     ArchiveFilter
	Auth              *Authorization
	EntitiesToArchive []metadata.MetadataEntity
	PlanFilter        *NlxPlanFilter
	SourceDirectory   string
}

type FilteredRequestBody struct {
	AppName string `json:"appName"`
	// pages flag does not work as expected so commenting out
	// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	//PageNames     []string `json:"pageNames"`
	PlanBytes     []byte `json:"plan"`
	IgnoreSkuidDb bool   `json:"ignoreSkuidDb"`
}

type metadataTypeResult struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type PermissionSetResult struct {
	metadataTypeResult
	AppId                 string      `json:"app_id"`
	AppName               string      `json:"app_name"`
	Description           string      `json:"description"`
	OrganizationId        string      `json:"organization_id"`
	DatasourcePermissions interface{} `json:"dataSourcePermissions"`
}

type metadataTypeResults struct {
	Inserts []metadataTypeResult `json:"inserts"`
	Updates []metadataTypeResult `json:"updates"`
	Deletes []metadataTypeResult `json:"deletes"`
}

type permissionSetResults struct {
	Inserts []PermissionSetResult `json:"inserts"`
	Updates []PermissionSetResult `json:"updates"`
	Deletes []PermissionSetResult `json:"deletes"`
}

type plinyResponse struct {
	Apps               metadataTypeResults  `json:"apps"`
	AuthProviders      metadataTypeResults  `json:"authproviders"`
	ComponentPacks     metadataTypeResults  `json:"componentpacks"`
	DataServices       metadataTypeResults  `json:"dataservices"`
	DataSources        metadataTypeResults  `json:"datasources"`
	DesignSystems      metadataTypeResults  `json:"designsystems"`
	Variables          metadataTypeResults  `json:"variables"`
	Files              metadataTypeResults  `json:"files"`
	Pages              metadataTypeResults  `json:"pages"`
	PermissionSets     permissionSetResults `json:"permissionSets"`
	SitePermissionSets metadataTypeResults  `json:"sitepermissionsets"`
	SessionVariables   metadataTypeResults  `json:"sessionvariables"`
	Site               metadataTypeResults  `json:"site"`
	Themes             metadataTypeResults  `json:"themes"`
}

type wardenResponse struct{}

type deploymentResponse interface {
	wardenResponse | plinyResponse | []byte
}

type deploymentValidator interface {
	Validate() (string, bool)
	Name() string
}

type deploymentResult[T any, R deploymentResponse] struct {
	name     string
	request  deploymentRequest[T]
	response R
}

func (r *deploymentResult[T, R]) Validate() (string, bool) {
	// Skuid Review Required - Except for the metadata plan (see plinyValidator) which returns partial results, the other deployment APIs return either
	// no results at all (data plan & update permission sets) or an unknown format (sync data sources) so there is no way to validate actual vs expected
	// which is critical due to all the server API issues (e.g., https://github.com/skuid/skuid-cli/issues/163).
	// See https://github.com/skuid/skuid-cli/issues/211
	//
	// TODO:  adjust once https://github.com/skuid/skuid-cli/issues/211 is addressed and all 4 APIs return results that are reliable and in a known
	// format
	logging.
		WithName("pkg.deploymentResult::Validate", logging.Fields{"deploymentName": r.name}).
		Debugf("%v %v", logging.ColorWarning.Text("Skipping validation of result for deployment "), logging.QuoteText(r.name))
	return r.name, true
}

func (r *deploymentResult[T, R]) Name() string {
	return r.name
}

type plinyValidator struct {
	*deploymentResult[NlxPlan, plinyResponse]
}

func (r *plinyValidator) Validate() (deploymentName string, isValid bool) {
	message := fmt.Sprintf("Validating result for deployment %v", r.name)
	fields := logging.Fields{
		"deploymentName": r.name,
	}
	logger := logging.WithTraceTracking("pkg.plinyValidator::Validate", message, fields).StartTracking()
	defer func() { logger.FinishTracking(nil) }()

	// Skuid Review Required - The metadata plan does not contain the full set of entities in the result returned.  For now, we treat it
	// as if there are full results and indicate in the message that the results returned are NOT RELIABLE so the warning may be a false positive.
	// see https://github.com/skuid/skuid-cli/issues/211
	//
	// TODO: Adjust once https://github.com/skuid/skuid-cli/issues/211 is addressed
	plan := r.request.payload
	response := r.response
	planEntityPaths := plan.EntityPaths
	modifiedEntityPaths, deletedEntityPaths := getDeploymentResultEntityPaths(response)

	modifiedEqual := planEntityPaths.Equal(modifiedEntityPaths)
	hasDeleted := len(deletedEntityPaths) > 0

	if modifiedEqual && !hasDeleted {
		logger = logger.WithSuccess()
		return r.name, true
	}

	msgPrefix := logging.ColorWarning.Sprintf("Unexpected deployment result %v:", logging.QuoteText(r.name))
	if !modifiedEqual {
		planMissing := set.Diff(planEntityPaths, modifiedEntityPaths)
		modifiedMissing := set.Diff(modifiedEntityPaths, planEntityPaths)

		for _, p := range slices.Sorted(planMissing.All()) {
			// TODO: Change to Warnf once https://github.com/skuid/skuid-cli/issues/211 is addressed
			logger.Debugf("%v entity %v was not deployed but was expected", msgPrefix, logging.ColorResource.QuoteText(p))
		}
		for _, p := range slices.Sorted(modifiedMissing.All()) {
			// TODO: Change to Warnf once https://github.com/skuid/skuid-cli/issues/211 is addressed
			logger.Debugf("%v entity %v was deployed but was not expected", msgPrefix, logging.ColorResource.QuoteText(p))
		}
	}
	if hasDeleted {
		for _, p := range slices.Sorted(deletedEntityPaths.All()) {
			// TODO: Change to Warnf once https://github.com/skuid/skuid-cli/issues/211 is addressed
			logger.Debugf("%v entity %v was deleted which was not expected", msgPrefix, logging.ColorResource.QuoteText(p))
		}
	}

	return r.name, false
}

type deploymentRequest[T any] struct {
	headers RequestHeaders
	url     string
	payload *T
	body    []byte
}

// TODO: This can be made private once improvements are made to address https://github.com/skuid/skuid-cli/issues/166 (e.g., test coverage, dependency injection to enable proper unit tests, etc.)
func GetDeployPlan(auth *Authorization, sourceDirectory string, archiveFilter ArchiveFilter, entitiesToArchive []metadata.MetadataEntity, planFilter *NlxPlanFilter) (plans *NlxPlans, err error) {
	message := "Getting deployment plan(s)"
	fields := logging.Fields{
		"sourceDirectory":      sourceDirectory,
		"archiveFilterNil":     archiveFilter == nil,
		"entitiesToArchiveLen": len(entitiesToArchive),
		"planFilter":           fmt.Sprintf("%+v", planFilter),
	}
	logger := logging.WithTracking("pkg.GetDeployPlan", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if !filepath.IsAbs(sourceDirectory) {
		return nil, fmt.Errorf("sourceDirectory %v must be an absolute path", logging.QuoteText(sourceDirectory))
	}

	payload, _, archivedEntities, err := archive(sourceDirectory, archiveFilter, metadata.MetadataEntityPaths(entitiesToArchive))
	if err != nil {
		return nil, err
	}

	// pliny request, use access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)
	headers[HeaderContentType] = ZipContentType
	reqBody, err := getDeployPlanRequestBody(headers, payload, planFilter, logger)
	if err != nil {
		return nil, err
	}

	loggerEntityPaths := logging.CSV(metadata.MetadataEntityPaths(archivedEntities).All())
	logger.WithFields(logging.Fields{
		"entities":     loggerEntityPaths,
		"entitiesFrom": "Get deployment plan(s) request",
	}).Debugf("Requesting deployment plan(s) for entities %v", logging.ColorResource.Text(loggerEntityPaths))
	plans, err = RequestNlxPlans(fmt.Sprintf("%s/%s", auth.Host, DeployPlanRoute), headers, reqBody, PlanModeDeploy)
	if err != nil {
		return nil, err
	}
	planNames := logging.CSV(slices.Values(plans.PlanNames))
	logger = logger.WithField("planNames", planNames)
	logger.Debugf("Received deployment plan(s) %v", planNames)

	// Skuid Review Required - Based on testing, it appears that a Metadata service plan should always be present, even if there is no metadata in it.  Modifying behavior
	// to perform a validation to ensure we have a metadata service plan.  Is there any situation where a metadata service plan would not be present?  Should a Cloud data
	// service that is present be processed if there isn't a metadata service plan?
	// See https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	//
	// TODO: Adjust based on answer to above and/or https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	if plans.MetadataService == nil {
		return nil, fmt.Errorf("unexpected deployment plan(s) received, expected a %v plan but did not receive one, %v", logging.QuoteText(PlanNamePliny), logging.FileAnIssueText)
	}

	if err := validateDeployPlans(plans, archivedEntities, logger); err != nil {
		// intentionally ignoring any error as currently using validation for logging purposes only
		// TODO: This could be potentially relaxed to Debug/Trace level once all the known server-side API
		//       issues are addressed and things are more reliable in general.  For example, attempting
		//       to upload a Files entity with a name that starts with a period (e.g., .my-file.txt) will
		//       result in the deployment plan not containing that file even though the web ui allows
		//       it (see https://github.com/skuid/skuid-cli/issues/158).
		logger.Warn(logging.ColorWarning.Text(err.Error()))
	}

	logger = logger.WithSuccess()
	return plans, nil
}

func Deploy(options DeployOptions) (err error) {
	message := fmt.Sprintf("Deploying site %v from directory %v", logging.ColorResource.Text(options.Auth.Host), logging.ColorResource.QuoteText(options.SourceDirectory))
	fields := logging.Fields{
		"host":                 options.Auth.Host,
		"sourceDirectory":      options.SourceDirectory,
		"archiveFilterNil":     options.ArchiveFilter == nil,
		"entitiesToArchiveLen": len(options.EntitiesToArchive),
		"planFilter":           fmt.Sprintf("%+v", options.PlanFilter),
	}
	logger := logging.WithTracking("pkg.Deploy", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if !filepath.IsAbs(options.SourceDirectory) {
		return fmt.Errorf("sourceDirectory %v must be an absolute path", logging.QuoteText(options.SourceDirectory))
	}

	plans, err := GetDeployPlan(options.Auth, options.SourceDirectory, options.ArchiveFilter, options.EntitiesToArchive, options.PlanFilter)
	if err != nil {
		return err
	}

	results, err := ExecuteDeployPlan(options.Auth, plans, options.SourceDirectory)
	if err != nil {
		return err
	}

	logger = logger.WithSuccess(logging.Fields{
		"resultsLen": len(results),
	})
	return nil
}

// ExecuteDeployPlan executes a map of plan items in a deployment plan
// A summary of its steps:
// 1. Deploy metadata plan aka pliny resources
// 2. After pliny is deployed, take the results from that deploy and pull out the app permission set ids
// (and their datasource permissions) that were inserted / updated
// 3. Deploy data plan aka warden
// 4. After its deployed take the app permission set ids from the pliny deploy and deploy those permission sets to warden
// 5. If metadata and data was deployed, send a request to pliny to sync its datasources' external_ids with warden datasource
// ids, in case they changed during the deploy
// TODO: This can be made private once improvements are made to address https://github.com/skuid/skuid-cli/issues/166 (e.g., test coverage, dependency injection to enable proper unit tests, etc.)
func ExecuteDeployPlan(auth *Authorization, plans *NlxPlans, sourceDirectory string) (deploymentResults []deploymentValidator, err error) {
	planNames := logging.CSV(slices.Values(plans.PlanNames))
	message := "Executing deployment plan(s)"
	fields := logging.Fields{
		"planNames":       planNames,
		"sourceDirectory": sourceDirectory,
	}
	logger := logging.WithTracking("pkg.ExecuteDeployPlan", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if !filepath.IsAbs(sourceDirectory) {
		return nil, fmt.Errorf("sourceDirectory %v must be an absolute path", logging.QuoteText(sourceDirectory))
	}

	// Skuid Review Required - v0.6.7 would check if both plans were nil and if so, Tracef and return resulting in the user thinking the deployment was successful but nothing actually happened.  Also,
	// if metadata service was nil but cloud data service was not, cloud data service would be processed.  Based on my testing and what I've experienced, it would
	// seem that a MetadataService should always be present, even if it is empty and some code that exists in v0.6.7 assumes it's there and some code doesn't assume leading me to believe its always
	// expected.  Given this, modifying the logic to explicitly require MetadataService plan (even if empty inside) and fail if not present.  Is it correct that MetadataService should always be present
	// in all situations?  Should a CloudDataService plan ever be processed if MetadataService isn't present?
	// See https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	//
	// TODO: Adjust based on answer to above and/or to https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	if plans.MetadataService == nil {
		return nil, fmt.Errorf("unable to execute deployment plan(s), expected a %v plan but did not receive one, %v", logging.QuoteText(PlanNamePliny), logging.FileAnIssueText)
	}

	allEntityPaths := logging.CSV(plans.EntityPaths.All())
	logger.WithFields(logging.Fields{
		"entities":     allEntityPaths,
		"entitiesFrom": "Execute deployment plan(s) " + planNames,
	}).Debugf("Deploying entities %v", logging.ColorResource.Text(allEntityPaths))

	// initialize to max of what we expect
	deploymentResults = make([]deploymentValidator, 0, 4)

	metaResult, err := executeDeploymentPlan[plinyResponse](auth, plans.MetadataService, sourceDirectory, logger)
	if err != nil {
		return nil, err
	}
	deploymentResults = append(deploymentResults, &plinyValidator{metaResult})

	if plans.CloudDataService != nil {
		// Skuid Review Required - See comment/question in syncPermissionSets function
		// TODO: Adjust based on answer to above
		syncPermissionSets(metaResult.response, plans.CloudDataService)
		dataResult, err := executeDeploymentPlan[wardenResponse](auth, plans.CloudDataService, sourceDirectory, logger)
		if err != nil {
			return nil, err
		}
		deploymentResults = append(deploymentResults, dataResult)

		if len(plans.CloudDataService.AllPermissionSets) > 0 {
			if result, err := updatePermissionSets(auth, plans.CloudDataService, logger); err != nil {
				return nil, err
			} else {
				deploymentResults = append(deploymentResults, result)
			}
		}

		// Tell pliny to sync datasource external_id field with warden ids
		if result, err := syncDataSources(auth, plans.MetadataService, logger); err != nil {
			return nil, err
		} else {
			deploymentResults = append(deploymentResults, result)
		}
	}

	if logger.DiagEnabled() {
		// deployment results are unreliable and will lead to false positives so only inspect them when debug is enabled for perf reasons
		// intentionally ignoring any error as currently using validation for logging purposes only
		// TODO: when Issue #211 is addressed server side, DebugEnable check should be removed and the error could be acted upon as appropriate
		// see https://github.com/skuid/skuid-cli/issues/211
		if vErr := validateDeployResults(deploymentResults, logger); vErr != nil {
			logger.Warnf("%v %v", logging.ColorWarning.Sprintf("Deployment results do not match expected results, however the deployment results are NOT RELIABLE so there may be false positives:"), vErr)
		}
	}

	deploymentResultNames := logging.CSV(seqs.Map(slices.Values(deploymentResults), func(v deploymentValidator) string {
		return v.Name()
	}))
	logger = logger.WithSuccess(logging.Fields{"deploymentResultNames": deploymentResultNames})
	return deploymentResults, nil
}

func syncDataSources(auth *Authorization, metaPlan *NlxPlan, logger *logging.Logger) (result *deploymentResult[[]byte, []byte], err error) {
	message := "Syncing Data Sources"
	logger = logger.WithTraceTracking("pkg.syncDataSources", message).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	payload := []byte{}
	req := deploymentRequest[[]byte]{
		payload: &payload,
		body:    payload,
		headers: GeneratePlanHeaders(auth, metaPlan.Name, metaPlan.Host),
		url:     GeneratePlanRoute(auth, metaPlan.Name, metaPlan.Host, metaPlan.Port, "/metadata/deploy/sync"),
	}
	result, err = executeDeployment[[]byte, []byte]("Sync Data Sources", req, logger)
	if err != nil {
		return nil, err
	}

	logger = logger.WithSuccess()
	return result, nil
}

func updatePermissionSets(auth *Authorization, dataPlan *NlxPlan, logger *logging.Logger) (result *deploymentResult[[]PermissionSetResult, []byte], err error) {
	payload := dataPlan.AllPermissionSets
	permissionSetNames := logging.CSV(seqs.Map(slices.Values(payload), func(p PermissionSetResult) string {
		return resolveEntityName(p.metadataTypeResult)
	}))
	message := "Updating Permission Sets"
	fields := logging.Fields{
		"permissionSetNames": permissionSetNames,
	}
	logger = logger.WithTraceTracking("pkg.updatePermissionSets", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	reqBody, err := json.Marshal(payload)
	if err != nil {
		logger.WithError(err).Debugf("Error marshalling update permission sets request payload: %+v", payload)
		return nil, err
	}

	req := deploymentRequest[[]PermissionSetResult]{
		payload: &payload,
		body:    reqBody,
		headers: GeneratePlanHeaders(auth, dataPlan.Name, dataPlan.Host),
		url:     GeneratePlanRoute(auth, dataPlan.Name, dataPlan.Host, dataPlan.Port, "/metadata/update-permissionsets"),
	}
	logger.WithFields(logging.Fields{
		"entities":     permissionSetNames,
		"entitiesFrom": "Update Permission Sets request",
	}).Tracef("Requesting update of Permission Sets %v", logging.ColorResource.Text(permissionSetNames))
	result, err = executeDeployment[[]PermissionSetResult, []byte]("Update Permission Sets", req, logger)
	if err != nil {
		return nil, err
	}
	// TODO: Log result either here or within Validate() once https://github.com/skuid/skuid-cli/issues/211 is addressed

	logger = logger.WithSuccess()
	return result, nil
}

func syncPermissionSets(result plinyResponse, dataPlan *NlxPlan) {
	// Skuid Review Required - In v0.6.7, NlxPlan contained an AllPermissionSets property that was set on the dataPlan when there is a metadata
	// plan.  The AllPermissionSets property would then be sent in the payload for the "execute dataPlan" request and then extracted and sent
	// in the "/metadata/update-permissionsets" request.  The question here is does the server expect AllPermissionSets property when executing
	// the request for the dataPlan itself, or was the property just added to NlxPlan for "convenience" and only the "update-permissionsets"
	// api actually needs it?
	//
	// TODO: If the answer to above is that only "update-permissionsets" needs AllPermissionSets, refactor the code to not modify the dataPlan
	// itself and just use the metaPlan directly when building the "update-permissionsets" payload.  If the "execute dataPlan" API itself
	// expects AllPermissionSets property, then everything can stay as-is.
	//
	// Collect all app permission set UUIDs and datasource permissions from pliny so we can use them to create
	// datasource permissions in warden. This is necessary because we need the UUIDs and the deploy plan only has the names
	insertLen := len(result.PermissionSets.Inserts)
	updateLen := len(result.PermissionSets.Updates)
	dataPlan.AllPermissionSets = make([]PermissionSetResult, insertLen+updateLen)
	i := 0
	for _, v := range result.PermissionSets.Inserts {
		dataPlan.AllPermissionSets[i] = v
		i++
	}
	for _, v := range result.PermissionSets.Updates {
		dataPlan.AllPermissionSets[i] = v
		i++
	}
}

func executeDeployment[T any, R deploymentResponse](deploymentName string, req deploymentRequest[T], logger *logging.Logger) (result *deploymentResult[T, R], err error) {
	message := fmt.Sprintf("Executing deployment %v", logging.QuoteText(deploymentName))
	fields := logging.Fields{
		"deploymentName": deploymentName,
	}
	logger = logger.WithTraceTracking("pkg.executeDeployment", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	respBody, err := Request(req.url, http.MethodPost, req.body, req.headers)
	if err != nil {
		return nil, err
	}

	var response R
	switch r := any(&response).(type) {
	case *wardenResponse, *plinyResponse:
		if err := json.Unmarshal(respBody, r); err != nil {
			return nil, fmt.Errorf("unable to parse JSON returned from executing plan %v: %w", logging.QuoteText(deploymentName), err)
		}
	case *[]byte:
		*r = respBody
	default:
		// should not happen in production
		panic(fmt.Errorf("unexpected type %T for deployment request", r))
	}

	result = &deploymentResult[T, R]{
		name:     deploymentName,
		request:  req,
		response: response,
	}
	logger = logger.WithSuccess()
	return result, nil
}

func executeDeploymentPlan[R deploymentResponse](auth *Authorization, plan *NlxPlan, sourceDirectory string, logger *logging.Logger) (result *deploymentResult[NlxPlan, R], err error) {
	message := fmt.Sprintf("Executing deployment for plan %v", logging.QuoteText(plan.Name))
	fields := logging.Fields{
		"planName": plan.Name,
		"planType": plan.Type,
	}
	logger = logger.WithTraceTracking("pkg.executeDeployPlan", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	reqBody, _, archivedEntities, err := archive(sourceDirectory, MetadataArchiveFilter(&plan.Metadata), plan.EntityPaths)
	if err != nil {
		return nil, err
	}

	for _, warning := range plan.Warnings {
		logger.Warn(logging.ColorWarning.Sprintf("%v plan warning: %v", plan.Name, warning))
	}

	req := deploymentRequest[NlxPlan]{
		payload: plan,
		body:    reqBody,
		headers: GeneratePlanHeaders(auth, plan.Name, plan.Host),
		url:     GeneratePlanRoute(auth, plan.Name, plan.Host, plan.Port, plan.Endpoint),
	}
	loggerEntityPaths := logging.CSV(metadata.MetadataEntityPaths(archivedEntities).All())
	logger.WithFields(logging.Fields{
		"entities":     loggerEntityPaths,
		"entitiesFrom": "Execute deployment plan " + logging.QuoteText(plan.Name) + " request",
	}).Tracef("Requesting deployment of plan %v for entities %v", logging.QuoteText(plan.Name), logging.ColorResource.Text(loggerEntityPaths))
	result, err = executeDeployment[NlxPlan, R](string(plan.Name), req, logger)
	if err != nil {
		return nil, err
	}
	// TODO: Log result either here or within Validate() once https://github.com/skuid/skuid-cli/issues/211 is addressed

	logger = logger.WithSuccess()
	return result, nil
}

func archive(sourceDirectory string, archiveFilter ArchiveFilter, entitiesToArchive set.Of[string]) ([]byte, []string, []metadata.MetadataEntity, error) {
	if deploymentRequest, archivedFilePaths, archivedEntities, err := Archive(os.DirFS(sourceDirectory), util.NewFileUtil(), archiveFilter); err != nil {
		return deploymentRequest, archivedFilePaths, archivedEntities, err
	} else if err = validateArchive(entitiesToArchive, archivedEntities); err != nil {
		return deploymentRequest, archivedFilePaths, archivedEntities, err
	} else {
		return deploymentRequest, archivedFilePaths, archivedEntities, nil
	}
}

func validateArchive(expectedEntityPaths set.Of[string], actualEntities []metadata.MetadataEntity) (err error) {
	message := "Validating archive"
	fields := logging.Fields{
		"expectedEntityPathsLen": len(expectedEntityPaths),
		"actualEntitiesLen":      len(actualEntities),
	}
	logger := logging.WithTraceTracking("pkg.validateArchive", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	// if no expected entities were specified (e.g., --entities flag, archiving entities from deploy plan) we're unable
	// to validate the entities themselves since we should have gathered all files on disk.  In that case, we only ensure
	// that we have something in the archive.
	if len(expectedEntityPaths) == 0 {
		if len(actualEntities) == 0 {
			return ErrArchiveNoFiles
		}
		logger = logger.WithSuccess()
		return nil
	}

	actualEntityPaths := metadata.MetadataEntityPaths(actualEntities)
	if expectedEntityPaths.Equal(actualEntityPaths) {
		logger = logger.WithSuccess()
		return nil
	}

	expectedMissing := set.Diff(expectedEntityPaths, actualEntityPaths)
	actualMissing := set.Diff(actualEntityPaths, expectedEntityPaths)

	msgPrefix := logging.ColorWarning.Sprintf("Unable to prepare archive:")
	for _, p := range slices.Sorted(expectedMissing.All()) {
		logger.Warnf("%v requested entity %v was not found", msgPrefix, logging.ColorResource.QuoteText(p))
	}
	for _, p := range slices.Sorted(actualMissing.All()) {
		logger.Warnf("%v includes entity %v which was not requested", msgPrefix, logging.ColorResource.QuoteText(p))
	}
	return fmt.Errorf("unable to prepare archive, %v", flags.UseDebugMessage())
}

func validateDeployResults(results []deploymentValidator, logger *logging.Logger) (err error) {
	deploymentNames := logging.CSV(seqs.Map(slices.Values(results), func(v deploymentValidator) string {
		return v.Name()
	}))
	message := fmt.Sprintf("Validating deployment result(s) %v", deploymentNames)
	fields := logging.Fields{
		"resultsLen":      len(results),
		"deploymentNames": deploymentNames,
	}
	logger = logger.WithTraceTracking("validateDeployResults", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	var invalidDeploymentNames []string
	for _, result := range results {
		if deploymentName, ok := result.Validate(); !ok {
			invalidDeploymentNames = append(invalidDeploymentNames, deploymentName)
		}
	}

	if len(invalidDeploymentNames) > 0 {
		return fmt.Errorf("unexpected deployment result(s) %v, %v", logging.CSV(slices.Values(invalidDeploymentNames)), flags.UseDebugMessage())
	}

	logger = logger.WithSuccess()
	return nil
}

func resolveEntityName(r metadataTypeResult) string {
	// Skuid Review Required - Apps in the response payload do not include the name property
	// see https://github.com/skuid/skuid-cli/issues/211
	if len(strings.TrimSpace(r.Name)) > 0 {
		return r.Name
	}

	if len(strings.TrimSpace(r.Id)) > 0 {
		return r.Id
	}

	return "noname-" + uuid.NewString()
}

func getDeploymentResultEntityPaths(result plinyResponse) (set.Of[string], set.Of[string]) {
	modifiedEntityPaths := set.New[string]()
	deletedEntityPaths := set.New[string]()

	for _, mdt := range metadata.MetadataTypes.Members() {
		typeResults := getTypeResults(mdt, result)
		modifiedEntityPaths.AddSeq(seqs.Map(slices.Values(typeResults.Inserts), func(r metadataTypeResult) string {
			return getMetadataTypeResultEntityPaths(mdt, r)
		}))
		modifiedEntityPaths.AddSeq(seqs.Map(slices.Values(typeResults.Updates), func(r metadataTypeResult) string {
			return getMetadataTypeResultEntityPaths(mdt, r)
		}))
		deletedEntityPaths.AddSeq(seqs.Map(slices.Values(typeResults.Deletes), func(r metadataTypeResult) string {
			return getMetadataTypeResultEntityPaths(mdt, r)
		}))
	}

	return modifiedEntityPaths, deletedEntityPaths
}

func getMetadataTypeResultEntityPaths(mdt metadata.MetadataType, r metadataTypeResult) string {
	return path.Join(mdt.DirName(), resolveEntityName(r))
}

func getTypeResults(mdt metadata.MetadataType, result plinyResponse) metadataTypeResults {
	switch mdt {
	case metadata.MetadataTypeApps:
		return result.Apps
	case metadata.MetadataTypeAuthProviders:
		return result.AuthProviders
	case metadata.MetadataTypeComponentPacks:
		return result.ComponentPacks
	case metadata.MetadataTypeDataServices:
		return result.DataServices
	case metadata.MetadataTypeDataSources:
		return result.DataSources
	case metadata.MetadataTypeDesignSystems:
		return result.DesignSystems
	case metadata.MetadataTypeVariables:
		return result.Variables
	case metadata.MetadataTypeFiles:
		return result.Files
	case metadata.MetadataTypePages:
		return result.Pages
	case metadata.MetadataTypePermissionSets:
		inserted := slices.Map(result.PermissionSets.Inserts, func(r PermissionSetResult) metadataTypeResult {
			return r.metadataTypeResult
		})
		updated := slices.Map(result.PermissionSets.Updates, func(r PermissionSetResult) metadataTypeResult {
			return r.metadataTypeResult
		})
		deleted := slices.Map(result.PermissionSets.Deletes, func(r PermissionSetResult) metadataTypeResult {
			return r.metadataTypeResult
		})
		return metadataTypeResults{inserted, updated, deleted}
	case metadata.MetadataTypeSitePermissionSets:
		return result.SitePermissionSets
	case metadata.MetadataTypeSessionVariables:
		return result.SessionVariables
	case metadata.MetadataTypeSite:
		return result.Site
	case metadata.MetadataTypeThemes:
		return result.Themes
	default:
		// should not happen in production
		panic(fmt.Errorf("unexpected metadata type encountered while obtaining deployment result"))
	}
}

func validateDeployPlans(plans *NlxPlans, archivedEntities []metadata.MetadataEntity, logger *logging.Logger) (err error) {
	planNames := logging.CSV(slices.Values(plans.PlanNames))
	message := "Validating deployment plan(s)"
	fields := logging.Fields{
		"planNames":           planNames,
		"archivedEntitiesLen": len(archivedEntities),
	}
	logger = logger.WithTraceTracking("validateDeployPlans", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	var invalidPlanNames []PlanName
	archivedEntityPaths := metadata.MetadataEntityPaths(archivedEntities)
	for _, plan := range plans.Plans {
		if planName, ok := validateDeployPlan(plan, archivedEntityPaths, logger); !ok {
			invalidPlanNames = append(invalidPlanNames, planName)
		}
	}

	if len(invalidPlanNames) != 0 {
		return fmt.Errorf("unexpected deployment plan(s) %v, possibly due to a filter and may be benign, %v", logging.CSV(slices.Values(invalidPlanNames)), flags.UseDebugMessage())
	}

	logger = logger.WithSuccess()
	return nil
}

func validateDeployPlan(plan *NlxPlan, archivedEntityPaths set.Of[string], logger *logging.Logger) (PlanName, bool) {
	logger = logger.WithName("validateDeployPlan", logging.Fields{"planName": plan.Name})
	logger.Tracef("Validating deployment plan %v", logging.QuoteText(plan.Name))

	// only the metadata plan will have the full set of entities that were archived so if we don't have one, nothing we can do
	if plan.Name != constants.Pliny {
		return plan.Name, true
	}

	planEntityPaths := plan.EntityPaths
	if archivedEntityPaths.Equal(plan.EntityPaths) {
		return plan.Name, true
	}

	archivedMissing := set.Diff(archivedEntityPaths, planEntityPaths)
	planMissing := set.Diff(planEntityPaths, archivedEntityPaths)

	msgPrefix := logging.ColorWarning.Sprintf("Unexpected deployment plan %v:", logging.QuoteText(plan.Name))
	for _, p := range slices.Sorted(archivedMissing.All()) {
		// Intentionally using Debug instead of Warn here because entities may have been filtered by server (e.g., app name)
		// so there isn't a reliable way to determine if the missing entity is valid or invalid.  Calling function
		// emits warning with message to use debug log level for details
		logger.Debugf("%v cannot deploy entity %v because plan did not contain it", msgPrefix, logging.ColorResource.QuoteText(p))
	}
	for _, p := range slices.Sorted(planMissing.All()) {
		// Skuid Review Required - Would this situation ever be expected to happen?
		//
		// TODO: Change to warn based on answer to above
		logger.Debugf("%v includes entity %v which was not requested ", msgPrefix, logging.ColorResource.QuoteText(p))
	}

	return plan.Name, false
}

func getDeployPlanRequestBody(headers RequestHeaders, plan []byte, filter *NlxPlanFilter, logger *logging.Logger) ([]byte, error) {
	logger = logger.WithName("getDeployPlanRequestBody", logging.Fields{
		"planLen": len(plan),
		"filter":  filter,
	})
	logger.Trace("Getting deployment plan request body")

	if filter == nil {
		return plan, nil
	}

	logger.Debugf("Including %v in get deployment plan(s) request", logging.ColorFilter.Text("plan filter"))
	// change content type to json and add content encoding
	headers[HeaderContentType] = JsonContentType
	headers[HeaderContentEncoding] = GzipContentEncoding
	// add the deployment plan bytes to the payload
	// instead of just using that as the payload
	requestBody := FilteredRequestBody{
		filter.AppName,
		// pages flag does not work as expected so commenting out
		// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
		//filter.PageNames,
		plan,
		filter.IgnoreSkuidDb,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		// intentionally not logging requestBody because the plan is a zip file
		return nil, fmt.Errorf("unable to convert deploy plan with filter to JSON bytes: %w", err)
	}

	return body, nil
}
