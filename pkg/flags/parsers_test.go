package flags_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSince(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		testDescription string
		giveValue       string
		wantValue       *time.Time
		wantError       bool
	}{
		{"valid value", now.Format(time.RFC3339Nano), &now, false},
		{"valid value spaces around", fmt.Sprintf("  %v  ", now.Format(time.RFC3339Nano)), &now, false},
		{"valid value empty", "", nil, false},
		{"invalid value future", "-1d", nil, true},
		{"invalid value blank", "   ", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := flags.ParseSince(tc.giveValue)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantValue == nil {
				assert.Nil(t, actualValue)
			} else {
				require.NotNil(t, actualValue)
				assert.True(t, tc.wantValue.Equal(*actualValue))
			}
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

func TestParseLogLevel(t *testing.T) {
	expectError := func(val string) error {
		return fmt.Errorf("not a valid log level: %q", val)
	}

	testCases := []struct {
		testDescription string
		giveValue       string
		wantValue       logrus.Level
		wantError       error
	}{
		{
			testDescription: "empty",
			giveValue:       "",
			wantError:       expectError(""),
		},
		{
			testDescription: "blank",
			giveValue:       "    ",
			wantError:       expectError("    "),
		},
		{
			testDescription: "foobar",
			giveValue:       "foobar",
			wantError:       expectError("foobar"),
		},
		{
			testDescription: "space within",
			giveValue:       "in fo",
			wantError:       expectError("in fo"),
		},
		{
			testDescription: "space around",
			giveValue:       "  info   ",
			wantValue:       logrus.InfoLevel,
		},
		{
			testDescription: "trace",
			giveValue:       "trace",
			wantValue:       logrus.TraceLevel,
		},
		{
			testDescription: "debug",
			giveValue:       "debug",
			wantValue:       logrus.DebugLevel,
		},
		{
			testDescription: "info",
			giveValue:       "info",
			wantValue:       logrus.InfoLevel,
		},
		{
			testDescription: "warn",
			giveValue:       "warn",
			wantValue:       logrus.WarnLevel,
		},
		{
			testDescription: "warning",
			giveValue:       "warning",
			wantValue:       logrus.WarnLevel,
		},
		{
			testDescription: "error",
			giveValue:       "error",
			wantValue:       logrus.ErrorLevel,
		},
		{
			testDescription: "fatal",
			giveValue:       "fatal",
			wantValue:       logrus.FatalLevel,
		},
		{
			testDescription: "panic",
			giveValue:       "panic",
			wantValue:       logrus.PanicLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := flags.ParseLogLevel(tc.giveValue)
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantValue, actualValue)
			}
		})
	}
}

func TestParseMetadataEntity(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveValue       string
		wantValue       metadata.MetadataEntity
		wantError       bool
	}{
		{
			testDescription: "empty",
			giveValue:       "",
			wantError:       true,
		},
		{
			testDescription: "blank",
			giveValue:       "      ",
			wantError:       true,
		},
		{
			testDescription: "invalid format valid metadata type",
			giveValue:       "pages",
			wantError:       true,
		},
		{
			testDescription: "invalid format invalid metadata type",
			giveValue:       "unknownentityformat",
			wantError:       true,
		},
		{
			testDescription: "invalid metadata type",
			giveValue:       "unknown/my_entity",
			wantError:       true,
		},
		{
			testDescription: "invalid entity name for valid metadata type",
			giveValue:       "pages/my_entity.json",
			wantError:       true,
		},
		{
			testDescription: "valid entity",
			giveValue:       "pages/my_entity",
			wantValue:       metadata.MetadataEntity{Type: metadata.MetadataTypePages, SubType: metadata.MetadataSubTypeNone, Name: "my_entity", Path: "pages/my_entity", PathRelative: "my_entity"},
		},
		{
			testDescription: "valid nested entity",
			giveValue:       "site/favicon/my_entity.ico",
			wantValue:       metadata.MetadataEntity{Type: metadata.MetadataTypeSite, SubType: metadata.MetadataSubTypeSiteFavicon, Name: "my_entity.ico", Path: "site/favicon/my_entity.ico", PathRelative: "favicon/my_entity.ico"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := flags.ParseMetadataEntity(tc.giveValue)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantValue, actualValue)
		})
	}
}

func TestParseDirectory(t *testing.T) {
	wd, err := os.Getwd()
	require.Nil(t, err)

	var vol string
	if runtime.GOOS == "windows" {
		vol = filepath.VolumeName(wd)
	}

	testCases := []struct {
		testDescription string
		giveValue       string
		wantValue       string
		wantError       bool
	}{
		{
			testDescription: "empty",
			giveValue:       "",
			wantValue:       wd,
		},
		{
			testDescription: "relative path .",
			giveValue:       ".",
			wantValue:       wd,
		},
		{
			testDescription: "relative path . with segment",
			giveValue:       "./foo",
			wantValue:       filepath.Join(wd, "foo"),
		},
		{
			testDescription: "relative path ..",
			giveValue:       "..",
			wantValue:       filepath.Join(wd, ".."),
		},
		{
			testDescription: "relative path .. with segment",
			giveValue:       "../foo",
			wantValue:       filepath.Join(wd, "..", "foo"),
		},
		{
			testDescription: "relative path single segment",
			giveValue:       "foo",
			wantValue:       filepath.Join(wd, "foo"),
		},
		{
			testDescription: "relative path multi segment",
			giveValue:       "foo/bar/mydir",
			wantValue:       filepath.Join(wd, "foo", "bar", "mydir"),
		},
		{
			testDescription: "absolute path single segment",
			giveValue:       "/foo",
			wantValue:       vol + filepath.FromSlash("/foo"),
		},
		{
			testDescription: "absolute path multi segment",
			giveValue:       "/foo/bar/mydir",
			wantValue:       vol + filepath.FromSlash("/foo/bar/mydir"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue, err := flags.ParseDirectory(tc.giveValue)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantValue, actualValue)
		})
	}
}
