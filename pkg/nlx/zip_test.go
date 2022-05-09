package nlx_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg/nlx"
)

func TestZip(t *testing.T) {
	cd, _ := os.Getwd()
	relpath := filepath.Join(cd, "..", "..", "_deploy")
	bb, err := nlx.Archive(relpath, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(bb))

}
