package errutil_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/skuid/skuid-cli/pkg/errutil"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	errMsg := "panic has set in"

	// panics
	assert.PanicsWithError(t, errMsg, func() {
		errutil.Must(errors.New(errMsg))
	})

	// does not panic
	errutil.Must(nil)
}

func TestMustConditionf(t *testing.T) {
	baseMsg := "panic has set in %v"
	arg := "hello"
	errMsg := fmt.Sprintf(baseMsg, arg)

	// panics
	assert.PanicsWithError(t, errMsg, func() {
		errutil.MustConditionf(false, baseMsg, arg)
	})

	// does not panic
	errutil.MustConditionf(true, baseMsg, arg)
}
