package nlx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"

	"github.com/skuid/tides/pkg/logging"
)

var (
	DeployPlanRoute = fmt.Sprintf("api/%v/metadata/deploy/plan", DEFAULT_API_VERSION)
)

type NlxDynamicPlanMap map[string]NlxPlan

// maybe?
func (df *NlxPlanFilter) Filter(plans NlxDynamicPlanMap) (filtered NlxDynamicPlanMap) {
	// filter by app name
	for k, plan := range plans {
		// filter by app name
		if df.AppName != "" {
			var apps []string
			for _, app := range plan.Metadata.Apps {
				if strings.Contains(app, df.AppName) {
					apps = append(apps, app)
				}
			}
			plan.Metadata.Apps = apps
		}
		filtered[k] = plan
	}
	return
}

type FilteredRequestBody struct {
	AppName   string   `json:"appName"`
	PageNames []string `json:"pageNames"`
	PlanBytes []byte   `json:"plan"`
}

func PrepareDeployment(auth *Authorization, deploymentPlan []byte, filter *NlxPlanFilter) (duration time.Duration, results NlxDynamicPlanMap, err error) {
	logging.VerboseSection("Getting Deploy Plan")
	start := time.Now()
	defer func() { duration = time.Since(start) }()

	// pliny request, use access token
	headers := GenerateHeaders(auth.Host, auth.AccessToken)
	headers[fasthttp.HeaderContentType] = ZIP_CONTENT_TYPE

	var body []byte
	if filter != nil {
		// change content type to json
		headers[fasthttp.HeaderContentType] = JSON_CONTENT_TYPE
		// we instead add the deployment plan bytes to the payload
		// instead of just using that as the payload
		requestBody := FilteredRequestBody{
			filter.AppName,
			filter.PageNames,
			deploymentPlan,
		}
		if body, err = json.Marshal(requestBody); err != nil {
			return
		}
	} else {
		// set the deployment plan as the payload
		body = deploymentPlan
	}

	// make the request
	results, err = FastJsonBodyRequest[NlxDynamicPlanMap](
		fmt.Sprintf("%s/%s", auth.Host, DeployPlanRoute),
		fasthttp.MethodPost,
		body,
		headers,
	)

	return
}

// ExecuteDeployPlan executes a map of plan items in a deployment plan
func ExecuteDeployPlan(auth *Authorization, plans NlxDynamicPlanMap, targetDir string) (duration time.Duration, planResults []NlxDeploymentResult, err error) {
	start := time.Now()
	defer func() { duration = time.Since(start) }()
	logging.VerboseSection("Executing Deploy Plan")

	eg := &errgroup.Group{}
	ch := make(chan NlxDeploymentResult)

	executePlan := func(plan NlxPlan) func() error {
		return func() error {

			logging.VerboseF("Archiving %v", targetDir)
			deploy, err := Archive(targetDir, &plan.Metadata)
			if err != nil {
				logging.Println("Error creating deployment ZIP archive")
				return err
			}

			headers := GeneratePlanHeaders(auth, plan)
			logging.VerboseF("Plan Headers: %v\n", headers)

			url := GenerateRoute(auth, plan)
			logging.VerboseF("Plan Request: %v\n", url)

			if result, err := FastRequest(
				url,
				http.MethodPost,
				deploy,
				headers,
			); err == nil {
				ch <- NlxDeploymentResult{
					Plan: plan,
					Url:  url,
					Data: result,
				}
			} else {
				logging.VerboseF("Url: %v", url)
				logging.VerboseF("Error on request: %v\n", err.Error())
				return err
			}

			return nil

		}
	}

	for _, plan := range plans {
		p := plan
		eg.Go(executePlan(p))
	}

	go func() {
		err := eg.Wait()
		close(ch)
		// if there's an error, we won't consume the results below
		// and we'll output the error
		if err != nil {
			logging.PrintError("Error when executing retrieval plan:", err)
		}
	}()

	planResults = make([]NlxDeploymentResult, 0)
	for planResult := range ch {
		planResults = append(planResults, planResult)
	}

	return
}

type NlxDeploymentResult struct {
	Plan     NlxPlan
	PlanName string
	Url      string
	Data     []byte
}
