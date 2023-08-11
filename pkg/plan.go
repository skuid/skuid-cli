package pkg

import (
	"encoding/json"
	"fmt"
	"time"
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
	Host              string                `json:"host"`
	Port              string                `json:"port"`
	Endpoint          string                `json:"url"`
	Type              string                `json:"type"`
	Metadata          NlxMetadata           `json:"metadata"`
	Since             string                `json:"since"`
	Warnings          []string              `json:"warnings"`
	AllPermissionSets []PermissionSetResult `json:"allPermissionSets"`
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	tstr := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(tstr), nil
}

// NlxPlanFilter should be serialized and provided with the
// request for retrieval
type NlxPlanFilter struct {
	AppName       string    `json:"appName,omitempty"`
	PageNames     []string  `json:"pages,omitempty"`
	IgnoreSkuidDb bool      `json:"ignoreSkuidDb,omitempty"`
	Since         time.Time `json:"since,omitempty"`
}

// NewRetrievalRequestBody marshals the NlxMetadata into json and returns
// the body. This is the payload expected for the Retrieval Request
func NewRetrievalRequestBody(metadata NlxMetadata, since string) (body []byte) {
	// there should be no issue marshalling this thing.
	body, _ = json.Marshal(struct {
		Metadata NlxMetadata `json:"metadata"`
		Since    string      `json:"since"`
	}{
		Metadata: metadata,
		Since:    since,
	})
	return
}
