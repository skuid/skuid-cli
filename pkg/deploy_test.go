package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/util"
)

func TestGetDeployPlan(t *testing.T) {
	util.SkipIntegrationTest(t)

	auth, err := pkg.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	wd, _ := os.Getwd()
	fp := filepath.Join(wd, "..", "..", "_deploy")

	deploymentPlan, err := pkg.Archive(fp, nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	duration, plans, err := pkg.PrepareDeployment(auth, deploymentPlan, nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(duration)

	duration, _, err = pkg.ExecuteDeployPlan(auth, plans, fp)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(duration)

}

func BenchmarkDeploymentPlan(b *testing.B) {
	util.SkipBenchmark(b)
	auth, _ := pkg.Authorize(authHost, authUser, authPass)
	wd, _ := os.Getwd()
	fp := filepath.Join(wd, ".", ".", "_deploy")
	deploymentPlan, _ := pkg.Archive(fp, nil)
	_, plans, _ := pkg.PrepareDeployment(auth, deploymentPlan, nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = pkg.ExecuteDeployPlan(auth, plans, fp)
	}
}
