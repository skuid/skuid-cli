package nlx_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg/nlx"
	"github.com/skuid/tides/pkg/util"
)

func TestFasthttpMethods(t *testing.T) {
	util.SkipIntegrationTest(t)
	const YES_NO_API = "yesno.wtf/api"

	type YesNoResponse struct {
		Answer string `json:"answer"`
		Forced bool   `json:"forced"`
	}

	for _, tc := range []struct {
		description  string
		givenHost    string
		givenHeaders map[string]string // should ignore nil map

		expectedErrorMsg string
	}{
		{
			description: "https",
			givenHost:   fmt.Sprintf("https://%v", YES_NO_API),
		},
		{
			description:  "empty header map",
			givenHost:    fmt.Sprintf("https://%v", YES_NO_API),
			givenHeaders: make(map[string]string),
		},
		{
			description: "no https, should add",
			givenHost:   YES_NO_API,
		},
		{
			description: "http should replace",
			givenHost:   fmt.Sprintf("http://%v", YES_NO_API),
		},
		{
			description:      "301 redirect to 404 page",
			givenHost:        "https://skuid.com/google",
			expectedErrorMsg: "301",
		},
		{
			description:      "bad unmarshal",
			givenHost:        "https://www.uuidtools.com/api/generate/v1",
			expectedErrorMsg: "json: cannot unmarshal",
		},
	} {
		t.Run(tc.description, func(subtest *testing.T) {
			actual, actualError := nlx.FastJsonBodyRequest[YesNoResponse](
				tc.givenHost,
				fasthttp.MethodGet,
				[]byte{},
				tc.givenHeaders,
			)

			if actualError != nil && tc.expectedErrorMsg == "" {
				subtest.Log(actualError)
				subtest.FailNow()
			}

			if tc.expectedErrorMsg == "" && actual.Answer == "" {
				subtest.Log(actualError)
				subtest.FailNow()
			}

			if tc.expectedErrorMsg != "" {
				assert.Contains(subtest, actualError.Error(), tc.expectedErrorMsg)
			}
		})
	}

}
