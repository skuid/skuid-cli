package pkg

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bobg/go-generics/v4/set"
	"github.com/bobg/go-generics/v4/slices"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	DeployPlanRoute = fmt.Sprintf("api/%v/metadata/deploy/plan", DEFAULT_API_VERSION)
)

const (
	METADATA_PLAN_KEY  = "skuidMetadataService"
	METADATA_PLAN_TYPE = "metadataService"
	DATA_PLAN_KEY      = "skuidCloudDataService"
	DATA_PLAN_TYPE     = "dataService"
)

type NlxDynamicPlanMap map[string]NlxPlan

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

type plinyResult struct {
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

// TODO: This can be made private once improvements are made to address https://github.com/skuid/skuid-cli/issues/166 (e.g., test coverage, dependency injection to enable proper unit tests, etc.)
func GetDeployPlan(auth *Authorization, deploymentPlan []byte, filter *NlxPlanFilter) (duration time.Duration, results NlxDynamicPlanMap, err error) {
	logging.Get().Trace("Getting Deploy Plan")
	start := time.Now()
	defer func() { duration = time.Since(start) }()

	// pliny request, use access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)
	headers[HeaderContentType] = ZIP_CONTENT_TYPE

	var body []byte
	if filter != nil {
		logging.Get().Debug("Using file filter")
		// change content type to json and add content encoding
		headers[HeaderContentType] = JSON_CONTENT_TYPE
		headers[HeaderContentEncoding] = GZIP_CONTENT_ENCODING
		// add the deployment plan bytes to the payload
		// instead of just using that as the payload
		requestBody := FilteredRequestBody{
			filter.AppName,
			// pages flag does not work as expected so commenting out
			// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
			//filter.PageNames,
			deploymentPlan,
			filter.IgnoreSkuidDb,
		}

		if body, err = json.Marshal(requestBody); err != nil {
			logging.Get().Warnf("Error marshalling filter request: %v", err)
			return
		}
	} else {
		// set the deployment plan as the payload
		body = deploymentPlan
	}

	// make the request
	results, err = JsonBodyRequest[NlxDynamicPlanMap](
		fmt.Sprintf("%s/%s", auth.Host, DeployPlanRoute),
		http.MethodPost,
		body,
		headers,
	)

	return
}

func Deploy(auth *Authorization, targetDirectory string, archiveFilter ArchiveFilter, entitiesToArchive []metadata.MetadataEntity, planFilter *NlxPlanFilter) error {
	if !filepath.IsAbs(targetDirectory) {
		return fmt.Errorf("targetDirectory must be an absolute path")
	}

	logging.Get().Info("Building deployment request")
	deploymentRequest, archivedFilePaths, archivedEntities, err := archive(targetDirectory, archiveFilter, getMetadataEntityPaths(entitiesToArchive))
	if err != nil {
		return err
	}
	logging.WithFields(logrus.Fields{
		"deploymentBytes":     len(deploymentRequest),
		"deploymentEntities":  len(archivedEntities),
		"deploymentFilePaths": len(archivedFilePaths),
	}).Tracef("Built deployment request")

	logging.Get().Info("Getting deployment plan(s)")
	duration, plans, err := GetDeployPlan(auth, deploymentRequest, planFilter)
	if err != nil {
		return err
	} else if err := validateDeployPlans(plans, archivedEntities); err != nil {
		// intentionally ignoring any error as currently using validation for logging purposes only
		logging.Get().Warn(err)
	}
	logging.WithFields(logrus.Fields{
		"plans":    len(plans),
		"duration": duration,
	}).Tracef("Received deployment plan(s)")

	logging.Get().Info("Executing deployment plan(s)")
	duration, results, err := ExecuteDeployPlan(auth, plans, targetDirectory)
	if err != nil {
		return err
	} else if logging.DebugEnabled() {
		// deployment results are unreliable and will lead to false positives so only inspect them when debug is enabled for perf reasons
		// intentionally ignoring any error and using Trace to log since its only for informational purposes at this point
		// TODO: when Issue #211 is addressed server side, the error could be acted upon as appropriate
		// see https://github.com/skuid/skuid-cli/issues/211
		if vErr := validateDeployResults(results); vErr != nil {
			logging.Get().Tracef("deployment results do not match expected results, however the deployment results are NOT RELIABLE so there may be false positives: %v", vErr)
		}
	}
	logging.WithFields(logrus.Fields{
		"results":  len(results),
		"duration": duration,
	}).Tracef("Executed deployment plan(s)")

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
func ExecuteDeployPlan(auth *Authorization, plans NlxDynamicPlanMap, targetDir string) (duration time.Duration, planResults []NlxDeploymentResult, err error) {
	logging.Get().Trace("Executing Deploy Plan")

	start := time.Now()
	defer func() { duration = time.Since(start) }()

	if !filepath.IsAbs(targetDir) {
		err = fmt.Errorf("targetDir must be an absolute path")
		return
	}

	metaPlan, mok := plans[METADATA_PLAN_KEY]
	dataPlan, dok := plans[DATA_PLAN_KEY]
	if !mok && !dok {
		return
	}

	// Warning for metaplan
	for _, warning := range metaPlan.Warnings {
		logging.Get().Warnf("Warning %v", warning)
	}

	// Warning for dataplan
	for _, warning := range dataPlan.Warnings {
		logging.Get().Warnf("Warning %v", warning)
	}

	planResults = make([]NlxDeploymentResult, 0)

	executePlan := func(plan NlxPlan) (err error) {
		logging.Get().Infof("Deploying %v", color.Magenta.Sprint(plan.Type))

		logging.Get().Tracef("Archiving %v", targetDir)
		payload, archivedFilePaths, archivedEntities, err := archive(targetDir, MetadataArchiveFilter(&plan.Metadata), getPlanEntityPaths(plan))
		if err != nil {
			logging.Get().Trace("Error creating deployment ZIP archive")
			return
		}
		logging.WithFields(logrus.Fields{
			"payloadBytes":     len(payload),
			"payloadEntities":  len(archivedEntities),
			"payloadFilePaths": len(archivedFilePaths),
		}).Tracef("Built payload for deploy plan %q", plan.Type)

		headers := GeneratePlanHeaders(auth, plan)
		logging.Get().Tracef("Plan Headers: %v\n", headers)

		url := GenerateRoute(auth, plan)
		logging.Get().Tracef("Plan Request: %v\n", url)

		var response []byte
		// Skuid Review Required - Results are not complete for metadataService and completely blank for dataService.  Need full results in
		// order to validate against plan.
		// see https://github.com/skuid/skuid-cli/issues/211
		// TODO: dataService is skipped during validation because its blank, update validationDeployResults once above is answered/addressed.
		response, err = Request(url, http.MethodPost, payload, headers)
		if err != nil {
			logging.Get().Tracef("Url: %v", url)
			logging.Get().Tracef("Error on request: %v\n", err.Error())
			return
		}
		logging.Get().Infof("Finished Deploying %v", color.Magenta.Sprint(plan.Type))

		var resultMap plinyResult
		err = json.Unmarshal(response, &resultMap)
		if err != nil {
			return
		}
		planResults = append(planResults, NlxDeploymentResult{
			Plan:           plan,
			ResultName:     plan.Type,
			Url:            url,
			ResponseLength: len(response),
			Data:           resultMap,
		})

		if plan.Type == METADATA_PLAN_TYPE {
			// Collect all app permission set UUIDs and datasource permissions from pliny so we can use them to create
			// datasource permissions in warden. This is necessary because we need the UUIDs and the deploy plan only has the names
			insertLen := len(resultMap.PermissionSets.Inserts)
			updateLen := len(resultMap.PermissionSets.Updates)
			dataPlan.AllPermissionSets = make([]PermissionSetResult, insertLen+updateLen)
			i := 0
			for _, v := range resultMap.PermissionSets.Inserts {
				dataPlan.AllPermissionSets[i] = v
				i++
			}
			for _, v := range resultMap.PermissionSets.Updates {
				dataPlan.AllPermissionSets[i] = v
				i++
			}
		}

		if plan.Type == DATA_PLAN_TYPE && len(plan.AllPermissionSets) > 0 {
			// Create permission set datasource permissions with the UUIDs that the metadata service generated and assigned
			newPsPlan := plan // we do not need a deep copy, as we are only changing the path
			newPsPlan.Endpoint = "/metadata/update-permissionsets"
			psurl := GenerateRoute(auth, newPsPlan)
			var pspayload []byte
			pay := newPsPlan.AllPermissionSets
			pspayload, err = json.Marshal(pay)
			if err != nil {
				return
			}
			// Skuid Review Required - How can the response be validated here as an empty result is returned?
			// see https://github.com/skuid/skuid-cli/issues/211
			// TODO: Add result to planResults with a PlanName of "PermissionSets" to differentiate it from "dataService" since that is the "Plan.Type" for both and update validateDeployResults
			// accordingly once above is answered
			_, err = Request(
				psurl,
				http.MethodPost,
				pspayload,
				headers,
			)
			if err != nil {
				return
			}
		}
		return
	}

	// Run metadata plan first, because there may be PermissionSet data to create
	if mok {
		err = executePlan(metaPlan)
		if err != nil {
			return
		}
	}
	if dok {
		err = executePlan(dataPlan)
		if err != nil {
			return
		}
	}

	// Tell pliny to sync datasource external_id field with warden ids
	if mok && dok {
		syncPlan := metaPlan
		headers := GeneratePlanHeaders(auth, syncPlan)
		syncPlan.Endpoint = "/metadata/deploy/sync"
		url := GenerateRoute(auth, syncPlan)
		// Skuid Review Required - How can the response be validated here?  It appears that an array is returned and the array includes objects with properties similar to the results returned
		// from metaPlan & dataPlan deploys but they aren't shaped within metadata types.  Is Sync only for datasources?  Can it affect other metadata types?  What is the response structure
		// and how can it be validated against syncPlan?
		// see https://github.com/skuid/skuid-cli/issues/211
		// TODO: Add result to planResults with a PlanName of "Sync" to differentiate it from "metadataService" since that is the "Plan.Type" for both and update validateDeployResults
		// accordingly once above is answered
		_, err = Request(
			url,
			http.MethodPost,
			[]byte{},
			headers,
		)
	}

	return
}

type NlxDeploymentResult struct {
	Plan           NlxPlan
	ResultName     string
	Url            string
	ResponseLength int
	Data           plinyResult
}

func (result NlxDeploymentResult) String() string {
	return fmt.Sprintf("( Name: '%v', Url: %v => %v bytes )",
		result.ResultName,
		result.Url,
		result.ResponseLength,
	)
}

func archive(targetDirectory string, archiveFilter ArchiveFilter, entitiesToArchive set.Of[string]) ([]byte, []string, []metadata.MetadataEntity, error) {
	if deploymentRequest, archivedFilePaths, archivedEntities, err := Archive(os.DirFS(targetDirectory), util.NewFileUtil(), archiveFilter); err != nil {
		return deploymentRequest, archivedFilePaths, archivedEntities, err
	} else if err = validateArchive(entitiesToArchive, archivedEntities); err != nil {
		return deploymentRequest, archivedFilePaths, archivedEntities, err
	} else {
		return deploymentRequest, archivedFilePaths, archivedEntities, nil
	}
}

func getPlanEntityPaths(plan NlxPlan) set.Of[string] {
	planEntityPaths := set.New[string]()
	for _, mdt := range metadata.MetadataTypes.Members() {
		entities := plan.Metadata.GetFieldValue(mdt)
		planEntityPaths.Add(slices.Map(entities, func(e string) string {
			return path.Join(mdt.DirName(), e)
		})...)
	}

	return planEntityPaths
}

func getMetadataEntityPaths(entities []metadata.MetadataEntity) set.Of[string] {
	return set.New(slices.Map(entities, func(me metadata.MetadataEntity) string {
		return me.Path
	})...)
}

func validateArchive(expectedEntityPaths set.Of[string], actualEntities []metadata.MetadataEntity) error {
	if len(expectedEntityPaths) == 0 {
		return nil
	}

	actualEntityPaths := getMetadataEntityPaths(actualEntities)

	if expectedEntityPaths.Equal(actualEntityPaths) {
		return nil
	}

	expectedMissing := set.Diff(expectedEntityPaths, actualEntityPaths)
	actualMissing := set.Diff(actualEntityPaths, expectedEntityPaths)

	var errors []string
	if len(expectedMissing) > 0 {
		errors = append(errors, fmt.Sprintf("requested entities %q were not found", expectedMissing.Slice()))
	}

	if len(actualMissing) > 0 {
		errors = append(errors, fmt.Sprintf("found entities %q which were not requested", actualMissing.Slice()))
	}

	errMsg := strings.Join(errors, " **AND** ")
	if len(errors) == 0 {
		// should never happen since len(errors) should always be > 0 but just in case
		errMsg = fmt.Sprintf("requested entities %q does not match found entities %q", expectedEntityPaths.Slice(), actualEntityPaths.Slice())
	}
	return fmt.Errorf("unable to prepare archive: %v", errMsg)
}

func validateDeployResults(results []NlxDeploymentResult) error {
	var errors []string
	for _, result := range results {
		// Skuid Review Required - dataService and permission-set APIs return empty responses and sync API returns unknown format so its not known how to
		// reliably parse.  Need to have complete results from all APIs to be able to reliably validate results.
		// see https://github.com/skuid/skuid-cli/issues/211
		// TODO: All 4 APIs should be validated once the results are reliable and in a known format that can be parsed
		if result.Plan.Type != METADATA_PLAN_TYPE {
			continue
		}
		errMsg, ok := validateDeployResult(result)
		if !ok {
			errors = append(errors, errMsg)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%v", strings.Join(errors, "; "))
	}

	return nil
}

func validateDeployResult(result NlxDeploymentResult) (string, bool) {
	planEntityPaths := getPlanEntityPaths(result.Plan)
	modifiedEntityPaths, deletedEntityPaths := getResultEntityPaths(result)

	modifiedEqual := planEntityPaths.Equal(modifiedEntityPaths)
	hasDeleted := len(deletedEntityPaths) > 0

	if modifiedEqual && !hasDeleted {
		return "", true
	}

	var errors []string
	if !modifiedEqual {
		planMissing := set.Diff(planEntityPaths, modifiedEntityPaths)
		modifiedMissing := set.Diff(modifiedEntityPaths, planEntityPaths)

		if len(planMissing) > 0 {
			errors = append(errors, fmt.Sprintf("result indicated that expected entities %q were not deployed", planMissing.Slice()))
		}

		if len(modifiedMissing) > 0 {
			errors = append(errors, fmt.Sprintf("result indicated that entities %q were deployed but they were not expected", modifiedMissing.Slice()))
		}
	}

	if hasDeleted {
		errors = append(errors, fmt.Sprintf("result indicated that entites %q were deleted which is not expected", deletedEntityPaths.Slice()))
	}

	errMsg := strings.Join(errors, " **AND** ")
	if len(errors) == 0 {
		// should never happen since len(errors) should always be > 0 but just in case
		// Note - hasDeleted must be false here so do not need to include it in message
		errMsg = fmt.Sprintf("expected result to contain entities %q but it contained %q", planEntityPaths.Slice(), modifiedEntityPaths.Slice())
	}
	return fmt.Sprintf("unexpected deployment result for %v: %v", result.ResultName, errMsg), false
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

func getResultEntityPaths(result NlxDeploymentResult) (set.Of[string], set.Of[string]) {
	modifiedEntityPaths := set.New[string]()
	deletedEntityPaths := set.New[string]()

	for _, mdt := range metadata.MetadataTypes.Members() {
		typeResults := getTypeResults(mdt, result)
		modifiedEntityPaths.Add(slices.Map(typeResults.Inserts, func(r metadataTypeResult) string {
			return path.Join(mdt.DirName(), resolveEntityName(r))
		})...)
		modifiedEntityPaths.Add(slices.Map(typeResults.Updates, func(r metadataTypeResult) string {
			return path.Join(mdt.DirName(), resolveEntityName(r))
		})...)
		deletedEntityPaths.Add(slices.Map(typeResults.Deletes, func(r metadataTypeResult) string {
			return path.Join(mdt.DirName(), resolveEntityName(r))
		})...)
	}

	return modifiedEntityPaths, deletedEntityPaths
}

func getTypeResults(mdt metadata.MetadataType, result NlxDeploymentResult) metadataTypeResults {
	switch mdt {
	case metadata.MetadataTypeApps:
		return result.Data.Apps
	case metadata.MetadataTypeAuthProviders:
		return result.Data.AuthProviders
	case metadata.MetadataTypeComponentPacks:
		return result.Data.ComponentPacks
	case metadata.MetadataTypeDataServices:
		return result.Data.DataServices
	case metadata.MetadataTypeDataSources:
		return result.Data.DataSources
	case metadata.MetadataTypeDesignSystems:
		return result.Data.DesignSystems
	case metadata.MetadataTypeVariables:
		return result.Data.Variables
	case metadata.MetadataTypeFiles:
		return result.Data.Files
	case metadata.MetadataTypePages:
		return result.Data.Pages
	case metadata.MetadataTypePermissionSets:
		inserted := slices.Map(result.Data.PermissionSets.Inserts, func(r PermissionSetResult) metadataTypeResult {
			return r.metadataTypeResult
		})
		updated := slices.Map(result.Data.PermissionSets.Updates, func(r PermissionSetResult) metadataTypeResult {
			return r.metadataTypeResult
		})
		deleted := slices.Map(result.Data.PermissionSets.Deletes, func(r PermissionSetResult) metadataTypeResult {
			return r.metadataTypeResult
		})
		return metadataTypeResults{inserted, updated, deleted}
	case metadata.MetadataTypeSitePermissionSets:
		return result.Data.SitePermissionSets
	case metadata.MetadataTypeSessionVariables:
		return result.Data.SessionVariables
	case metadata.MetadataTypeSite:
		return result.Data.Site
	case metadata.MetadataTypeThemes:
		return result.Data.Themes
	default:
		// should not happen in production
		panic(fmt.Errorf("unexpected metadata type encountered while obtaining deployment result"))
	}
}

func validateDeployPlans(plans NlxDynamicPlanMap, archivedEntities []metadata.MetadataEntity) error {
	metaPlan, ok := plans[METADATA_PLAN_KEY]
	if !ok {
		// only the metadata plan will have the full set of entities that were archived so if we don't have one, nothing we can do
		return nil
	}

	archivedEntityPaths := getMetadataEntityPaths(archivedEntities)
	planEntityPaths := getPlanEntityPaths(metaPlan)

	if archivedEntityPaths.Equal(planEntityPaths) {
		return nil
	}

	archivedMissing := set.Diff(archivedEntityPaths, planEntityPaths)
	planMissing := set.Diff(planEntityPaths, archivedEntityPaths)

	var errors []string
	if len(archivedMissing) > 0 {
		errors = append(errors, fmt.Sprintf("cannot deploy entities %q because plan did not contain them", archivedMissing.Slice()))
	}

	if len(planMissing) > 0 {
		errors = append(errors, fmt.Sprintf("plan includes entities %q which were not expected", planMissing.Slice()))
	}

	errMsg := strings.Join(errors, " **AND** ")
	if len(errors) == 0 {
		// should never happen since len(errors) should always be > 0 but just in case
		errMsg = fmt.Sprintf("expected plan to contain entities %q but it contained %q", archivedEntityPaths.Slice(), planEntityPaths.Slice())
	}
	return fmt.Errorf("unexpected deployment plan: %v", errMsg)
}
