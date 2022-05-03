package nlx_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
)

type NlxDeploymentPlans map[string]nlx.NlxPlan

var (
	DeployPlanRoute = fmt.Sprintf("api/%v/metadata/deploy/plan", nlx.DEFAULT_API_VERSION)
)

type NlxDeploymentPlanRetrievalPayload struct {
	AppName   string             `json:"appName,omitempty"`
	PlanBytes NlxDeploymentPlans `json:"plan"`
}

func GetDeployPlan(auth *nlx.Authorization, filter *NlxDeploymentPlanRetrievalPayload) (duration time.Duration, results NlxDeploymentPlans, err error) {
	logging.VerboseSection("Getting Deploy Plan")
	start := time.Now()
	defer func() { duration = time.Since(start) }()

	// pliny request, use access token
	headers := nlx.GenerateHeaders(auth.Host, auth.AccessToken)
	headers[fasthttp.HeaderContentType] = nlx.ZIP_CONTENT_TYPE

	var body []byte = make([]byte, 0)
	if filter != nil {
		if body, err = json.Marshal(filter); err != nil {
			return
		}
	}

	// make the request
	results, err = nlx.FastJsonBodyRequest[NlxDeploymentPlans](
		fmt.Sprintf("%s/%s", auth.Host, DeployPlanRoute),
		fasthttp.MethodPost,
		body,
		headers,
	)

	return
}

/*
currentDirectory, err := os.Getwd()
		if err != nil {
			logging.PrintError("Unable to get working directory", err)
			return
		}

		defer func() {
			err := os.Chdir(currentDirectory)
			if err != nil {
				logging.PrintError("Unable to change directory", err)
				log.Fatal(err)
			}
		}()

		var targetDir string
		if targetDir, err = cmd.Flags().GetString(flags.Directory.Name); err != nil {
			return
		}

		// If target directory is provided,
		// switch to that target directory and later switch back.
		if targetDir != "" {
			err = os.Chdir(targetDir)
			if err != nil {
				logging.PrintError("Unable to change working directory", err)
				return
			}
		}

		dotDir := "."
		currDir, err = filepath.Abs(filepath.Dir(dotDir))
		if err != nil {
			logging.PrintError("Unable to form filepath", err)
			return
		}

		logging.VerboseLn("Deploying site from", currDir)

		// Create a buffer to write our archive to.
		bufPlan := new(bytes.Buffer)
		err = pkg.Archive(".", bufPlan, nil)
		if err != nil {
			logging.PrintError("Error creating deployment ZIP archive", err)
			return
		}

		var deployPlan io.Reader
		mimeType := "application/zip"

		var appName string
		if appName, err = cmd.Flags().GetString(flags.AppName.Name); err != nil {
			return
		}

		if appName != "" {
			filter := pkg.DeployFilter{
				AppName: appName,
				Plan:    bufPlan.Bytes(),
			}
			deployBytes, err := json.Marshal(filter)
			if err != nil {
				logging.PrintError("Error creating deployment plan payload", err)
				return err
			}
			deployPlan = bytes.NewReader(deployBytes)
			mimeType = "application/json"
		} else {
			deployPlan = bufPlan
		}

		plan, err := api.GetDeployPlan(deployPlan, mimeType)
*/

func TestGetDeployPlan(t *testing.T) {

	auth, err := nlx.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	// td := ".deploy/jredhoop-subdomain"
	// // directoryPath := util.GetAbs(td)

	duration, plans, err := GetDeployPlan(auth, nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(duration)

	for k, v := range plans {
		logging.VerboseF("plan k: %v\n", k)
		logging.VerboseF("plan v: %v\n", v)
	}

}
