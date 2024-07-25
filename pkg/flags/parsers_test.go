package flags_test

import (
	"testing"
	"time"

	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseSince(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		testDescription string
		giveValue       string
		giveNoFuture    bool
		wantValue       flags.CustomString
		wantError       bool
	}{
		{"valid value", now.Format(time.RFC3339Nano), true, flags.CustomString(util.FormatTimestamp(now)), false},
		{"invalid value future", "-1d", false, flags.CustomString(""), true},
		{"invalid value empty", "", false, flags.CustomString(""), true},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := flags.ParseSince(tc.giveValue)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantValue, actualValue)
		})
	}
}
