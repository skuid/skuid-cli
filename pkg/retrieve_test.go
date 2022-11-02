package pkg_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/util"
)

func TestRetrievePlan(t *testing.T) {
	util.SkipIntegrationTest(t)

	auth, err := pkg.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		description string

		givenFilter *pkg.NlxPlanFilter

		expectedError string
	}{
		{},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, result, err := pkg.GetRetrievePlan(auth, tc.givenFilter)
			assert.NoError(t, err)

			data, err := json.Marshal(result)
			t.Log(string(data))
			assert.NoError(t, err)
		})
	}
}

func TestExecuteRetrieval(t *testing.T) {
	util.SkipIntegrationTest(t)
	auth, err := pkg.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		description string

		expectedError string
	}{
		{},
	} {
		t.Run(tc.description, func(t *testing.T) {
			duration, plans, err := pkg.GetRetrievePlan(auth, nil)
			t.Log(duration)
			t.Log(plans)
			t.Log(err)

			duration, results, err := pkg.ExecuteRetrieval(auth, plans)
			t.Log(duration)
			t.Log(results)
			t.Log(err)
		})
	}
}
