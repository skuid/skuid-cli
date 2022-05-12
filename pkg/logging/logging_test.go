package logging_test

import (
	"testing"

	"github.com/skuid/tides/pkg/logging"
)

func TestDebug(t *testing.T) {
	// this test only exists so we don't
	// send "DEBUG=TRUE" to production
	if logging.GetDebug() {
		t.Fail()
	}
}
