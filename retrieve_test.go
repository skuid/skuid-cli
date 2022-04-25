package main_test

import (
	"testing"

	tides "github.com/skuid/tides"
)

func TestGetRetrievePlan(t *testing.T) {
	for _, tc := range []struct {
		description     string
		givenApi        *tides.PlatformRestApi
		givenAppName    string
		expectedPlanMap map[string]tides.Plan
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
