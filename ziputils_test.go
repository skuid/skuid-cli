package main_test

import (
	"bytes"
	"testing"

	tides "github.com/skuid/tides"
)

func TestArchive(t *testing.T) {
	var buffer bytes.Buffer
	tides.Archive("a", &buffer, &tides.Metadata{})
}
