package nlx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
	"github.com/skuid/tides/pkg/util"
)

func TestGetDeployPlan(t *testing.T) {
	util.SkipIntegrationTest(t)

	auth, err := nlx.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	wd, _ := os.Getwd()
	fp := filepath.Join(wd, "..", "..", "_deploy")

	deploymentPlan, err := nlx.Archive(fp, nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	duration, plans, err := nlx.PrepareDeployment(auth, deploymentPlan, nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(duration)

	for k, v := range plans {
		logging.DebugF("plan k: %v\n", k)
		logging.DebugF("plan v: %v\n", v)
	}

	metadataFilter := &nlx.NlxMetadata{}
	deploymentPlan, err = nlx.Archive(fp, metadataFilter)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	duration, plans, err = nlx.PrepareDeployment(auth, deploymentPlan, &nlx.DeploymentFilter{AppName: "alpha"})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(duration)

	for k, v := range plans {
		logging.DebugF("plan k: %v\n", k)
		logging.DebugF("plan v: %v\n", v)
	}

}
