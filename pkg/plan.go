package pkg

import (
	"encoding/json"
)

// NlxPlanPayload is the result of getting the plan from the pliny
// deployment retrieval plan endpoint
type NlxPlanPayload struct {
	// Cloud Data Service is WARDEN
	CloudDataService *NlxPlan `json:"skuidCloudDataService"`
	// Metadata Service is PLINY
	MetadataService *NlxPlan `json:"skuidMetadataService"`
}

type NlxPlan struct {
	Host     string      `json:"host"`
	Port     string      `json:"port"`
	Endpoint string      `json:"url"`
	Type     string      `json:"type"`
	Metadata NlxMetadata `json:"metadata"`
	Warnings []string    `json:"warnings"`
}

// NlxPlanFilter should be serialized and provided with the
// request for retrieval
type NlxPlanFilter struct {
	AppName   string   `json:"appName,omitempty"`
	PageNames []string `json:"pages,omitempty"`
}

// NewRetrievalRequestBody marshals the NlxMetadata into json and returns
// the body. This is the payload expected for the Retrieval Request
func NewRetrievalRequestBody(metadata NlxMetadata) (body []byte) {
	// there should be no issue marshalling this thing.
	body, _ = json.Marshal(struct {
		Metadata NlxMetadata `json:"metadata"`
	}{
		Metadata: metadata,
	})
	return
}
