package pkg_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
)

func TestRedactHeaders(t *testing.T) {
	auth := &pkg.Authorization{
		Host:               "example.skuidsite.com",
		AccessToken:        "pliny-access-token",
		AuthorizationToken: "warden-jwt",
	}

	for _, tc := range []struct {
		description string
		given       pkg.RequestHeaders
		expected    pkg.RequestHeaders
	}{
		{
			description: "nil headers",
			given:       nil,
			expected:    pkg.RequestHeaders{},
		},
		{
			description: "no credentials",
			given: pkg.RequestHeaders{
				pkg.HeaderContentType:     pkg.JSON_CONTENT_TYPE,
				pkg.HeaderContentEncoding: pkg.GZIP_CONTENT_ENCODING,
			},
			expected: pkg.RequestHeaders{
				pkg.HeaderContentType:     pkg.JSON_CONTENT_TYPE,
				pkg.HeaderContentEncoding: pkg.GZIP_CONTENT_ENCODING,
			},
		},
		{
			description: "authorization",
			given: pkg.RequestHeaders{
				pkg.HeaderAuthorization: "Bearer secret-token",
				pkg.HeaderContentType:   pkg.ZIP_CONTENT_TYPE,
			},
			expected: pkg.RequestHeaders{
				pkg.HeaderAuthorization: pkg.REDACTED_HEADER_VALUE,
				pkg.HeaderContentType:   pkg.ZIP_CONTENT_TYPE,
			},
		},
		{
			description: "lowercase authorization",
			given:       pkg.RequestHeaders{"authorization": "Bearer secret-token"},
			expected:    pkg.RequestHeaders{"authorization": pkg.REDACTED_HEADER_VALUE},
		},
		{
			description: "uppercase authorization",
			given:       pkg.RequestHeaders{"AUTHORIZATION": "Bearer secret-token"},
			expected:    pkg.RequestHeaders{"AUTHORIZATION": pkg.REDACTED_HEADER_VALUE},
		},
		{
			description: "proxy authorization",
			given:       pkg.RequestHeaders{pkg.HeaderProxyAuthorization: "Basic secret"},
			expected:    pkg.RequestHeaders{pkg.HeaderProxyAuthorization: pkg.REDACTED_HEADER_VALUE},
		},
		{
			description: "lowercase proxy authorization",
			given:       pkg.RequestHeaders{"proxy-authorization": "Basic secret"},
			expected:    pkg.RequestHeaders{"proxy-authorization": pkg.REDACTED_HEADER_VALUE},
		},
		{
			description: "pliny plan headers",
			given:       pkg.GeneratePlanHeaders(auth, pkg.NlxPlan{}),
			expected: pkg.RequestHeaders{
				pkg.HeaderAuthorization:              pkg.REDACTED_HEADER_VALUE,
				pkg.HEADER_SKUID_PUBLIC_KEY_ENDPOINT: "https://example.skuidsite.com/api/v1/site/verificationkey",
			},
		},
		{
			description: "warden plan headers",
			given:       pkg.GeneratePlanHeaders(auth, pkg.NlxPlan{Host: "warden.example.com"}),
			expected: pkg.RequestHeaders{
				pkg.HeaderAuthorization:              pkg.REDACTED_HEADER_VALUE,
				pkg.HEADER_SKUID_PUBLIC_KEY_ENDPOINT: "https://example.skuidsite.com/api/v1/site/verificationkey",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			before := fmt.Sprint(tc.given)

			actual := pkg.RedactHeaders(tc.given)
			assert.Equal(t, tc.expected, actual)

			// the given headers are the ones sent with the request: they must keep
			// their real values and must not share storage with the redacted copy
			assert.Equal(t, before, fmt.Sprint(tc.given))
			actual["X-Redact-Test"] = "copy"
			assert.NotContains(t, tc.given, "X-Redact-Test")
		})
	}
}
