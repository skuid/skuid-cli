package main_test

import (
	"os"
	"strings"
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

// Skuid Review Required - See https://github.com/skuid/skuid-cli/issues/213
func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	rc := m.Run()

	var threshold float64 = 0

	for _, arg := range os.Args {
		if strings.Contains(arg, "coverfail") {
			if strings.Contains(arg, "true") {
				threshold = 0.8
				color.Green.Printf("failing if coverage is below threshold of %v\n", threshold)
			}
		}
	}

	// rc 0 means we've passed,
	// and CoverMode will be non empty if run with -cover
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < threshold {
			color.Red.Printf("\nTests passed but coverage failed at: %2.2f%%\n", c*100)
			rc = -1
		}
	}
	os.Exit(rc)

}

// There are no tests in `main_test` so `go test` will exit with an exit code of 1 (assuming all other tests pass)
// and a warning that there "no tests to run".  This entire file could be removed since there aren't any tests
// currently but until input is received on https://github.com/skuid/skuid-cli/issues/213, leaving `TestMain` as-is.
// Given that, there needs to be at least one test to avoid the non-zero exit code which is why TestAvoidNoTestsWarning
// exists.
// TODO: Remove this test and possibly entire file once https://github.com/skuid/skuid-cli/issues/213 is addressed
func TestAvoidNoTestsWarning(t *testing.T) {
	assert.True(t, true)
}
