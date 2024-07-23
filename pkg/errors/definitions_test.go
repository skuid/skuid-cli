package errors_test

import (
	"fmt"
	"testing"

	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	errMsg := "panic has set in"

	// panics
	assert.PanicsWithError(t, errMsg, func() {
		errors.Must(errors.Critical(errMsg))
	})

	// does not panic
	errors.Must(nil)
}

func TestMustConditionf(t *testing.T) {
	baseMsg := "panic has set in %v"
	arg := "hello"
	errMsg := fmt.Sprintf(baseMsg, arg)

	// panics
	assert.PanicsWithError(t, errMsg, func() {
		errors.MustConditionf(false, baseMsg, arg)
	})

	// does not panic
	errors.MustConditionf(true, baseMsg, arg)
}
