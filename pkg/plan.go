package pkg

import (
	"encoding/json"
	"time"

	"github.com/skuid/skuid-cli/pkg/metadata"
)

// Skuid Review Required - Is the NlxPlan struct accurate based on current APIs?
// Any residual portions of it remaining that should be removed based on removal of SFX support?
// Any items missing that should be here given current API functionality?
//
// NlxPlanPayload is the result of getting the plan from the pliny
// deployment retrieval plan endpoint
type NlxPlanPayload struct {
	// Cloud Data Service is WARDEN
	CloudDataService *NlxPlan `json:"skuidCloudDataService"`
	// Metadata Service is PLINY
	MetadataService *NlxPlan `json:"skuidMetadataService"`
}

// Skuid Review Required - Is the NlxPlan struct accurate based on current APIs?
// Any residual portions of it remaining that should be removed based on removal of SFX support?
// Any items missing that should be here given current API functionality?
type NlxPlan struct {
	Host              string                `json:"host"`
	Port              string                `json:"port"`
	Endpoint          string                `json:"url"`
	Type              string                `json:"type"`
	Metadata          metadata.NlxMetadata  `json:"metadata"`
	Since             string                `json:"since"`
	AppSpecific       bool                  `json:"appSpecific"`
	Warnings          []string              `json:"warnings"`
	AllPermissionSets []PermissionSetResult `json:"allPermissionSets"`
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
	IgnoreSkuidDb bool      `json:"ignoreSkuidDb,omitempty"`
	Since         time.Time `json:"since,omitempty"`
}

// NewRetrievalRequestBody marshals the NlxMetadata into json and returns
// the body. This is the payload expected for the Retrieval Request
func NewRetrievalRequestBody(md metadata.NlxMetadata, since string, appSpecific bool) (body []byte) {
	// there should be no issue marshalling this thing.
	body, _ = json.Marshal(struct {
		Metadata    metadata.NlxMetadata `json:"metadata"`
		Since       string               `json:"since"`
		AppSpecific bool                 `json:"appSpecific"`
	}{
		Metadata:    md,
		Since:       since,
		AppSpecific: appSpecific,
	})
	return
}
