package pkg_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg"
)

func TestZip(t *testing.T) {
	cd, _ := os.Getwd()
	relpath := filepath.Join(cd, "..", "..", "_deploy")
	bb, err := pkg.Archive(relpath, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(bb))

}
