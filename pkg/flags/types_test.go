package flags_test

import (
	"testing"

	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FlagTestSuite struct {
	suite.Suite
}

func (suite *FlagTestSuite) TestEnvVarName() {
	testCases := []struct {
		testDescription string
		giveFlag        *flags.Flag[string]
		wantName        string
	}{
		{
			testDescription: "default",
			giveFlag:        &flags.Flag[string]{},
			wantName:        "SKUID_",
		},
		{
			testDescription: "empty string",
			giveFlag:        &flags.Flag[string]{Name: ""},
			wantName:        "SKUID_",
		},
		{
			testDescription: "simple name",
			giveFlag:        &flags.Flag[string]{Name: "dir"},
			wantName:        "SKUID_DIR",
		},
		{
			testDescription: "with underscore",
			giveFlag:        &flags.Flag[string]{Name: "dir_name"},
			wantName:        "SKUID_DIR_NAME",
		},
		{
			testDescription: "with hyphen",
			giveFlag:        &flags.Flag[string]{Name: "dir-name"},
			wantName:        "SKUID_DIR_NAME",
		},
		{
			testDescription: "with double hyphen",
			giveFlag:        &flags.Flag[string]{Name: "dir--name"},
			wantName:        "SKUID_DIR__NAME",
		},
		{
			testDescription: "with multiple hyphen",
			giveFlag:        &flags.Flag[string]{Name: "log-dir-name"},
			wantName:        "SKUID_LOG_DIR_NAME",
		},
		{
			testDescription: "mix case",
			giveFlag:        &flags.Flag[string]{Name: "lOG-dir-nAMe"},
			wantName:        "SKUID_LOG_DIR_NAME",
		},
		{
			testDescription: "upper case",
			giveFlag:        &flags.Flag[string]{Name: "LOG-DIR-NAME"},
			wantName:        "SKUID_LOG_DIR_NAME",
		},
		{
			testDescription: "special chars",
			giveFlag:        &flags.Flag[string]{Name: "LoG@DIr$NAmE-^#!()"},
			wantName:        "SKUID_LOG@DIR$NAME_^#!()",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			actualName := tc.giveFlag.EnvVarName()
			assert.Equal(t, tc.wantName, actualName)
		})
	}
}

func (suite *FlagTestSuite) TestLegacyEnvVarNames() {
	testCases := []struct {
		testDescription string
		giveFlag        *flags.Flag[string]
		wantNames       []string
	}{
		{
			testDescription: "default",
			giveFlag:        &flags.Flag[string]{},
			wantNames:       []string{},
		},
		{
			testDescription: "empty string",
			giveFlag:        &flags.Flag[string]{LegacyEnvVars: []string{""}},
			wantNames:       []string{""},
		},
		{
			testDescription: "single name",
			giveFlag:        &flags.Flag[string]{LegacyEnvVars: []string{"LEGACY_VAR"}},
			wantNames:       []string{"LEGACY_VAR"},
		},
		{
			testDescription: "multiple names",
			giveFlag:        &flags.Flag[string]{LegacyEnvVars: []string{"LEGACY_VAR", "FOO_VAR"}},
			wantNames:       []string{"LEGACY_VAR", "FOO_VAR"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			actualNames := tc.giveFlag.LegacyEnvVarNames()
			assert.ElementsMatch(t, tc.wantNames, actualNames)
		})
	}
}

func (suite *FlagTestSuite) TestIsRequired() {
	testCases := []struct {
		testDescription string
		giveFlag        *flags.Flag[string]
		wantRequired    bool
	}{
		{
			testDescription: "default",
			giveFlag:        &flags.Flag[string]{},
			wantRequired:    false,
		},
		{
			testDescription: "false",
			giveFlag:        &flags.Flag[string]{Required: false},
			wantRequired:    false,
		},
		{
			testDescription: "true",
			giveFlag:        &flags.Flag[string]{Required: true},
			wantRequired:    true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			actualRequired := tc.giveFlag.IsRequired()
			assert.Equal(t, tc.wantRequired, actualRequired)
		})
	}
}

func TestFlagTestSuite(t *testing.T) {
	suite.Run(t, new(FlagTestSuite))
}

type GetPasswordSuite struct {
	suite.Suite
}

func (suite *GetPasswordSuite) TestGetPasswordSuccess() {
	t := suite.T()
	f := flags.Password
	expectedRedactedValue := "!!!!"
	expectedPassword := "foobar"
	expectedBlankValue := "hellothere"

	fs := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	p := new(flags.RedactedString)
	v := flags.NewValueWithRedact(f.Default, p, func(val string) (flags.RedactedString, error) { return flags.RedactedString(val), nil }, func(rs flags.RedactedString) string {
		if rs == "" {
			return expectedBlankValue
		} else {
			return expectedRedactedValue
		}
	})
	fs.VarPF(v, f.Name, f.Shorthand, f.Usage)
	actualFlag := fs.Lookup(f.Name)
	require.NotNil(t, actualFlag)
	assert.Equal(t, actualFlag.Value.String(), expectedBlankValue)

	fs.Set(f.Name, expectedPassword)
	assert.Equal(t, expectedRedactedValue, actualFlag.Value.String())

	actualPassword, err := flags.GetPassword(fs)
	require.NoError(t, err)
	assert.Equal(t, expectedPassword, actualPassword.Unredacted().String())
}

func (suite *GetPasswordSuite) TestGetPasswordErrorNotExist() {
	t := suite.T()

	fs := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	actualPassword, err := flags.GetPassword(fs)
	assert.ErrorContains(t, err, "code=1")
	assert.Nil(t, actualPassword)
}

func (suite *GetPasswordSuite) TestGetPasswordErrorWrongType() {
	t := suite.T()

	fs := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	fs.StringP(flags.Password.Name, flags.Password.Shorthand, "", flags.Password.Usage)
	passwordFlag := fs.Lookup(flags.Password.Name)
	require.NotNil(t, passwordFlag)

	actualPassword, err := flags.GetPassword(fs)
	assert.ErrorContains(t, err, "code=2")
	assert.Nil(t, actualPassword)
}

func TestGetPasswordSuite(t *testing.T) {
	suite.Run(t, new(GetPasswordSuite))
}
