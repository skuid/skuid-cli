package pkg

import (
	"fmt"
	"iter"
	"net/http"
	"path"
	"time"

	"github.com/bobg/go-generics/v4/set"
	"github.com/bobg/go-generics/v4/slices"
	"github.com/bobg/seqs"
	"github.com/goccy/go-json"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

type PlanName string
type PlanMode int

const (
	PlanNamePliny  PlanName = constants.Pliny
	PlanNameWarden PlanName = constants.Warden

	PlanModeRetrieve PlanMode = iota + 1
	PlanModeDeploy
)

// Skuid Review Required - Is the NlxPlan struct accurate based on current APIs?
// Any residual portions of it remaining that should be removed based on removal of SFX support?
// Any items missing that should be here given current API functionality?
//
// NlxPlanPayload is the result of getting the plan from the pliny
// deployment retrieval plan endpoint
type nlxPlansPayload struct {
	// Cloud Data Service is WARDEN
	CloudDataService *nlxPlanPayload `json:"skuidCloudDataService"`
	// Metadata Service is PLINY
	MetadataService *nlxPlanPayload `json:"skuidMetadataService"`
}

type NlxPlans struct {
	Plans            []*NlxPlan
	PlanNames        []PlanName
	EntityPaths      set.Of[string]
	CloudDataService *NlxPlan
	MetadataService  *NlxPlan
}

// Skuid Review Required - Is the NlxPlan struct accurate based on current APIs?
// Any residual portions of it remaining that should be removed based on removal of SFX support?
// Any items missing that should be here given current API functionality?
type nlxPlanPayload struct {
	Host              string                `json:"host"`
	Port              string                `json:"port"`
	Endpoint          string                `json:"url"`
	Type              string                `json:"type"`
	Metadata          metadata.NlxMetadata  `json:"metadata"`
	Since             *time.Time            `json:"since"`
	AppSpecific       bool                  `json:"appSpecific"`
	Warnings          []string              `json:"warnings"`
	AllPermissionSets []PermissionSetResult `json:"allPermissionSets"`
}

type NlxPlan struct {
	nlxPlanPayload
	Name        PlanName
	EntityPaths set.Of[string]
	Mode        PlanMode
}

// Skuid Review Required - Is the NlxPlanFilter struct accurate based on current APIs?
// Any residual portions of it remaining that should be removed based on removal of SFX support?
// Any items missing that should be here given current API functionality?
//
// NlxPlanFilter should be serialized and provided with the
// request for retrieval
type NlxPlanFilter struct {
	AppName string `json:"appName,omitempty"`
	// pages flag does not work as expected so commenting out
	// TODO: Remove completely or fix issues depending on https://github.com/skuid/skuid-cli/issues/147 & https://github.com/skuid/skuid-cli/issues/148
	//PageNames     []string  `json:"pages,omitempty"`
	IgnoreSkuidDb bool       `json:"ignoreSkuidDb,omitempty"`
	Since         *time.Time `json:"since,omitempty"`
}

// NewRetrievalRequestBody marshals the NlxMetadata into json and returns
// the body. This is the payload expected for the Retrieval Request
func NewRetrievalRequestBody(md metadata.NlxMetadata, since *time.Time, appSpecific bool) ([]byte, error) {
	logger := logging.WithName("pkg.NewRetrievalRequestBody")

	content := struct {
		Metadata    metadata.NlxMetadata `json:"metadata"`
		Since       *time.Time           `json:"since,omitempty"`
		AppSpecific bool                 `json:"appSpecific"`
	}{
		Metadata:    md,
		Since:       since,
		AppSpecific: appSpecific,
	}

	if body, err := json.Marshal(content); err != nil {
		logger.WithError(err).Debugf("Error marshalling retrieval request body: %+v", content)
		return nil, fmt.Errorf("unable to convert retrieval request to JSON bytes: %w", err)
	} else {
		return body, nil
	}
}

func RequestNlxPlans(route string, headers RequestHeaders, body []byte, planMode PlanMode) (*NlxPlans, error) {
	if payload, err := JsonBodyRequest[nlxPlansPayload](
		route,
		http.MethodPost,
		body,
		headers,
	); err != nil {
		return nil, err
	} else {
		return newNlxPlans(payload, planMode), nil
	}
}

func newNlxPlans(payload *nlxPlansPayload, planMode PlanMode) *NlxPlans {
	plans := NlxPlans{EntityPaths: set.New[string]()}
	addPlan := func(plan *nlxPlanPayload, planName PlanName) *NlxPlan {
		planEntityPaths := getPlanEntityPaths(plan, planName, planMode)
		details := &NlxPlan{*plan, planName, planEntityPaths, planMode}
		plans.Plans = append(plans.Plans, details)
		plans.PlanNames = append(plans.PlanNames, details.Name)
		plans.EntityPaths.AddSeq(planEntityPaths.All())
		return details
	}

	// Skuid Review Required - Modifying to expect that a MetadataService plan will always exist.  Is this correct or is it possible metadata service
	// may not exist and if there is a cloud service we should process it?
	// See https://github.com/skuid/skuid-cli/issues/225, https://github.com/skuid/skuid-cli/issues/226 & https://github.com/skuid/skuid-cli/issues/229
	//
	// TODO: Adjust below with conditional if its possible that metadataservice may not exist and cloud data service, if present, should be processed
	// even though there isn't a metadata service.
	plans.MetadataService = addPlan(payload.MetadataService, PlanNamePliny)
	if payload.CloudDataService != nil {
		plans.CloudDataService = addPlan(payload.CloudDataService, PlanNameWarden)
	}
	return &plans
}

func getPlanEntityPaths(plan *nlxPlanPayload, planName PlanName, planMode PlanMode) set.Of[string] {
	planEntityPaths := set.New[string]()
	for _, mdt := range metadata.MetadataTypes.Members() {
		entities := plan.Metadata.GetFieldValue(mdt)
		var entityNames iter.Seq[string]
		if planMode == PlanModeRetrieve && planName == PlanNameWarden && mdt == metadata.MetadataTypeSitePermissionSets {
			// TODO: Eliminate or adjust/improve once above is answered and https://github.com/skuid/skuid-cli/issues/232 addressed
			entityNames = getWardenSitePermissionSetEntities(entities, planName)
		} else {
			entityNames = slices.Values(entities)
		}
		planEntityPaths.AddSeq(seqs.Map(entityNames, func(e string) string {
			return path.Join(mdt.DirName(), e)
		}))
	}

	return planEntityPaths
}

// Skuid Review Required - This is a workaround for issue https://github.com/skuid/skuid-cli/issues/232.  Please
// see the issue described there and advise - does the CLI need to change the NlxMetadata to be different between
// Warden & Pliny retrieve plans or is there a server bug?
// For now, leaving CLI as-is except for this scenario which obtains the "name" of the permission set
// for logging purposes only.
//
// TODO: Eliminate or adjust/improve once above is answered and https://github.com/skuid/skuid-cli/issues/232 addressed
func getWardenSitePermissionSetEntities(entities []string, planName PlanName) iter.Seq[string] {
	return seqs.Reduce(slices.Values(entities), set.New[string](), func(a set.Of[string], e string) set.Of[string] {
		var wsps wardenSitePermissionSet
		err := json.Unmarshal([]byte(e), &wsps)
		if err != nil {
			// (hopefully) should not happen in production
			// temporary workaround for https://github.com/skuid/skuid-cli/issues/232
			// TODO: eliminate or adjust/improve once https://github.com/skuid/skuid-cli/issues/232 is addressed
			panic(fmt.Errorf("unexpected site permission sets in plan %v", planName))
		}
		a.Add(wsps.Name)
		return a
	}).All()
}

type wardenSitePermissionSet struct {
	Name        string   `json:"name"`
	DataSources []string `json:"datasources"`
}
