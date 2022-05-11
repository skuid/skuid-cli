package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/logging"
)

func TestZip(t *testing.T) {
	cd, _ := os.Getwd()
	relpath := filepath.Join(cd, "..", "..", "_deploy")
	bb, err := pkg.Archive(relpath, nil)
	if err != nil {
		logging.Fatal(err)
	}
	logging.Println(len(bb))

}
