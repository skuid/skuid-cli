package pkg_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/util"
)

func TestHttpMethods(t *testing.T) {
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
			actual, actualError := pkg.JsonBodyRequest[YesNoResponse](
				tc.givenHost,
				http.MethodGet,
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

// roundTripFunc stands in for the network so a test can choose the response
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRequestErrorRedactsCredentials(t *testing.T) {
	const token = "secret-token"

	for _, tc := range []struct {
		description      string
		givenStatusCode  int
		givenBody        string
		expectedRequests int
	}{
		{
			description:      "cloudflare timeout",
			givenStatusCode:  524,
			givenBody:        "error code: 524",
			expectedRequests: 1,
		},
		{
			description:      "unauthorized after retry",
			givenStatusCode:  http.StatusUnauthorized,
			givenBody:        "unauthorized",
			expectedRequests: 1 + pkg.MAX_AUTHORIZATION_ATTEMPTS,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			// requests go through a zero http.Client, which uses http.DefaultTransport
			var sentAuthorization []string
			defaultTransport := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = defaultTransport })
			http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				sentAuthorization = append(sentAuthorization, req.Header.Get(pkg.HeaderAuthorization))
				return &http.Response{
					StatusCode: tc.givenStatusCode,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tc.givenBody)),
					Request:    req,
				}, nil
			})

			headers := pkg.GenerateHeaders("example.skuidsite.com", token)
			_, err := pkg.Request(
				fmt.Sprintf("example.skuidsite.com/%v", pkg.DeployPlanRoute),
				http.MethodPost,
				[]byte{},
				headers,
			)

			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.givenBody)
				assert.Contains(t, err.Error(), pkg.REDACTED_HEADER_VALUE)
				assert.NotContains(t, err.Error(), token)
			}

			// the real token is still sent, on retries too
			assert.Equal(t, fmt.Sprintf("Bearer %v", token), headers[pkg.HeaderAuthorization])
			assert.Len(t, sentAuthorization, tc.expectedRequests)
			for _, sent := range sentAuthorization {
				assert.Equal(t, fmt.Sprintf("Bearer %v", token), sent)
			}
		})
	}
}
