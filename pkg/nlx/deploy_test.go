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

	duration, _, err = nlx.ExecuteDeployPlan(auth, plans, fp)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(duration)

}

func BenchmarkDeploymentPlan(b *testing.B) {
	util.SkipBenchmark(b)
	logging.SetVerbose(false)
	logging.SetDebug(false)
	auth, _ := nlx.Authorize(authHost, authUser, authPass)
	wd, _ := os.Getwd()
	fp := filepath.Join(wd, "..", "..", "_deploy")
	deploymentPlan, _ := nlx.Archive(fp, nil)
	_, plans, _ := nlx.PrepareDeployment(auth, deploymentPlan, nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = nlx.ExecuteDeployPlan(auth, plans, fp)
	}
}
