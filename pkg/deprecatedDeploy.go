package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gookit/color"

	"github.com/skuid/tides/pkg/logging"
)

type NlxApi struct {
	Connection *NlxConnection
}

// GetDeployPlan fetches a deploymnent plan from Skuid Platform API
func (api *NlxApi) GetDeployPlan(payload io.Reader, mimeType string) (plans map[string]Plan, err error) {
	logging.VerboseSection("Getting Deploy Plan")

	if mimeType == "" {
		mimeType = "application/zip"
	}

	planStart := time.Now()
	// Get a deploy plan
	planResult, err := api.Connection.MakeAccessTokenRequest(
		http.MethodPost,
		"/metadata/deploy/plan",
		payload,
		mimeType,
	)

	if err != nil {
		return
	}

	logging.VerboseSuccess("Success Getting Deploy Plan", planStart)

	if err = json.Unmarshal(planResult, &plans); err != nil {
		return
	}

	return
}

// ExecuteDeployPlan executes a map of plan items in a deployment plan
func (api *NlxApi) ExecuteDeployPlan(plans map[string]Plan, targetDir string) (planResults [][]byte, err error) {

	logging.VerboseSection("Executing Deploy Plan")

	// eg := &errgroup.Group{}
	// ch := make(chan *io.ReadCloser)

	// for _, plan := range plans {
	// 		p := plan
	// 	eg.Go(func() error {
	// 		planResult, err := api.ExecutePlanItem(p, targetDir, verbose)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		ch <- planResult
	// 		return nil
	// 	})
	// }

	// go func() {
	// 	eg.Wait()
	// 	close(ch)
	// }()

	// planResults := []*io.ReadCloser{}
	// for pr := range ch {
	// 	planResults = append(planResults, pr)
	// }

	planResults = make([][]byte, 0)
	for _, plan := range plans {
		var planResult []byte
		planResult, err = api.ExecutePlanItem(plan, targetDir)
		if err != nil {
			return
		}
		planResults = append(planResults, planResult)
	}

	return
}

// ExecutePlanItem executes a particular item in a deployment plan
func (api *NlxApi) ExecutePlanItem(plan Plan, targetDir string) (result []byte, err error) {
	// Create a buffer to write our archive to.
	var planResult []byte
	bufDeploy := new(bytes.Buffer)
	err = Archive(targetDir, bufDeploy, &plan.Metadata)
	if err != nil {
		log.Print("Error creating deployment ZIP archive")
		log.Fatal(err)
	}

	deployStart := time.Now()

	if plan.Host == "" {
		logging.VerboseLn(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", plan.Url, plan.Type))

		planResult, err = api.Connection.MakeAccessTokenRequest(
			http.MethodPost,
			plan.Url,
			bufDeploy,
			"application/zip",
		)
		if err != nil {
			return nil, err
		}
	} else {

		url := fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.Url)
		logging.VerboseF("Making Deploy Request: URL: [%s] Type: [%s]", color.Yellow.Sprint(url), color.Cyan.Sprint(plan.Type))

		planResult, err = api.Connection.MakeAuthorizationBearerRequest(
			http.MethodPost,
			url,
			bufDeploy,
			"application/zip",
		)
		if err != nil {
			return
		}

	}

	logging.VerboseSuccess("Success Deploying to Source", deployStart)

	return planResult, nil
}