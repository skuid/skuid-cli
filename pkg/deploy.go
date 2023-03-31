package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gookit/color"
	"github.com/skuid/skuid-cli/pkg/logging"
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
	AppName       string   `json:"appName"`
	PageNames     []string `json:"pageNames"`
	PlanBytes     []byte   `json:"plan"`
	IgnoreSkuidDb bool     `json:"ignoreSkuidDb"`
}

type PermissionSetResult struct {
	AppId          string `json:"app_id"`
	AppName        string `json:"app_name"`
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	OrganizationId string `json:"organization_id"`
}
type permissionSetResults struct {
	Inserts []PermissionSetResult `json:"inserts"`
	Updates []PermissionSetResult `json:"updates"`
	Deletes []PermissionSetResult `json:"deletes"`
}
type plinyResult struct {
	PermissionSets permissionSetResults `json:"permissionSets"`
}

func GetDeployPlan(auth *Authorization, deploymentPlan []byte, filter *NlxPlanFilter) (duration time.Duration, results NlxDynamicPlanMap, err error) {
	logging.Get().Trace("Getting Deploy Plan")
	start := time.Now()
	defer func() { logging.Get().Tracef("Prepare deployment took: %v", time.Since(start)) }()

	// pliny request, use access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)
	headers[HeaderContentType] = ZIP_CONTENT_TYPE

	var body []byte
	if filter != nil {
		logging.Get().Debug("Using file filter")
		// change content type to json
		headers[HeaderContentType] = JSON_CONTENT_TYPE
		// we instead add the deployment plan bytes to the payload
		// instead of just using that as the payload
		requestBody := FilteredRequestBody{
			filter.AppName,
			filter.PageNames,
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

func DeployModifiedFiles(auth *Authorization, targetDir, modifiedFile string) (err error) {
	planBody, err := ArchivePartial(targetDir, modifiedFile)
	if err != nil {
		return
	}

	logging.Get().Tracef("Getting Deployment Plan for Modified File (%v)", modifiedFile)

	_, plan, err := GetDeployPlan(auth, planBody, nil)
	if err != nil {
		return
	}

	logging.Get().Tracef("Received Deployment Plan for (%v), Deploying", modifiedFile)

	_, _, err = ExecuteDeployPlan(auth, plan, targetDir)
	if err != nil {
		return
	}

	logging.Get().Tracef("Successfully deployed metadata to Skuid Site: %v", modifiedFile)

	return
}

// ExecuteDeployPlan executes a map of plan items in a deployment plan
func ExecuteDeployPlan(auth *Authorization, plans NlxDynamicPlanMap, targetDir string) (duration time.Duration, planResults []NlxDeploymentResult, err error) {
	start := time.Now()
	defer func() { duration = time.Since(start) }()
	logging.Get().Trace("Executing Deploy Plan")

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
		payload, err := Archive(targetDir, &plan.Metadata)
		if err != nil {
			logging.Get().Trace("Error creating deployment ZIP archive")
			return
		}

		headers := GeneratePlanHeaders(auth, plan)
		logging.Get().Tracef("Plan Headers: %v\n", headers)

		url := GenerateRoute(auth, plan)
		logging.Get().Tracef("Plan Request: %v\n", url)

		if plan.Type == DATA_PLAN_TYPE && len(plan.NewPermissionSets) > 0 {
			// Create new permission sets with the UUIDs that the metadata service generated and assigned
			newPsPlan := plan // we do not need deep copy, as we are only changing the path
			newPsPlan.Endpoint = "/metadata/permissionsets"
			psurl := GenerateRoute(auth, newPsPlan)
			var pspayload []byte
			pspayload, err = json.Marshal(newPsPlan.NewPermissionSets)
			if err != nil {
				return
			}
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

		var response []byte
		if response, err = Request(
			url,
			http.MethodPost,
			payload,
			headers,
		); err == nil {
			planResults = append(planResults, NlxDeploymentResult{
				Plan: plan,
				Url:  url,
				Data: response,
			})
		} else {
			logging.Get().Tracef("Url: %v", url)
			logging.Get().Tracef("Error on request: %v\n", err.Error())
			return
		}
		logging.Get().Infof("Finished Deploying %v", color.Magenta.Sprint(plan.Type))

		if plan.Type == METADATA_PLAN_TYPE {
			// Collect new permission set UUIDs so we can create them in warden with matching UUIDs
			resultMap := plinyResult{}
			err = json.Unmarshal(response, &resultMap)
			if err != nil {
				return
			}
			dataPlan.NewPermissionSets = make([]PermissionSetResult, len(resultMap.PermissionSets.Inserts))
			for i, v := range resultMap.PermissionSets.Inserts {
				dataPlan.NewPermissionSets[i] = v
			}
		}
		return
	}

	// Run metadata plan first, because there may be new PermissionSets to sync up with the data plan
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
	Plan     NlxPlan
	PlanName string
	Url      string
	Data     []byte
}

func (result NlxDeploymentResult) String() string {
	return fmt.Sprintf("( Name: '%v', Url: %v => %v bytes )",
		result.PlanName,
		result.Url,
		len(result.Data),
	)
}
