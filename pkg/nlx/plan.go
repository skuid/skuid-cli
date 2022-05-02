package nlx

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"

	"github.com/skuid/tides/pkg/logging"
)

const (
	ZIP_CONTENT_TYPE  = "application/zip"
	JSON_CONTENT_TYPE = "application/json"

	DEFAULT_CONTENT_TYPE = ZIP_CONTENT_TYPE

	HEADER_SKUID_PUBLIC_KEY_ENDPOINT = "x-skuid-public-key-endpoint"
)

// This is the result of getting the plan from the pliny
// deployment retrieval plan endpoint
type NlxPlanResult struct {
	// Cloud Data Service is WARDEN
	CloudDataService NlxPlan `json:"skuidCloudDataService"`
	// Metadata Service is PLINY
	MetadataService NlxPlan `json:"skuidMetadataService"`
}

type NlxRetrieveFilter struct {
}

func AuthorizationHeaderMap(host, token string) map[string]string {
	return map[string]string{
		fasthttp.HeaderAuthorization:     fmt.Sprintf("Bearer %v", token),
		HEADER_SKUID_PUBLIC_KEY_ENDPOINT: fmt.Sprintf("%v/api/v1/site/verificationkey", host),
	}
}

func GetRetrievePlan(auth *Authorization, filter *NlxRetrieveFilter) (duration time.Duration, result NlxPlanResult, err error) {

	planStart := time.Now()

	var body []byte = make([]byte, 0)
	if filter != nil {
		if body, err = json.Marshal(filter); err != nil {
			return
		}
	}

	headers := AuthorizationHeaderMap(auth.Host, auth.AccessToken)
	headers[fasthttp.HeaderContentType] = JSON_CONTENT_TYPE

	if result, err = FastJsonBodyRequest[NlxPlanResult](
		fmt.Sprintf("%s/api/v2/metadata/retrieve/plan", auth.Host),
		fasthttp.MethodPost,
		body,
		headers,
	); err != nil {
		return
	}

	duration = time.Since(planStart)

	return
}

func GetDeployPlan(auth *Authorization) (duration time.Duration, result NlxPlanResult, err error) {
	planStart := time.Now()
	defer func() { duration = time.Since(planStart) }()

	if result, err = FastJsonBodyRequest[NlxPlanResult](
		fmt.Sprintf("%s/api/v2/metadata/deploy/plan", auth.Host),
		fasthttp.MethodPost,
		[]byte{},
		map[string]string{
			fasthttp.HeaderContentType:   DEFAULT_CONTENT_TYPE,
			fasthttp.HeaderAuthorization: fmt.Sprintf("Bearer %v", auth.AuthorizationToken),
		},
	); err != nil {
		return
	}

	return
}

type NlxRetrievalRequest struct {
	Metadata NlxMetadata `json:"metadata"`
	DoZip    bool        `json:"zip"`
}

func NewRetrievalRequestBody(metadata NlxMetadata, zip bool) (body []byte) {
	// there should be no issue marshalling this thing.
	body, _ = json.Marshal(NlxRetrievalRequest{
		Metadata: metadata,
		DoZip:    zip,
	})
	return
}

type NlxRetrievalResult struct {
	Plan     NlxPlan
	PlanName string
	Url      string
	Data     []byte
}

func (x NlxRetrievalResult) String() string {
	return fmt.Sprintf("(%v => %v (size: %v))", x.PlanName, x.Url, len(x.Data))
}

func GetRequestInformation(info *Authorization, plan NlxPlan) (headers map[string]string, url string) {
	// warden requests all have a different host than the one we originall authenticated
	// with.
	wardenRequest := plan.Host != ""

	// for legibility
	plinyRequest := !wardenRequest

	// when given a warden request we need to provide the authorization / jwt token
	if wardenRequest {
		headers = AuthorizationHeaderMap(plan.Host, info.AuthorizationToken)
		if plan.Port != "" {
			url = fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.Endpoint)
		} else {
			url = fmt.Sprintf("%s/api/v2%s", plan.Host, plan.Endpoint)
		}
	}

	// with a pliny request we just attach the access token
	if plinyRequest {
		headers = AuthorizationHeaderMap(plan.Host, info.AccessToken)
		url = fmt.Sprintf("%s/api/v2%s", info.Host, plan.Endpoint)
	}

	return
}

func ExecuteRetrieval(plans NlxPlanResult, info *Authorization, zip bool) (duration time.Duration, err error) {
	// for timing sake
	start := time.Now()
	defer func() { duration = time.Since(start) }()

	eg := &errgroup.Group{}
	results := make(chan NlxRetrievalResult)

	executePlan := func(name string, plan NlxPlan) func() error {
		return func() error {
			logging.VerboseF("Firing off %v", name)

			headers, url := GetRequestInformation(info, plan)

			logging.VerboseF("Plan Request: %v\n", url)

			if zip {
				headers[fasthttp.HeaderContentType] = ZIP_CONTENT_TYPE
			} else {
				headers[fasthttp.HeaderContentType] = JSON_CONTENT_TYPE
			}

			if result, err := FastRequest(
				url, fasthttp.MethodPost, NewRetrievalRequestBody(plan.Metadata, zip), headers,
			); err == nil {
				results <- NlxRetrievalResult{
					Plan:     plan,
					PlanName: name,
					Url:      url,
					Data:     result,
				}
			} else {
				logging.VerboseF("Plan: %v", plan)
				logging.VerboseF("PlanName: %v", name)
				logging.VerboseF("Url: %v", url)
				logging.VerboseF("Error on request: %v\n", err.Error())
				return err
			}

			return nil
		}
	}

	eg.Go(executePlan("Warden", plans.CloudDataService))
	eg.Go(executePlan("Pliny", plans.MetadataService))

	go func() {
		err := eg.Wait()
		close(results)
		if err != nil {
			logging.PrintError("Error when executing retrieval plan:", err)
		}
	}()

	for result := range results {
		logging.VerboseF("%v\n", result)
	}

	return
}
