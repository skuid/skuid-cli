package nlx

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	ZIP_CONTENT_TYPE  = "application/zip"
	JSON_CONTENT_TYPE = "application/json"

	DEFAULT_CONTENT_TYPE = ZIP_CONTENT_TYPE
)

type NlxPlanResult map[string]NlxPlan

type NlxRetrieveFilter struct {
}

func GetRetrievePlan(host, authorizationToken string, filter *NlxRetrieveFilter) (duration time.Duration, result NlxPlanResult, err error) {

	planStart := time.Now()

	var body []byte = make([]byte, 0)
	if filter != nil {
		if body, err = json.Marshal(filter); err != nil {
			return
		}
	}

	if result, err = FastJsonBodyRequest[NlxPlanResult](
		fmt.Sprintf("%s/api/v2/metadata/retrieve/plan", host),
		fasthttp.MethodPost,
		body,
		map[string]string{
			fasthttp.HeaderContentType:   JSON_CONTENT_TYPE,
			fasthttp.HeaderAuthorization: fmt.Sprintf("Bearer %v", authorizationToken),
		},
	); err != nil {
		return
	}

	duration = time.Since(planStart)

	return
}

func GetDeployPlan(host, authorizationToken string) (duration time.Duration, result NlxPlanResult, err error) {

	planStart := time.Now()

	if result, err = FastJsonBodyRequest[NlxPlanResult](
		fmt.Sprintf("%s/api/v2/metadata/deploy/plan", host),
		fasthttp.MethodPost,
		[]byte{},
		map[string]string{
			fasthttp.HeaderContentType:   DEFAULT_CONTENT_TYPE,
			fasthttp.HeaderAuthorization: fmt.Sprintf("Bearer %v", authorizationToken),
		},
	); err != nil {
		return
	}

	duration = time.Since(planStart)

	return
}
