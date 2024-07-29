package flags_test

import (
	"fmt"
	"reflect"
	"strconv"
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

func TestGetRedactedString(t *testing.T) {
	flagName := "myflag"
	testCases := []struct {
		testDescription string
		giveSetup       func(*pflag.FlagSet)
		wantRedacted    string
		wantUnredacted  string
		wantError       error
	}{
		{
			testDescription: "gets RedactedString",
			giveSetup: func(fs *pflag.FlagSet) {
				parse := func(val string) (flags.RedactedString, error) { return flags.RedactedString(val), nil }
				redact := func(rs flags.RedactedString) string { return "***" }
				p := new(flags.RedactedString)
				v := flags.NewValueWithRedact("foobar", p, parse, redact)
				fs.VarPF(v, flagName, "", "")
			},
			wantRedacted:   "***",
			wantUnredacted: "foobar",
			wantError:      nil,
		},
		{
			testDescription: "does not exist in flagset",
			giveSetup: func(fs *pflag.FlagSet) {
			},
			wantRedacted:   "",
			wantUnredacted: "",
			wantError:      fmt.Errorf("flag accessed but not defined: %s", flagName),
		},
		{
			testDescription: "wrong type string",
			giveSetup: func(fs *pflag.FlagSet) {
				fs.String(flagName, "", "")
			},
			wantRedacted:   "",
			wantUnredacted: "",
			wantError:      fmt.Errorf("trying to get %T value of flag of type %s", *new(flags.CustomString), "string"),
		},
		{
			testDescription: "wrong type CustomString",
			giveSetup: func(fs *pflag.FlagSet) {
				parse := func(val string) (flags.CustomString, error) { return flags.CustomString(val), nil }
				p := new(flags.CustomString)
				v := flags.NewValue("default", p, parse)
				fs.VarP(v, flagName, "", "")
			},
			wantRedacted:   "",
			wantUnredacted: "",
			wantError:      fmt.Errorf("trying to get %T value of flag of type %s", *new(flags.RedactedString), *new(flags.CustomString)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			tc.giveSetup(fs)
			s, err := flags.GetRedactedString(fs, flagName)
			if tc.wantError != nil {
				assert.Error(t, err, tc.wantError.Error())
				assert.Nil(t, s)
			} else {
				require.NoError(t, err)
				assert.Equal(t, s.String(), tc.wantRedacted)
				assert.Equal(t, s.Unredacted().String(), tc.wantUnredacted)
			}
		})
	}
}

func TestGetCustomString(t *testing.T) {
	flagName := "myflag"
	testCases := []struct {
		testDescription string
		giveSetup       func(*pflag.FlagSet)
		wantValue       string
		wantError       error
	}{
		{
			testDescription: "gets string",
			giveSetup: func(fs *pflag.FlagSet) {
				parse := func(val string) (flags.CustomString, error) { return flags.CustomString(val), nil }
				p := new(flags.CustomString)
				v := flags.NewValue("default", p, parse)
				fs.VarP(v, flagName, "", "")
			},
			wantValue: "default",
			wantError: nil,
		},
		{
			testDescription: "does not exist in flagset",
			giveSetup: func(fs *pflag.FlagSet) {
			},
			wantValue: "",
			wantError: fmt.Errorf("flag accessed but not defined: %s", flagName),
		},
		{
			testDescription: "wrong type string",
			giveSetup: func(fs *pflag.FlagSet) {
				fs.String(flagName, "", "")
			},
			wantValue: "",
			wantError: fmt.Errorf("trying to get %T value of flag of type %s", *new(flags.CustomString), "string"),
		},
		{
			testDescription: "wrong type RedactedString",
			giveSetup: func(fs *pflag.FlagSet) {
				parse := func(val string) (flags.RedactedString, error) { return flags.RedactedString(val), nil }
				redact := func(rs flags.RedactedString) string { return "***" }
				p := new(flags.RedactedString)
				v := flags.NewValueWithRedact("", p, parse, redact)
				fs.VarPF(v, flagName, "", "")
			},
			wantValue: "",
			wantError: fmt.Errorf("trying to get %T value of flag of type %s", *new(flags.CustomString), *new(flags.RedactedString)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			tc.giveSetup(fs)
			s, err := flags.GetCustomString(fs, flagName)
			if tc.wantError != nil {
				assert.Error(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, s, tc.wantValue)
		})
	}
}

type ValueTestSuite struct {
	suite.Suite
}

func (suite *ValueTestSuite) TestNewValue() {
	t := suite.T()
	flagName := "myflag"
	expectedDefaultValue := 100
	multiplier := 500

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	parse := func(val string) (int, error) {
		if iVal, err := strconv.Atoi(val); err != nil {
			return 0, err
		} else {
			return iVal * multiplier, nil
		}
	}
	p := new(int)
	v := flags.NewValue(expectedDefaultValue, p, parse)
	fs.VarP(v, flagName, "", "")

	actualValue, err := fs.GetInt(flagName)
	require.NoError(t, err)
	assert.Equal(t, expectedDefaultValue, actualValue)
	fs.Set(flagName, "2")
	actualValue, err = fs.GetInt(flagName)
	require.NoError(t, err)
	assert.Equal(t, multiplier*2, actualValue)
}

func (suite *ValueTestSuite) TestNewValueWithRedact() {
	t := suite.T()
	flagName := "myflag"
	expectedDefaultValue := "default"
	expectedRedactedValue := "!!!!"
	unredactedValue := "foobar"

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	parse := func(val string) (string, error) {
		return fmt.Sprintf("%v::%v", val, val), nil
	}
	redact := func(string) string { return expectedRedactedValue }
	p := new(string)
	v := flags.NewValueWithRedact(expectedDefaultValue, p, parse, redact)
	fs.VarPF(v, flagName, "", "")

	actualValue, err := fs.GetString(flagName)
	require.NoError(t, err)
	assert.Equal(t, expectedRedactedValue, actualValue)

	actualFlag := fs.Lookup(flagName)
	require.NotNil(t, actualFlag)
	assert.Equal(t, expectedRedactedValue, actualFlag.Value.String())
	rs := actualFlag.Value.(*flags.Value[string])
	require.NotNil(t, rs)
	assert.Equal(t, expectedDefaultValue, rs.Unredacted().String())

	fs.Set(flagName, unredactedValue)
	assert.Equal(t, expectedRedactedValue, actualFlag.Value.String())
	require.NotNil(t, rs)
	assert.Equal(t, fmt.Sprintf("%v::%v", unredactedValue, unredactedValue), rs.Unredacted().String())
}

func (suite *ValueTestSuite) TestValueType() {
	testCases := []struct {
		testDescription string
		giveValue       pflag.Value
		wantTypeName    string
	}{
		{
			testDescription: "string",
			giveValue:       &flags.Value[string]{},
			wantTypeName:    "string",
		},
		{
			testDescription: "int",
			giveValue:       &flags.Value[int]{},
			wantTypeName:    "int",
		},
		{
			testDescription: "bool",
			giveValue:       &flags.Value[bool]{},
			wantTypeName:    "bool",
		},
		{
			testDescription: "RedactedString",
			giveValue:       &flags.Value[flags.RedactedString]{},
			wantTypeName:    fmt.Sprint(reflect.TypeOf(*new(flags.RedactedString))),
		},
		{
			testDescription: "CustomString",
			giveValue:       &flags.Value[flags.CustomString]{},
			wantTypeName:    fmt.Sprint(reflect.TypeOf(*new(flags.CustomString))),
		},
		{
			testDescription: "StringSlice",
			giveValue:       &flags.Value[flags.StringSlice]{},
			wantTypeName:    fmt.Sprint(reflect.TypeOf(*new(flags.StringSlice))),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			actualTypeName := tc.giveValue.Type()
			assert.Equal(t, tc.wantTypeName, actualTypeName)
		})
	}
}

func TestValueTestSuite(t *testing.T) {
	suite.Run(t, new(ValueTestSuite))
}
