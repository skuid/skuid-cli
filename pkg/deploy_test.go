package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/util"
)

func TestGetDeployPlan(t *testing.T) {
	util.SkipIntegrationTest(t)

	auth, err := pkg.Authorize(authOptions)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	var fp string
	envdir := os.Getenv(constants.ENV_SKUID_DEFAULT_FOLDER)
	if envdir != "" {
		fp = envdir
	} else {
		wd, _ := os.Getwd()
		fp = filepath.Join(wd, "..", "..", "_deploy")
	}

	plans, err := pkg.GetDeployPlan(auth, fp, nil, nil, nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	_, err = pkg.ExecuteDeployPlan(auth, plans, fp)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func BenchmarkDeploymentPlan(b *testing.B) {
	util.SkipBenchmark(b)
	auth, _ := pkg.Authorize(authOptions)
	wd, _ := os.Getwd()
	fp := filepath.Join(wd, ".", ".", "_deploy")
	plans, _ := pkg.GetDeployPlan(auth, fp, nil, nil, nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = pkg.ExecuteDeployPlan(auth, plans, fp)
	}
}
