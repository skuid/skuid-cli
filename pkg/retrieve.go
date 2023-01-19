package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var (
	RetrievePlanRoute = fmt.Sprintf("api/%v/metadata/retrieve/plan", DEFAULT_API_VERSION)
)

func GetRetrievePlan(auth *Authorization, filter *NlxPlanFilter) (duration time.Duration, result NlxPlanPayload, err error) {

	planStart := time.Now()
	defer func() { duration = time.Since(planStart) }()

	var body = make([]byte, 0)
	if filter != nil {
		if body, err = json.Marshal(filter); err != nil {
			return
		}
	}

	// this is a pliny request, so we provide the access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)

	// no matter what we want to pass application/json
	// because the application/zip is discarded by pliny
	// and warden will throw an error
	headers[HeaderContentType] = JSON_CONTENT_TYPE

	result, err = JsonBodyRequest[NlxPlanPayload](
		fmt.Sprintf("%s/%s", auth.Host, RetrievePlanRoute),
		http.MethodPost,
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
	executePlan := func(name string, plan NlxPlan) error {

		logging.WithField("planName", name)
		logging.Get().Debugf("Beginning plan %v", color.Magenta.Sprint(name))

		headers := GeneratePlanHeaders(auth, plan)
		headers[HeaderContentType] = JSON_CONTENT_TYPE

		for k, header := range headers {
			logging.Get().Tracef("header: (%v => %v)", color.Yellow.Sprint(k), color.Green.Sprint(header))
		}

		url := GenerateRoute(auth, plan)

		logging.Get().Tracef("URL: %v", color.Blue.Sprint(url))

		result, err := Request(
			url, http.MethodPost, NewRetrievalRequestBody(plan.Metadata), headers,
		)

		if err != nil {
			logging.Get().WithFields(logrus.Fields{
				"plan":     plan,
				"planName": name,
				"url":      url,
			})
			logging.Get().WithError(err)
			logging.Get().Errorf("error with %v request", color.Magenta.Sprint(name))
			return err
		}

		results = append(results, NlxRetrievalResult{
			Plan:     plan,
			PlanName: name,
			Url:      url,
			Data:     result,
		})

		return nil
	}

	// has to be pliny, then warden
	if plans.MetadataService != nil {
		if err = executePlan(constants.PLINY, *plans.MetadataService); err != nil {
			return
		}
	}

	if plans.CloudDataService != nil {
		if err = executePlan(constants.WARDEN, *plans.CloudDataService); err != nil {
			return
		}
	}

	return
}
