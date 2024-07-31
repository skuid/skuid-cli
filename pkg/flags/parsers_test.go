package flags_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSince(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		testDescription string
		giveValue       string
		wantValue       string
		wantError       bool
	}{
		{"valid value", now.Format(time.RFC3339Nano), util.FormatTimestamp(now), false},
		{"valid value spaces around", fmt.Sprintf("  %v  ", now.Format(time.RFC3339Nano)), util.FormatTimestamp(now), false},
		{"valid value empty", "", "", false},
		{"invalid value future", "-1d", "", true},
		{"invalid value blank", "   ", "", true},
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

func TestParseHost(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveHost        string
		wantHost        string
		wantError       error
	}{
		{
			testDescription: "https base url subdomain without trailing slash",
			giveHost:        "https://test.skuidsite.com",
			wantHost:        "https://test.skuidsite.com",
			wantError:       nil,
		},
		{
			testDescription: "https base url subdomain with trailing slash",
			giveHost:        "https://test.skuidsite.com/",
			wantHost:        "https://test.skuidsite.com",
			wantError:       nil,
		},
		{
			testDescription: "https base url domain without trailing slash",
			giveHost:        "https://mycustomdomain.com",
			wantHost:        "https://mycustomdomain.com",
			wantError:       nil,
		},
		{
			testDescription: "https base url domain with trailing slash",
			giveHost:        "https://mycustomdomain.com/",
			wantHost:        "https://mycustomdomain.com",
			wantError:       nil,
		},
		{
			testDescription: "http base url without trailing slash",
			giveHost:        "http://test.skuidsite.com",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "http base url with trailing slash",
			giveHost:        "http://test.skuidsite.com/",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "ftp base url with trailing slash",
			giveHost:        "ftp://test.skuidsite.com/",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "ftp base url with trailing slash",
			giveHost:        "ftp://test.skuidsite.com/",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "contains basic auth",
			giveHost:        "https://user:pass@host.com",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "contains port",
			giveHost:        "https://mycustomdomain.com:8080",
			wantHost:        "https://mycustomdomain.com:8080",
			wantError:       nil,
		},
		{
			testDescription: "https tld only",
			giveHost:        "https://com",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "host name without trailing slash",
			giveHost:        "test.skuidsite.com",
			wantHost:        "https://test.skuidsite.com",
			wantError:       nil,
		},
		{
			testDescription: "host name with trailing slash",
			giveHost:        "test.skuidsite.com/",
			wantHost:        "https://test.skuidsite.com",
			wantError:       nil,
		},
		{
			testDescription: "host name tld only",
			giveHost:        "com",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "host name domain.tld only",
			giveHost:        "mycustomdomain.com",
			wantHost:        "https://mycustomdomain.com",
			wantError:       nil,
		},
		{
			testDescription: "url with non-root path",
			giveHost:        "https://test.skuidsite.com/test",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "host with non-root path",
			giveHost:        "test.skuidsite.com/test",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "not a url or host",
			giveHost:        "../test.skuidsite.com",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "empty string",
			giveHost:        "",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "no trailing slash with fragment",
			giveHost:        "https://test.skuidsite.com#foo",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "trailing slash with fragment",
			giveHost:        "https://test.skuidsite.com/#foo",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "no trailing slash with query",
			giveHost:        "https://test.skuidsite.com?foo",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "trailing slash with query",
			giveHost:        "https://test.skuidsite.com/?foo",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "no trailing slash with query and fragment",
			giveHost:        "https://test.skuidsite.com?foo#bar",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
		{
			testDescription: "trailing slash with query and fragment",
			giveHost:        "https://test.skuidsite.com/?foo#bar",
			wantHost:        "",
			wantError:       flags.ErrInvalidHost,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualHost, err := flags.ParseHost(tc.giveHost)
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err, "Expected ParseHost err to be nil, got not nil")
			}
			assert.Equal(t, tc.wantHost, actualHost)
		})
	}
}

func TestParseString(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveOptions     flags.ParseStringOptions
		giveValue       string
		wantError       bool
		wantValue       string
	}{
		{"valid value", flags.ParseStringOptions{AllowBlank: false}, "foobar", false, "foobar"},
		{"valid value with spaces", flags.ParseStringOptions{AllowBlank: false}, "  a   ", false, "  a   "},
		{"valid value blank", flags.ParseStringOptions{AllowBlank: true}, "     ", false, "     "},
		{"valid value empty", flags.ParseStringOptions{AllowBlank: true}, "", false, ""},
		{"invalid value blank", flags.ParseStringOptions{AllowBlank: false}, "     ", true, ""},
		{"invalid value empty", flags.ParseStringOptions{AllowBlank: false}, "", true, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := flags.ParseString(tc.giveOptions)(tc.giveValue)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantValue, actualValue)
		})
	}
}
