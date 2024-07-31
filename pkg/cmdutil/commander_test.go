package cmdutil_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/logging/mocks"
	"github.com/skuid/skuid-cli/pkg/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AddFlagsTestCase[T flags.FlagType] struct {
	testDescription string
	giveFlag        *flags.Flag[T]
}

type ApplyEnvVarsTestCase struct {
	testDescription   string
	giveFlags         []*flags.Flag[string]
	giveArgs          []string
	wantFlags         map[string]string
	wantErrorContains error
	wantChanged       bool
}

type SetupLoggingTestCase struct {
	testDescription   string
	giveSetup         func(*testing.T) (*cobra.Command, logging.LogInformer)
	giveArgs          []string
	wantErrorContains error
}

type CommanderAddFlagsTestSuite struct {
	suite.Suite
}

func (suite *CommanderAddFlagsTestSuite) TestAddStringFlags() {
	testCases := createAddCmdFlagsTests("stringdefault")
	getValue := func(fs *pflag.FlagSet, fn string) (string, error) {
		return fs.GetString(fn)
	}
	createFlags := func(f *flags.Flag[string]) *cmdutil.CommandFlags {
		return &cmdutil.CommandFlags{
			String: []*flags.Flag[string]{f},
		}
	}
	runAddFlagsTests(&suite.Suite, testCases, createFlags, getValue)
}

func (suite *CommanderAddFlagsTestSuite) TestAddIntFlags() {
	testCases := createAddCmdFlagsTests(5)
	getValue := func(fs *pflag.FlagSet, fn string) (int, error) {
		return fs.GetInt(fn)
	}
	createFlags := func(f *flags.Flag[int]) *cmdutil.CommandFlags {
		return &cmdutil.CommandFlags{
			Int: []*flags.Flag[int]{f},
		}
	}
	runAddFlagsTests(&suite.Suite, testCases, createFlags, getValue)
}

func (suite *CommanderAddFlagsTestSuite) TestAddBoolFlags() {
	testCases := createAddCmdFlagsTests(true)
	getValue := func(fs *pflag.FlagSet, fn string) (bool, error) {
		return fs.GetBool(fn)
	}
	createFlags := func(f *flags.Flag[bool]) *cmdutil.CommandFlags {
		return &cmdutil.CommandFlags{
			Bool: []*flags.Flag[bool]{f},
		}
	}
	runAddFlagsTests(&suite.Suite, testCases, createFlags, getValue)
}

func (suite *CommanderAddFlagsTestSuite) TestAddStringSliceFlags() {
	testCases := createAddCmdFlagsTests(flags.StringSlice{"strslice1", "strslice2"})
	getValue := func(fs *pflag.FlagSet, fn string) (flags.StringSlice, error) {
		// unable to use fs.GetStringSlice(fn) because if the value is nil, it will return
		// an empty slice which won't be the same as the Default.  Using below, we get the
		// raw value of the slice to ensure we can check for equality in all scenarios.
		f := fs.Lookup(fn)
		if f == nil {
			return nil, fmt.Errorf("unable to find flag %q", fn)
		}
		val, ok := f.Value.(pflag.SliceValue)
		if !ok {
			return nil, fmt.Errorf("could not assert SliceValue for flag %q", fn)
		}
		return val.GetSlice(), nil
	}
	createFlags := func(f *flags.Flag[flags.StringSlice]) *cmdutil.CommandFlags {
		return &cmdutil.CommandFlags{
			StringSlice: []*flags.Flag[flags.StringSlice]{f},
		}
	}
	runAddFlagsTests(&suite.Suite, testCases, createFlags, getValue)
}

func (suite *CommanderAddFlagsTestSuite) TestAddsAllTypesWithMultipleFlagsEach() {
	t := suite.T()
	flags := &cmdutil.CommandFlags{
		String:      []*flags.Flag[string]{{Name: "sflag1", Usage: "this is the usage"}, {Name: "sflag2", Usage: "this is the usage"}},
		Int:         []*flags.Flag[int]{{Name: "iflag1", Usage: "this is the usage"}, {Name: "iflag2", Usage: "this is the usage"}},
		Bool:        []*flags.Flag[bool]{{Name: "bflag1", Usage: "this is the usage"}, {Name: "bflag2", Usage: "this is the usage"}},
		StringSlice: []*flags.Flag[flags.StringSlice]{{Name: "ssflag1", Usage: "this is the usage"}, {Name: "ssflag2", Usage: "this is the usage"}},
	}
	flagMgr := &cmdutil.FlagManager{}
	cd := &cmdutil.Commander{
		FlagManager: flagMgr,
	}
	cmd := &cobra.Command{Use: "mycmd"}
	cmd.RunE = emptyCobraRun

	cd.AddFlags(cmd, flags)
	err := executeCommand(cmd, []string{}...)

	require.NoError(t, err)
	assertFlags(t, flagMgr, cmd, flags.String)
	assertFlags(t, flagMgr, cmd, flags.Int)
	assertFlags(t, flagMgr, cmd, flags.Bool)
	assertFlags(t, flagMgr, cmd, flags.StringSlice)
}

func (suite *CommanderAddFlagsTestSuite) TestInvalidFlagName() {
	testCases := []struct {
		testDescription string
		giveFlags       *cmdutil.CommandFlags
	}{
		{
			testDescription: "no name",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{}},
			},
		},
		{
			testDescription: "empty name",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: ""}},
			},
		},
		{
			testDescription: "name contains space",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "foo bar"}},
			},
		},
		{
			testDescription: "contains special char",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "foo$bar"}},
			},
		},
		{
			testDescription: "contains upper case",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "fooBar"}},
			},
		},
		{
			testDescription: "min length",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "fo"}},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			cd := &cmdutil.Commander{}
			cmd := &cobra.Command{Use: "mycmd"}
			expectedError := fmt.Sprintf("flag name %q is invalid", tc.giveFlags.String[0].Name)

			assert.PanicsWithError(t, expectedError, func() {
				cd.AddFlags(cmd, tc.giveFlags)
			})
		})
	}
}

func (suite *CommanderAddFlagsTestSuite) TestInvalidFlagUsage() {
	testCases := []struct {
		testDescription string
		giveFlags       *cmdutil.CommandFlags
	}{
		{
			testDescription: "no usage",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "foobar", Usage: ""}},
			},
		},
		{
			testDescription: "empty usage",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "foobar", Usage: ""}},
			},
		},
		{
			testDescription: "min length",
			giveFlags: &cmdutil.CommandFlags{
				String: []*flags.Flag[string]{{Name: "foobar", Usage: "012345678"}},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			cd := &cmdutil.Commander{}
			cmd := &cobra.Command{Use: "mycmd"}
			expectedError := fmt.Sprintf("flag usage %q is invalid for flag name %q", tc.giveFlags.String[0].Usage, tc.giveFlags.String[0].Name)

			assert.PanicsWithError(t, expectedError, func() {
				cd.AddFlags(cmd, tc.giveFlags)
			})
		})
	}
}

func (suite *CommanderAddFlagsTestSuite) TestFlagParse() {
	flagName := "myflag"
	flagArg := fmt.Sprintf("--%v", flagName)
	testCases := []struct {
		testDescription string
		giveFlags       func() *cmdutil.CommandFlags
		giveArgs        []string
		giveGetValue    func(fs *pflag.FlagSet, fn string) (any, error)
		wantType        string
		wantPanic       bool
		wantValue       any
	}{
		{
			testDescription: "string supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					String: []*flags.Flag[string]{{Name: flagName, Usage: "this is the usage", Parse: func(string) (string, error) { return "abcd", nil }}},
				}
			},
			giveArgs: []string{flagArg, "foobar"},
			giveGetValue: func(fs *pflag.FlagSet, fn string) (any, error) {
				return fs.GetString(fn)
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[string]{}),
			wantPanic: false,
			wantValue: "abcd",
		},
		{
			testDescription: "bool not supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					Bool: []*flags.Flag[bool]{{Name: flagName, Usage: "this is the usage", Parse: func(string) (bool, error) { return false, nil }}},
				}
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[bool]{}),
			wantPanic: true,
		},
		{
			testDescription: "int not supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					Int: []*flags.Flag[int]{{Name: flagName, Usage: "this is the usage", Parse: func(string) (int, error) { return 5, nil }}},
				}
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[int]{}),
			wantPanic: true,
		},
		{
			testDescription: "StringSlice not supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					StringSlice: []*flags.Flag[flags.StringSlice]{{Name: flagName, Usage: "this is the usage", Parse: func(string) (flags.StringSlice, error) { return flags.StringSlice{}, nil }}},
				}
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[flags.StringSlice]{}),
			wantPanic: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			cd := &cmdutil.Commander{
				FlagManager: cmdutil.NewFlagManager(),
			}
			cmd := &cobra.Command{Use: "mycmd"}
			if tc.wantPanic {
				expectedError := fmt.Sprintf("flag type %s does not support Parse for flag name %q", tc.wantType, flagName)
				assert.PanicsWithError(t, expectedError, func() {
					cd.AddFlags(cmd, tc.giveFlags())
				})
			} else {
				require.NotPanics(t, func() {
					cd.AddFlags(cmd, tc.giveFlags())
				})
				err := executeCommand(cmd, tc.giveArgs...)
				require.NoError(t, err)
				actualValue, err := tc.giveGetValue(cmd.Flags(), flagName)
				require.NoError(t, err)
				assert.Equal(t, tc.wantValue, actualValue)
			}
		})
	}
}

func (suite *CommanderAddFlagsTestSuite) TestFlagRedact() {
	flagName := "myflag"
	flagArg := fmt.Sprintf("--%v", flagName)
	testCases := []struct {
		testDescription        string
		giveFlags              func() *cmdutil.CommandFlags
		giveArgs               []string
		giveGetUnredactedValue func(fv pflag.Value) (any, bool)
		wantType               string
		wantPanic              bool
		wantRedactedValue      string
		wantUnredactedValue    any
	}{
		{
			testDescription: "string supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					String: []*flags.Flag[string]{{Name: flagName, Usage: "this is the usage", Redact: func(string) string { return "!!!!" }}},
				}
			},
			giveArgs: []string{flagArg, "foobar"},
			giveGetUnredactedValue: func(fv pflag.Value) (any, bool) {
				if f, ok := fv.(*flags.Value[string]); !ok {
					return nil, false
				} else {
					return f.Unredacted().String(), true
				}
			},
			wantType:            fmt.Sprintf("%T", &flags.Flag[string]{}),
			wantPanic:           false,
			wantRedactedValue:   "!!!!",
			wantUnredactedValue: "foobar",
		},
		{
			testDescription: "bool not supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					Bool: []*flags.Flag[bool]{{Name: flagName, Usage: "this is the usage", Redact: func(bool) string { return "!!!!" }}},
				}
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[bool]{}),
			wantPanic: true,
		},
		{
			testDescription: "int not supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					Int: []*flags.Flag[int]{{Name: flagName, Usage: "this is the usage", Redact: func(int) string { return "!!!!" }}},
				}
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[int]{}),
			wantPanic: true,
		},
		{
			testDescription: "StringSlice not supported",
			giveFlags: func() *cmdutil.CommandFlags {
				return &cmdutil.CommandFlags{
					StringSlice: []*flags.Flag[flags.StringSlice]{{Name: flagName, Usage: "this is the usage", Redact: func(flags.StringSlice) string { return "!!!!" }}},
				}
			},
			wantType:  fmt.Sprintf("%T", &flags.Flag[flags.StringSlice]{}),
			wantPanic: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			cd := &cmdutil.Commander{
				FlagManager: cmdutil.NewFlagManager(),
			}
			cmd := &cobra.Command{Use: "mycmd"}
			if tc.wantPanic {
				expectedError := fmt.Sprintf("flag type %s does not support Redact for flag name %q", tc.wantType, flagName)
				assert.PanicsWithError(t, expectedError, func() {
					cd.AddFlags(cmd, tc.giveFlags())
				})
			} else {
				require.NotPanics(t, func() {
					cd.AddFlags(cmd, tc.giveFlags())
				})
				err := executeCommand(cmd, tc.giveArgs...)
				require.NoError(t, err)
				actualFlag := cmd.Flags().Lookup(flagName)
				require.NotNil(t, actualFlag)
				assert.Equal(t, tc.wantRedactedValue, actualFlag.Value.String())
				actualUnredactedValue, ok := tc.giveGetUnredactedValue(actualFlag.Value)
				require.True(t, ok)
				assert.Equal(t, tc.wantUnredactedValue, actualUnredactedValue)
			}
		})
	}
}

func TestCommanderAddFlagsTestSuite(t *testing.T) {
	suite.Run(t, new(CommanderAddFlagsTestSuite))
}

type CommanderMarkFlagsMutuallyExclusiveTestSuite struct {
	suite.Suite
}

func (suite *CommanderMarkFlagsMutuallyExclusiveTestSuite) TestFlagsMarked() {
	createCommand := func() *cobra.Command {
		cmd := &cobra.Command{}
		flags := cmd.Flags()
		flags.String("myflags1", "", "this is my usage")
		flags.String("myflags2", "", "this is my usage")
		flags.Bool("myflagb", false, "this is my usage")
		return cmd
	}
	testCases := []struct {
		testDescription       string
		giveMutuallyExclusive [][]string
		giveArgs              []string
		wantError             []string
	}{
		{
			testDescription:       "args violate rule",
			giveMutuallyExclusive: [][]string{{"myflags1", "myflagb"}},
			giveArgs:              []string{"--myflags1=foo", "--myflagb"},
			wantError:             []string{"myflags1", "myflagb"},
		},
		{
			testDescription:       "args do not violate rule",
			giveMutuallyExclusive: [][]string{{"myflags1", "myflagb"}},
			giveArgs:              []string{"--myflags2=foo", "--myflagb"},
			wantError:             nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			cd := &cmdutil.Commander{}
			cmd := createCommand()
			cmd.RunE = emptyCobraRun

			cd.MarkFlagsMutuallyExclusive(cmd, tc.giveMutuallyExclusive)
			err := executeCommand(cmd, tc.giveArgs...)

			if tc.wantError != nil {
				for _, me := range tc.wantError {
					assert.ErrorContains(t, err, me)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCommanderMarkFlagsMutuallyExclusiveTestSuite(t *testing.T) {
	suite.Run(t, new(CommanderMarkFlagsMutuallyExclusiveTestSuite))
}

type CommanderApplyEnvVarsTestSuite struct {
	suite.Suite
}

func (suite *CommanderApplyEnvVarsTestSuite) TestUpdatesValue() {
	envVars := map[string]testutil.EnvVar[string]{
		"SKUID_FOO":   {EnvValue: "hello", Value: "hello"},
		"LEGACY_FOO":  {EnvValue: "goodbye", Value: "goodbye"},
		"LEGACY_FOO2": {EnvValue: "goodbye2222", Value: "goodbye2222"},
		"SKUID_BAR":   {EnvValue: "", Value: ""},
	}
	testCases := []ApplyEnvVarsTestCase{
		{
			testDescription:   "default env var",
			giveFlags:         []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag"}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"foo": "hello"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "legacy env var",
			giveFlags:         []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"legacy": "goodbye"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "legacy env var multiple options",
			giveFlags:         []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FAKE1", "LEGACY_FAKE2", "LEGACY_FOO"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"legacy": "goodbye"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "legacy env var first found in specified order",
			giveFlags:         []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FOO2", "LEGACY_FOO"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"legacy": "goodbye2222"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "default env var when has legacy",
			giveFlags:         []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"foo": "hello"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "empty used when not required",
			giveFlags:         []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "somevalue"}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"bar": ""},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "empty used when default empty and legacy non-blank when not required",
			giveFlags:         []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "somevalue", LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"bar": ""},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "non-empty used when required",
			giveFlags:         []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag", Default: "somevalue", Required: true}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"foo": "hello"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
	}

	runApplyEnvVarsTests(&suite.Suite, envVars, testCases)
}

func (suite *CommanderApplyEnvVarsTestSuite) TestDoesNotUpdateValue() {
	envVars := map[string]testutil.EnvVar[string]{
		"SKUID_FOO":  {EnvValue: "hello", Value: "hello"},
		"LEGACY_FOO": {EnvValue: "goodbye", Value: "goodbye"},
		"SKUID_BAR":  {EnvValue: "", Value: ""},
	}
	testCases := []ApplyEnvVarsTestCase{
		{
			testDescription:   "passed on command line",
			giveFlags:         []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag"}},
			giveArgs:          []string{"--foo", "bar"},
			wantFlags:         map[string]string{"foo": "bar"},
			wantErrorContains: nil,
			wantChanged:       true,
		},
		{
			testDescription:   "flag required and env empty value",
			giveFlags:         []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "abcd", Required: true}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"bar": "abcd"},
			wantErrorContains: errors.New("bar"),
			wantChanged:       false,
		},
		{
			testDescription:   "flag required no default env var legacy is empty",
			giveFlags:         []*flags.Flag[string]{{Name: "nodefault", Usage: "this is a flag", Default: "abcd", Required: true, LegacyEnvVars: []string{"SKUID_BAR"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"nodefault": "abcd"},
			wantErrorContains: errors.New("nodefault"),
			wantChanged:       false,
		},
		{
			testDescription:   "flag required and env empty value when legacy has non-empty",
			giveFlags:         []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "somevalue", Required: true, LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"bar": "somevalue"},
			wantErrorContains: errors.New("bar"),
			wantChanged:       false,
		},
		{
			testDescription:   "not in env",
			giveFlags:         []*flags.Flag[string]{{Name: "abc", Usage: "this is a flag", Default: "somevalue", LegacyEnvVars: []string{"LEGACY_ABC"}}},
			giveArgs:          []string{},
			wantFlags:         map[string]string{"abc": "somevalue"},
			wantErrorContains: nil,
			wantChanged:       false,
		},
	}
	runApplyEnvVarsTests(&suite.Suite, envVars, testCases)
}

func (suite *CommanderApplyEnvVarsTestSuite) TestFlagNotTracked() {
	envVars := map[string]testutil.EnvVar[bool]{
		"SKUID_HELP":   {EnvValue: "true", Value: true},
		"HELP":         {EnvValue: "true", Value: true},
		"SKUID_FOOBAR": {EnvValue: "true", Value: true},
		"FOOBAR":       {EnvValue: "true", Value: true},
	}
	testCases := []struct {
		testDescription   string
		setupFlags        func(*cobra.Command)
		wantFlag          string
		wantErrorContains error
	}{
		{
			testDescription: "cobra automatic flag",
			setupFlags: func(cmd *cobra.Command) {
				// don't need to setup because it will automatically be added
			},
			wantFlag:          "help",
			wantErrorContains: nil,
		},
		{
			testDescription: "unknown flag",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool("foobar", false, "this is the usage")
			},
			wantFlag:          "foobar",
			wantErrorContains: nil,
		},
		{
			testDescription: "skuid cli flag not tracked",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool("skuidcliflag", false, "this is the usage")
				cmd.Flags().SetAnnotation("skuidcliflag", cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
			},
			wantFlag:          "skuidcliflag",
			wantErrorContains: errors.New("skuidcliflag"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			testutil.SetupEnv(t, envVars)
			cd := &cmdutil.Commander{
				FlagManager: &cmdutil.FlagManager{},
			}
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = emptyCobraRun
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				return cd.ApplyEnvVars(cmd)
			}
			tc.setupFlags(cmd)

			err := executeCommand(cmd, []string{}...)

			if tc.wantErrorContains != nil {
				assert.ErrorContains(t, err, tc.wantErrorContains.Error())
			} else {
				require.NoError(t, err)
			}
			actualFlag := cmd.Flags().Lookup(tc.wantFlag)
			require.NotNil(t, actualFlag)
			assert.False(t, actualFlag.Changed)
		})
	}
}

func (suite *CommanderApplyEnvVarsTestSuite) TestStringConversion() {
	envVarsSuccess := map[string]testutil.EnvVar[string]{
		"SKUID_GOOD_FOO_EMPTY":              {EnvValue: "", Value: ""},
		"SKUID_GOOD_FOO_NOT_EMPTY":          {EnvValue: "foobar", Value: "foobar"},
		"SKUID_GOOD_FOO_SPACES":             {EnvValue: "fo o ba r", Value: "fo o ba r"},
		"SKUID_GOOD_FOO_TABS":               {EnvValue: "fo	oba	r", Value: "fo	oba	r"},
		"SKUID_GOOD_FOO_NEWLINE":            {EnvValue: "fooba\nr", Value: "fooba\nr"},
		"SKUID_GOOD_FOO_SPECIAL_CHARACTERS": {EnvValue: "f$o@o\\%s*^()ar", Value: "f$o@o\\%s*^()ar"},
		"SKUID_GOOD_FOO_NUMBER":             {EnvValue: "12345", Value: "12345"},
		"SKUID_GOOD_FOO_ZERO":               {EnvValue: "0", Value: "0"},
		"SKUID_GOOD_FOO_TRUE":               {EnvValue: "true", Value: "true"},
	}
	envVarsError := map[string]testutil.EnvVar[string]{}
	addFlag := func(fs *pflag.FlagSet, name string) *flags.Flag[string] {
		flag := &flags.Flag[string]{Name: name, Usage: "this is a flag"}
		fs.String(flag.Name, flag.Default, flag.Usage)
		return flag
	}
	getValue := func(fs *pflag.FlagSet, fn string) (string, error) {
		return fs.GetString(fn)
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false, getValue)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true, getValue)
}

func (suite *CommanderApplyEnvVarsTestSuite) TestStringSliceConversion() {
	envVarsSuccess := map[string]testutil.EnvVar[flags.StringSlice]{
		"SKUID_GOOD_FOO_EMPTY":               {EnvValue: "", Value: []string{}},
		"SKUID_GOOD_FOO_SINGLE_VALUE":        {EnvValue: "foobar", Value: []string{"foobar"}},
		"SKUID_GOOD_FOO_TRAILING_COMMA":      {EnvValue: "foobar,", Value: []string{"foobar", ""}},
		"SKUID_GOOD_FOO_TRAILING_SPACES":     {EnvValue: "foobar,   ", Value: []string{"foobar", "   "}},
		"SKUID_GOOD_FOO_MULTIPLE_VALUE":      {EnvValue: "foobar,hello", Value: []string{"foobar", "hello"}},
		"SKUID_GOOD_FOO_SPACE_IN_BETWEEN":    {EnvValue: "foobar, hello", Value: []string{"foobar", " hello"}},
		"SKUID_GOOD_FOO_SPACE_VALUE":         {EnvValue: "foobar,   ,hello", Value: []string{"foobar", "   ", "hello"}},
		"SKUID_GOOD_FOO_WHITESPACE_VALUE":    {EnvValue: "foobar,	,hello", Value: []string{"foobar", "	", "hello"}},
		"SKUID_GOOD_FOO_WHITESPACE_IN_VALUE": {EnvValue: "foo	bar, a	b	,he	llo	", Value: []string{"foo	bar", " a	b	", "he	llo	"}},
		"SKUID_GOOD_FOO_SPACE_AROUND":        {EnvValue: "  fo ob ar  ,  h e l lo   ", Value: []string{"  fo ob ar  ", "  h e l lo   "}},
	}
	envVarsError := map[string]testutil.EnvVar[flags.StringSlice]{}
	addFlag := func(fs *pflag.FlagSet, name string) *flags.Flag[flags.StringSlice] {
		flag := &flags.Flag[flags.StringSlice]{Name: name, Usage: "this is a flag"}
		fs.StringSlice(flag.Name, flag.Default, flag.Usage)
		return flag
	}
	getValue := func(fs *pflag.FlagSet, fn string) (flags.StringSlice, error) {
		return fs.GetStringSlice(fn)
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false, getValue)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true, getValue)
}

func (suite *CommanderApplyEnvVarsTestSuite) TestBoolConversion() {
	envVarsSuccess := map[string]testutil.EnvVar[bool]{
		"SKUID_GOOD_FOO_TRUE_LOWER":  {EnvValue: "true", Value: true},
		"SKUID_GOOD_FOO_TRUE_UPPER":  {EnvValue: "TRUE", Value: true},
		"SKUID_GOOD_FOO_TRUE_CAMEL":  {EnvValue: "True", Value: true},
		"SKUID_GOOD_FOO_T_UPPER":     {EnvValue: "T", Value: true},
		"SKUID_GOOD_FOO_T_LOWER":     {EnvValue: "t", Value: true},
		"SKUID_GOOD_FOO_ONE":         {EnvValue: "1", Value: true},
		"SKUID_GOOD_FOO_FALSE_LOWER": {EnvValue: "false", Value: false},
		"SKUID_GOOD_FOO_FALSE_UPPER": {EnvValue: "FALSE", Value: false},
		"SKUID_GOOD_FOO_FALSE_CAMEL": {EnvValue: "False", Value: false},
		"SKUID_GOOD_FOO_F_UPPER":     {EnvValue: "F", Value: false},
		"SKUID_GOOD_FOO_F_LOWER":     {EnvValue: "f", Value: false},
		"SKUID_GOOD_FOO_ZERO":        {EnvValue: "0", Value: false},
	}
	envVarsError := map[string]testutil.EnvVar[bool]{
		"SKUID_BAD_FOO_EMPTY":           {EnvValue: "", Value: false},
		"SKUID_BAD_FOO_MIXED_FALSE":     {EnvValue: "fALse", Value: false},
		"SKUID_BAD_FOO_MIXED_TRUE":      {EnvValue: "tRUe", Value: false},
		"SKUID_BAD_FOO_NON_ZERO_OR_ONE": {EnvValue: "5", Value: false},
		"SKUID_BAD_FOO_NON_NUMERIC":     {EnvValue: "asdfasdf", Value: false},
	}
	addFlag := func(fs *pflag.FlagSet, name string) *flags.Flag[bool] {
		flag := &flags.Flag[bool]{Name: name, Usage: "this is a flag"}
		fs.Bool(flag.Name, flag.Default, flag.Usage)
		return flag
	}
	getValue := func(fs *pflag.FlagSet, fn string) (bool, error) {
		return fs.GetBool(fn)
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false, getValue)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true, getValue)
}

func (suite *CommanderApplyEnvVarsTestSuite) TestIntConversion() {
	envVarsSuccess := map[string]testutil.EnvVar[int]{
		"SKUID_GOOD_FOO_ZERO":               {EnvValue: "0", Value: 0},
		"SKUID_GOOD_FOO_ZERO_POS":           {EnvValue: "+0", Value: 0},
		"SKUID_GOOD_FOO_ZERO_NEG":           {EnvValue: "-0", Value: 0},
		"SKUID_GOOD_FOO_ONE":                {EnvValue: "1", Value: 1},
		"SKUID_GOOD_FOO_ONE_POS":            {EnvValue: "+1", Value: 1},
		"SKUID_GOOD_FOO_ONE_NEG":            {EnvValue: "-1", Value: -1},
		"SKUID_GOOD_FOO_THOUSAND":           {EnvValue: "12345", Value: 12345},
		"SKUID_GOOD_FOO_THOUSAND_POS":       {EnvValue: "+12345", Value: 12345},
		"SKUID_GOOD_FOO_THOUSAND_NEG":       {EnvValue: "-12345", Value: -12345},
		"SKUID_GOOD_FOO_BINARY":             {EnvValue: "0b11000000111001", Value: 12345},
		"SKUID_GOOD_FOO_OCTAL":              {EnvValue: "030071", Value: 12345},
		"SKUID_GOOD_FOO_OCTAL_00":           {EnvValue: "0o30071", Value: 12345},
		"SKUID_GOOD_FOO_HEX":                {EnvValue: "0x3039", Value: 12345},
		"SKUID_GOOD_FOO_UNDERSCORE":         {EnvValue: "1_2_3_4_5", Value: 12345},
		"SKUID_GOOD_FOO_PARTIAL_UNDERSCORE": {EnvValue: "123_45", Value: 12345},
		"SKUID_GOOD_FOO_BIG":                {EnvValue: "987654321", Value: 987654321},
	}
	envVarsError := map[string]testutil.EnvVar[int]{
		"SKUID_BAD_FOO_EMPTY":             {EnvValue: "", Value: 0},
		"SKUID_BAD_FOO_CONTAINS_LETTER":   {EnvValue: "123a45", Value: 0},
		"SKUID_BAD_FOO_STARTS_UNDERSCORE": {EnvValue: "_12345", Value: 0},
		"SKUID_BAD_FOO_ENDS_UNDERSCORE":   {EnvValue: "12345_", Value: 0},
		"SKUID_BAD_FOO_DOUBLE_UNDERSCORE": {EnvValue: "123__45", Value: 0},
		"SKUID_BAD_FOO_INVALID_BASE":      {EnvValue: "0i12345", Value: 0},
	}
	addFlag := func(fs *pflag.FlagSet, name string) *flags.Flag[int] {
		flag := &flags.Flag[int]{Name: name, Usage: "this is a flag"}
		fs.Int(flag.Name, flag.Default, flag.Usage)
		return flag
	}
	getValue := func(fs *pflag.FlagSet, fn string) (int, error) {
		return fs.GetInt(fn)
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false, getValue)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true, getValue)
}

func TestCommanderApplyEnvVarsTestSuite(t *testing.T) {
	suite.Run(t, new(CommanderApplyEnvVarsTestSuite))
}

type CommanderSetupLoggingTestSuite struct {
	suite.Suite
}

func (suite *CommanderSetupLoggingTestSuite) TestConfiguredSuccess() {
	createCommand := func() *cobra.Command {
		cmd := &cobra.Command{}
		fs := cmd.PersistentFlags()
		fs.String(flags.LogDirectory.Name, flags.LogDirectory.Default, "this is my usage")
		fs.Bool(flags.Verbose.Name, flags.Verbose.Default, "this is my usage")
		fs.Bool(flags.Trace.Name, flags.Trace.Default, "this is my usage")
		fs.Bool(flags.FileLogging.Name, flags.FileLogging.Default, "this is my usage")
		fs.Bool(flags.Diagnostic.Name, flags.Diagnostic.Default, "this is my usage")
		return cmd
	}
	testCases := []SetupLoggingTestCase{
		{
			testDescription: "defaults to info",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				return cmd, li
			},
			giveArgs:          []string{},
			wantErrorContains: nil,
		},
		{
			testDescription: "is verbose",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetVerbose().Once()
				return cmd, li
			},
			giveArgs:          []string{"--verbose"},
			wantErrorContains: nil,
		},
		{
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetTrace().Once()
				return cmd, li
			},
			testDescription:   "is trace",
			giveArgs:          []string{"--trace"},
			wantErrorContains: nil,
		},
		{
			testDescription: "is trace",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetDiagnostic().Once()
				return cmd, li
			},
			giveArgs:          []string{"--diagnostic"},
			wantErrorContains: nil,
		},
		{
			testDescription: "verbose & trace",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetTrace().Once()
				return cmd, li
			},
			giveArgs:          []string{"--verbose", "--trace"},
			wantErrorContains: nil,
		},
		{
			testDescription: "all level flags",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetDiagnostic().Once()
				return cmd, li
			},
			giveArgs:          []string{"--verbose", "--trace", "--diagnostic"},
			wantErrorContains: nil,
		},
		{
			testDescription: "file logging default dir",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetFileLogging(flags.LogDirectory.Default).Return(nil).Once()
				return cmd, li
			},
			giveArgs:          []string{"--file-logging"},
			wantErrorContains: nil,
		},
		{
			testDescription: "file logging override dir with non-empty",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetFileLogging("mylogs").Return(nil).Once()
				return cmd, li
			},
			giveArgs:          []string{"--file-logging", "--log-directory=mylogs"},
			wantErrorContains: nil,
		},
		{
			testDescription: "file logging override dir empty",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand()
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetFileLogging("").Return(nil).Once()
				return cmd, li
			},
			giveArgs:          []string{"--file-logging", "--log-directory="},
			wantErrorContains: nil,
		},
	}

	runSetupLoggingTests(&suite.Suite, testCases)
}

func (suite *CommanderSetupLoggingTestSuite) TestConfiguredError() {
	createCommand := func(skip []string) *cobra.Command {
		cmd := &cobra.Command{}
		fs := cmd.PersistentFlags()
		if !slices.Contains(skip, flags.Verbose.Name) {
			fs.Bool(flags.Verbose.Name, flags.Verbose.Default, "this is my usage")
		}
		if !slices.Contains(skip, flags.Trace.Name) {
			fs.Bool(flags.Trace.Name, flags.Trace.Default, "this is my usage")
		}
		if !slices.Contains(skip, flags.Diagnostic.Name) {
			fs.Bool(flags.Diagnostic.Name, flags.Diagnostic.Default, "this is my usage")
		}
		if !slices.Contains(skip, flags.FileLogging.Name) {
			fs.Bool(flags.FileLogging.Name, flags.FileLogging.Default, "this is my usage")
		}
		if !slices.Contains(skip, flags.LogDirectory.Name) {
			fs.String(flags.LogDirectory.Name, flags.LogDirectory.Default, "this is my usage")
		}
		return cmd
	}
	testCases := []SetupLoggingTestCase{
		{
			testDescription: "verbose missing",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand([]string{flags.Verbose.Name})
				li := mocks.NewLogInformer(t)
				return cmd, li
			},
			giveArgs:          []string{},
			wantErrorContains: errors.New(flags.Verbose.Name),
		},
		{
			testDescription: "trace missing",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand([]string{flags.Trace.Name})
				li := mocks.NewLogInformer(t)
				return cmd, li
			},
			giveArgs:          []string{},
			wantErrorContains: errors.New(flags.Trace.Name),
		},
		{
			testDescription: "diagnostic missing",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand([]string{flags.Diagnostic.Name})
				li := mocks.NewLogInformer(t)
				return cmd, li
			},
			giveArgs:          []string{},
			wantErrorContains: errors.New(flags.Diagnostic.Name),
		},
		{
			testDescription: "filelogging missing",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand([]string{flags.FileLogging.Name})
				li := mocks.NewLogInformer(t)
				return cmd, li
			},
			giveArgs:          []string{},
			wantErrorContains: errors.New(flags.FileLogging.Name),
		},
		{
			testDescription: "logdirectory missing",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand([]string{flags.LogDirectory.Name})
				li := mocks.NewLogInformer(t)
				return cmd, li
			},
			giveArgs:          []string{},
			wantErrorContains: errors.New(flags.LogDirectory.Name),
		},
		{
			testDescription: "set file logging fails",
			giveSetup: func(t *testing.T) (*cobra.Command, logging.LogInformer) {
				cmd := createCommand(nil)
				li := mocks.NewLogInformer(t)
				li.EXPECT().SetFileLogging("mylogs").Return(assert.AnError).Once()
				return cmd, li
			},
			giveArgs:          []string{"--file-logging", "--log-directory", "mylogs"},
			wantErrorContains: assert.AnError,
		},
	}

	runSetupLoggingTests(&suite.Suite, testCases)
}

func TestCommanderSetupLoggingTestSuite(t *testing.T) {
	suite.Run(t, new(CommanderSetupLoggingTestSuite))
}

func runConversionTest[T flags.FlagType](suite *suite.Suite, envVars map[string]testutil.EnvVar[T], addFlag func(*pflag.FlagSet, string) *flags.Flag[T], wantError bool, getValue func(*pflag.FlagSet, string) (T, error)) {
	for envKey, envValue := range envVars {
		flagName := strings.TrimPrefix(envKey, "SKUID_")
		flagName = strings.ReplaceAll(flagName, "_", "-")
		flagName = strings.ToLower(flagName)
		testDescription := fmt.Sprintf("convert %v %v", reflect.TypeOf(envValue.Value), envKey)
		suite.Run(testDescription, func() {
			t := suite.T()
			testutil.SetupEnv(t, envVars)
			flagMgr := &cmdutil.FlagManager{}
			cd := &cmdutil.Commander{
				FlagManager: flagMgr,
			}
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = emptyCobraRun
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				return cd.ApplyEnvVars(cmd)
			}
			flag := addFlag(cmd.Flags(), flagName)
			flagMgr.Track(cmd.Flags().Lookup(flagName), flag)
			cmd.Flags().SetAnnotation(flag.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
			err := executeCommand(cmd, []string{}...)
			if wantError {
				assert.ErrorContains(t, err, envKey)
			} else {
				require.NoError(t, err)
				actualValue, err := getValue(cmd.Flags(), flagName)
				require.NoError(t, err)
				assert.Equal(t, envValue.Value, actualValue)
			}
		})
	}
}

func runAddFlagsTests[T flags.FlagType](suite *suite.Suite, testCases []AddFlagsTestCase[T], createFlags func(*flags.Flag[T]) *cmdutil.CommandFlags, getValue func(*pflag.FlagSet, string) (T, error)) {
	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			flagMgr := cmdutil.NewFlagManager()
			cd := &cmdutil.Commander{
				FlagManager: flagMgr,
			}
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = emptyCobraRun
			cmdFlags := createFlags(tc.giveFlag)

			cd.AddFlags(cmd, cmdFlags)
			var flagSet *pflag.FlagSet
			if tc.giveFlag.Global {
				flagSet = cmd.PersistentFlags()
			} else {
				flagSet = cmd.LocalNonPersistentFlags()
			}
			actualFlag := flagSet.Lookup(tc.giveFlag.Name)
			// was it added to correct FlagSet on command
			require.NotNil(t, actualFlag, "Expected lookup to return not nil flag, got nil")

			actualType := actualFlag.Value.Type()
			expectedType := func() string {
				switch et := any(tc.giveFlag.Default).(type) {
				// pflag returns stringSlice and our type is StringSlice
				case flags.StringSlice:
					return "stringSlice"
				default:
					return fmt.Sprint(reflect.TypeOf(et))
				}
			}()
			// correct type?
			assert.Equal(t, expectedType, actualType)

			// usage updated to include environment variables
			assert.Contains(t, actualFlag.Usage, tc.giveFlag.EnvVarName())

			// shorthand
			assert.Equal(t, tc.giveFlag.Shorthand, actualFlag.Shorthand)

			// default value
			actualValue, err := getValue(flagSet, tc.giveFlag.Name)
			require.NoError(t, err)
			assert.Equal(t, tc.giveFlag.Default, actualValue)

			// annotation set
			_, aok := actualFlag.Annotations[cmdutil.FlagSetBySkuidCliAnnotation]
			assert.True(t, aok)

			// make sure its tracked so we can applyenvvars later
			fi, ok := flagMgr.Find(actualFlag)
			assert.True(t, ok)
			assert.NotNil(t, fi)

			// if required, since we aren't passing in any args, it should fail
			err = executeCommand(cmd, []string{}...)
			if tc.giveFlag.Required {
				assert.ErrorContains(t, err, "required")
				assert.ErrorContains(t, err, tc.giveFlag.Name)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createAddCmdFlagsTests[T flags.FlagType](defaultValue T) []AddFlagsTestCase[T] {
	testCases := []AddFlagsTestCase[T]{
		{
			testDescription: "required",
			giveFlag:        &flags.Flag[T]{Name: "testflag", Usage: "this is the usage", Required: true, Shorthand: "v", Default: defaultValue},
		},
		{
			testDescription: "not required",
			giveFlag:        &flags.Flag[T]{Name: "testflag", Usage: "this is the usage", Required: false, Shorthand: "v", Default: defaultValue},
		},
		{
			testDescription: "required global",
			giveFlag:        &flags.Flag[T]{Name: "testflag", Usage: "this is the usage", Required: true, Global: true, Shorthand: "v", Default: defaultValue},
		},
		{
			testDescription: "not required global",
			giveFlag:        &flags.Flag[T]{Name: "testflag", Usage: "this is the usage", Required: false, Global: true, Shorthand: "v", Default: defaultValue},
		},
	}
	return testCases
}

func runApplyEnvVarsTests(suite *suite.Suite, envVars map[string]testutil.EnvVar[string], testCases []ApplyEnvVarsTestCase) {
	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			testutil.SetupEnv(t, envVars)
			flagMgr := &cmdutil.FlagManager{}
			cd := &cmdutil.Commander{
				FlagManager: flagMgr,
			}
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = emptyCobraRun
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				return cd.ApplyEnvVars(cmd)
			}
			for _, f := range tc.giveFlags {
				cmd.Flags().String(f.Name, f.Default, f.Usage)
				flagMgr.Track(cmd.Flags().Lookup(f.Name), f)
				cmd.Flags().SetAnnotation(f.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
				if f.Required {
					cmd.MarkFlagRequired(f.Name)
				}
			}

			err := executeCommand(cmd, tc.giveArgs...)

			if tc.wantErrorContains != nil {
				assert.ErrorContains(t, err, tc.wantErrorContains.Error())
			} else {
				require.NoError(t, err)
			}

			for k, v := range tc.wantFlags {
				flag := cmd.Flags().Lookup(k)
				require.NotNil(t, flag)
				val, err := cmd.Flags().GetString(k)
				require.NoError(t, err)
				assert.Equal(t, v, val)
				assert.Equal(t, flag.Changed, tc.wantChanged)
			}
		})
	}
}

func runSetupLoggingTests(suite *suite.Suite, testCases []SetupLoggingTestCase) {
	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			cd := &cmdutil.Commander{}
			cmd, logCfg := tc.giveSetup(t)
			cmd.RunE = emptyCobraRun
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				return cd.SetupLogging(cmd, logCfg)
			}

			err := executeCommand(cmd, tc.giveArgs...)

			if tc.wantErrorContains != nil {
				require.ErrorContains(t, err, tc.wantErrorContains.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func assertFlags[T flags.FlagType](t *testing.T, flagMgr cmdutil.FlagTrackerFinder, cmd *cobra.Command, flags []*flags.Flag[T]) {
	for _, flag := range flags {
		assertFlag(t, flagMgr, cmd, flag)
	}
}

func assertFlag[T flags.FlagType](t *testing.T, flagMgr cmdutil.FlagTrackerFinder, cmd *cobra.Command, flag *flags.Flag[T]) {
	var actualFlag *pflag.Flag
	if flag.Global {
		actualFlag = cmd.PersistentFlags().Lookup(flag.Name)
	} else {
		actualFlag = cmd.LocalNonPersistentFlags().Lookup(flag.Name)
	}
	require.NotNil(t, actualFlag, "Expected lookup to return not nil flag, got nil")
	assert.Contains(t, actualFlag.Usage, flag.EnvVarName())
	_, aok := actualFlag.Annotations[cmdutil.FlagSetBySkuidCliAnnotation]
	assert.True(t, aok)
	fi, ok := flagMgr.Find(actualFlag)
	assert.True(t, ok)
	assert.NotNil(t, fi)
}
