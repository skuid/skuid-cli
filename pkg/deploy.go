package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
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
	AppId                 string      `json:"app_id"`
	AppName               string      `json:"app_name"`
	Id                    string      `json:"id"`
	Name                  string      `json:"name"`
	Description           string      `json:"description"`
	OrganizationId        string      `json:"organization_id"`
	DatasourcePermissions interface{} `json:"dataSourcePermissions"`
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
		// change content type to json and add content encoding
		headers[HeaderContentType] = JSON_CONTENT_TYPE
		headers[HeaderContentEncoding] = GZIP_CONTENT_ENCODING
		// add the deployment plan bytes to the payload
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

func DeployModifiedFiles(auth *Authorization, targetDir string, modifiedFile string) (err error) {
	if !filepath.IsAbs(targetDir) {
		err = fmt.Errorf("targetDir must be an absolute path")
		return
	}
	if !filepath.IsAbs(modifiedFile) {
		err = fmt.Errorf("modifiedFile must be an absolute path")
		return
	}

	var relativeFilePath string
	if relativeFilePath, err = filepath.Rel(targetDir, modifiedFile); err != nil {
		return
	}
	metadataType, _ := GetEntityDetails(relativeFilePath)
	if _, mdtok := GetMetadataTypeNameByDirName(metadataType); !mdtok {
		logging.Get().Warnf("Unexpected metadata type [%v] detected, skipping deployment of modified file: %v", metadataType, modifiedFile)
		return
	}

	var filesToDeploy []string
	if metadataType == "site" {
		if strings.HasSuffix(relativeFilePath, ".skuid.json") {
			filesToDeploy = append(filesToDeploy, relativeFilePath, strings.TrimSuffix(relativeFilePath, ".skuid.json"))
		} else if !strings.HasSuffix(relativeFilePath, ".json") {
			filesToDeploy = append(filesToDeploy, relativeFilePath, relativeFilePath+".skuid.json")
		} else {
			filesToDeploy = append(filesToDeploy, relativeFilePath)
		}
	} else if metadataType == "pages" {
		var pagePathNoExt string
		if strings.HasSuffix(relativeFilePath, ".xml") {
			pagePathNoExt = strings.TrimSuffix(relativeFilePath, ".xml")
		} else if strings.HasSuffix(relativeFilePath, ".json") {
			pagePathNoExt = strings.TrimSuffix(relativeFilePath, ".json")
		} else {
			logging.Get().Warnf("Unexpected [%v] file detected, skipping deployment of modified file: %v", metadataType, modifiedFile)
			return
		}
		filesToDeploy = append(filesToDeploy, pagePathNoExt+".xml", pagePathNoExt+".json")
	} else if metadataType == "files" {
		if strings.HasSuffix(relativeFilePath, ".skuid.json") {
			filesToDeploy = append(filesToDeploy, relativeFilePath, strings.TrimSuffix(relativeFilePath, ".skuid.json"))
		} else {
			filesToDeploy = append(filesToDeploy, relativeFilePath, relativeFilePath+".skuid.json")
		}
	} else {
		if !strings.HasSuffix(relativeFilePath, ".json") {
			logging.Get().Warnf("Unexpected [%v] file detected, skipping deployment of modified file: %v", metadataType, modifiedFile)
			return
		}
		filesToDeploy = append(filesToDeploy, relativeFilePath)
	}
	planBody, fileCount, err := ArchiveFiles(targetDir, filesToDeploy)
	if err != nil {
		return
	}
	if fileCount != len(filesToDeploy) {
		logging.Get().Warnf("Unable to locate all expected files, found %d, expected %d, skipping deployment of modified file: %v (expected files: %v)", fileCount, len(filesToDeploy), modifiedFile, filesToDeploy)
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
// A summary of its steps:
// 1. Deploy metadata plan aka pliny resources
// 2. After pliny is deployed, take the results from that deploy and pull out the app permission set ids
// (and their datasource permissions) that were inserted / updated
// 3. Deploy data plan aka warden
// 4. After its deployed take the app permission set ids from the pliny deploy and deploy those permission sets to warden
// 5. If metadata and data was deployed, send a request to pliny to sync its datasources' external_ids with warden datasource
// ids, in case they changed during the deploy
func ExecuteDeployPlan(auth *Authorization, plans NlxDynamicPlanMap, targetDir string) (duration time.Duration, planResults []NlxDeploymentResult, err error) {
	if !filepath.IsAbs(targetDir) {
		err = fmt.Errorf("targetDir must be an absolute path")
		return
	}

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
			// Collect all app permission set UUIDs and datasource permissions from pliny so we can use them to create
			// datasource permissions in warden. This is necessary because we need the UUIDs and the deploy plan only has the names
			resultMap := plinyResult{}
			err = json.Unmarshal(response, &resultMap)
			if err != nil {
				return
			}
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
