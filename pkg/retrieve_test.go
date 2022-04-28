package pkg_test

import (
	"testing"

	"github.com/skuid/tides/pkg"
)

func TestGetRetrievePlan(t *testing.T) {
	for _, tc := range []struct {
		description     string
		givenApi        *pkg.NlxApi
		givenAppName    string
		expectedPlanMap map[string]pkg.Plan
		expectedError   error
	}{
		{
			description: "test",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			// c, d := tides.GetRetrievePlan(a, b)
		})
	}
}
