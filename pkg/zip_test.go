package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
)

func TestZip(t *testing.T) {
	util.SkipIntegrationTest(t)
	cd, _ := os.Getwd()
	relpath := filepath.Join(cd, ".", ".", "_deploy")
	bb, err := pkg.Archive(relpath, nil)
	if err != nil {
		logging.Get().Fatal(err)
	}
	logging.Get().Info(len(bb))

}
