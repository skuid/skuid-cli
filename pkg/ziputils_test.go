package pkg_test

import (
	"bytes"
	"testing"

	"github.com/skuid/tides/pkg"
)

func TestArchive(t *testing.T) {
	var buffer bytes.Buffer
	pkg.Archive("a", &buffer, &pkg.Metadata{})
}
