package pkg

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"

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

	// we're going to create different threads
	// for both of the plans
	eg := &errgroup.Group{}
	ch := make(chan NlxRetrievalResult)

	// this function generically handles a plan based on name / stuff
	executePlan := func(name string, plan NlxPlan) func() error {
		return func() error {
			log := logging.WithField("planName", name)
			log.Debugf("Firing off %v", color.Magenta.Sprint(name))

			headers := GeneratePlanHeaders(auth, plan)
			headers[fasthttp.HeaderContentType] = JSON_CONTENT_TYPE

			for k, header := range headers {
				log.Tracef("header: (%v => %v)", color.Yellow.Sprint(k), color.Green.Sprint(header))
			}

			url := GenerateRoute(auth, plan)

			log.Tracef("URL: %v", color.Blue.Sprint(url))

			if result, err := FastRequest(
				url, fasthttp.MethodPost, NewRetrievalRequestBody(plan.Metadata), headers,
			); err == nil {
				ch <- NlxRetrievalResult{
					Plan:     plan,
					PlanName: name,
					Url:      url,
					Data:     result,
				}
			} else {
				log = log.WithFields(logrus.Fields{
					"plan":     plan,
					"planName": name,
					"url":      url,
				})
				log = log.WithError(err)
				log.Errorf("error with %v request", color.Magenta.Sprint(name))
				return err
			}

			return nil
		}
	}

	// fire off the threads
	eg.Go(executePlan("Warden", plans.CloudDataService))
	eg.Go(executePlan("Pliny", plans.MetadataService))

	// fire off another thread that polls for the conclusion of
	// the waitgroup, then closes the channel. The following lines (for range := chan)
	// is blocking until close(chan) is called, so this frees it up once that's done.
	go func() {
		err := eg.Wait()
		close(ch)
		// if there's an error, we won't consume the results below
		// and we'll output the error
		if err != nil {
			logging.Get().Errorf("Error when executing retrieval plan: %v", err)
		}
	}()

	// consume the closed channel (probably return an array; todo)
	for result := range ch {
		logging.Get().Tracef("%v\n", result)
		results = append(results, result)
	}

	return
}

// NlxRetrievalResult is only used

func (x NlxRetrievalResult) String() string {
	return fmt.Sprintf("(%v => %v (size: %v))", x.PlanName, x.Url, len(x.Data))
}
