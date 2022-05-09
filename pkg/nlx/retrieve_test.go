package nlx_test

import (
	"encoding/json"
	"testing"

	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
	"github.com/skuid/tides/pkg/util"
)

func TestRetrievePlan(t *testing.T) {
	util.SkipIntegrationTest(t)

	logging.SetVerbose(true)
	auth, err := nlx.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		description string

		givenFilter *nlx.NlxRetrievalFilter

		expectedError string
	}{
		{},
	} {
		t.Run(tc.description, func(t *testing.T) {
			duration, result, err := nlx.GetRetrievePlan(auth, tc.givenFilter)
			t.Log(duration)
			t.Log(err)

			data, err := json.Marshal(result)
			t.Log(string(data))
			t.Log(err)

			for _, plan := range []nlx.NlxPlan{
				result.CloudDataService, result.MetadataService,
			} {
				t.Logf("PLAN (%v):", plan.Type)
				b, _ := json.MarshalIndent(plan, "", " ")
				t.Log(string(b))
			}
		})
	}
}

func TestExecuteRetrieval(t *testing.T) {
	util.SkipIntegrationTest(t)

	logging.SetVerbose(true)
	auth, err := nlx.Authorize(authHost, authUser, authPass)
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
			duration, plans, err := nlx.GetRetrievePlan(auth, nil)
			t.Log(duration)
			t.Log(plans)
			t.Log(err)

			duration, results, err := nlx.ExecuteRetrieval(plans, auth, false)
			t.Log(duration)
			t.Log(results)
			t.Log(err)
		})
	}
}
