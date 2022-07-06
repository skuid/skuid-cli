package pkg

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg/logging"
)

var (
	RetrievePlanRoute = fmt.Sprintf("api/%v/metadata/retrieve/plan", DEFAULT_API_VERSION)
)

func GetRetrievePlan(auth *Authorization, filter *NlxPlanFilter) (duration time.Duration, result NlxPlanPayload, err error) {

	planStart := time.Now()
	defer func() { duration = time.Since(planStart) }()

	var body []byte = make([]byte, 0)
	if filter != nil {
		if body, err = json.Marshal(filter); err != nil {
			return
		}
	}

	// this is a pliny request, so we provide the access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)
	headers[fasthttp.HeaderContentType] = JSON_CONTENT_TYPE
	result, err = FastJsonBodyRequest[NlxPlanPayload](
		fmt.Sprintf("%s/%s", auth.Host, RetrievePlanRoute),
		fasthttp.MethodPost,
		body,
		headers,
	)

	return
}

type NlxRetrievalResult struct {
	Plan     NlxPlan
	PlanName string
	Url      string
	Data     []byte
}

func ExecuteRetrieval(auth *Authorization, plans NlxPlanPayload) (duration time.Duration, results []NlxRetrievalResult, err error) {
	logging.WithFields(logrus.Fields{
		"func": "ExecuteRetrieval",
	})
	// for timing sake
	start := time.Now()
	defer func() { duration = time.Since(start) }()

	// this function generically handles a plan based on name / stuff
	executePlan := func(name string, plan NlxPlan) (retrievalResult NlxRetrievalResult, err error) {
		log := logging.WithField("planName", name)
		log.Debugf("Firing off %v", color.Magenta.Sprint(name))

		headers := GeneratePlanHeaders(auth, plan)
		headers[fasthttp.HeaderContentType] = JSON_CONTENT_TYPE

		for k, header := range headers {
			log.Tracef("header: (%v => %v)", color.Yellow.Sprint(k), color.Green.Sprint(header))
		}

		url := GenerateRoute(auth, plan)

		log.Tracef("URL: %v", color.Blue.Sprint(url))

		result, err := FastRequest(
			url, fasthttp.MethodPost, NewRetrievalRequestBody(plan.Metadata), headers,
		)

		if err != nil {
			log = log.WithFields(logrus.Fields{
				"plan":     plan,
				"planName": name,
				"url":      url,
			})
			log = log.WithError(err)
			log.Errorf("error with %v request", color.Magenta.Sprint(name))
			return
		}

		retrievalResult = NlxRetrievalResult{
			Plan:     plan,
			PlanName: name,
			Url:      url,
			Data:     result,
		}

		return
	}

	// fire off the threads
	var warden, pliny NlxRetrievalResult
	warden, err = executePlan("Warden", plans.CloudDataService)
	if err != nil {
		return
	}
	pliny, err = executePlan("Pliny", plans.MetadataService)
	if err != nil {
		return
	}

	// consume the closed channel (probably return an array; todo)
	results = []NlxRetrievalResult{warden, pliny}

	return
}

// NlxRetrievalResult is only used

func (x NlxRetrievalResult) String() string {
	return fmt.Sprintf("(%v => %v (size: %v))", x.PlanName, x.Url, len(x.Data))
}
