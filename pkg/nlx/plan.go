package nlx

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

// This is the result of getting the plan from the pliny
// deployment retrieval plan endpoint
type NlxRetrievalPlans struct {
	// Cloud Data Service is WARDEN
	CloudDataService NlxPlan `json:"skuidCloudDataService"`
	// Metadata Service is PLINY
	MetadataService NlxPlan `json:"skuidMetadataService"`
}

// Serialize this and provide it with the
// request for retrieval
type NlxRetrieveFilter struct {
	AppName string `json:"appName"`
	// PageNames []string `json:"pageNames"`
}

func GetDeployPlan(auth *Authorization) (duration time.Duration, result NlxRetrievalPlans, err error) {
	planStart := time.Now()
	defer func() { duration = time.Since(planStart) }()

	if result, err = FastJsonBodyRequest[NlxRetrievalPlans](
		fmt.Sprintf("%s/api/%v/metadata/deploy/plan", auth.Host, DEFAULT_API_VERSION),
		fasthttp.MethodPost,
		[]byte{},
		RequestHeaders{
			fasthttp.HeaderContentType:   DEFAULT_CONTENT_TYPE,
			fasthttp.HeaderAuthorization: fmt.Sprintf("Bearer %v", auth.AuthorizationToken),
		},
	); err != nil {
		return
	}

	return
}

// NewRetrievalRequestBody marshals the NlxMetadata into json and returns
// the body. This is the payload expected for the Retrieval Request
func NewRetrievalRequestBody(metadata NlxMetadata, zip bool) (body []byte) {
	// there should be no issue marshalling this thing.
	body, _ = json.Marshal(struct {
		Metadata NlxMetadata `json:"metadata"`
		DoZip    bool        `json:"zip"`
	}{
		Metadata: metadata,
		DoZip:    zip,
	})
	return
}