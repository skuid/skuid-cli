package cmdutil_test

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ApplyEnvVarsTestCase struct {
	testDescription     string
	giveFlags           []*flags.Flag[string]
	giveArgs            []string
	wantFlags           map[string]string
	wantErrorContains   error
	wantChanged         bool
	wantChangedByEnvVar string
	wantDeprecated      bool
}

type ApplyEnvVarsTestSuite struct {
	suite.Suite
}

func (suite *ApplyEnvVarsTestSuite) TestUpdatesValue() {
	envVars := map[string]testutil.EnvVar[string]{
		"SKUID_FOO":   {EnvValue: "hello", Value: "hello"},
		"LEGACY_FOO":  {EnvValue: "goodbye", Value: "goodbye"},
		"LEGACY_FOO2": {EnvValue: "goodbye2222", Value: "goodbye2222"},
		"SKUID_BAR":   {EnvValue: "", Value: ""},
	}
	testCases := []ApplyEnvVarsTestCase{
		{
			testDescription:     "default env var",
			giveFlags:           []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag"}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"foo": "hello"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_FOO",
		},
		{
			testDescription:     "legacy env var",
			giveFlags:           []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"legacy": "goodbye"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      true,
			wantChangedByEnvVar: "LEGACY_FOO",
		},
		{
			testDescription:     "legacy env var multiple options",
			giveFlags:           []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FAKE1", "LEGACY_FAKE2", "LEGACY_FOO"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"legacy": "goodbye"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      true,
			wantChangedByEnvVar: "LEGACY_FOO",
		},
		{
			testDescription:     "legacy env var first found in specified order",
			giveFlags:           []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FOO2", "LEGACY_FOO"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"legacy": "goodbye2222"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      true,
			wantChangedByEnvVar: "LEGACY_FOO2",
		},
		{
			testDescription:     "default env var when has legacy",
			giveFlags:           []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag", LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"foo": "hello"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_FOO",
		},
		{
			testDescription:     "empty used when not required",
			giveFlags:           []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "somevalue"}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"bar": ""},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_BAR",
		},
		{
			testDescription:     "empty used when required",
			giveFlags:           []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "abcd", Required: true}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"bar": ""},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_BAR",
		},
		{
			testDescription:     "empty used when default empty and legacy non-empty when not required",
			giveFlags:           []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "somevalue", LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"bar": ""},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_BAR",
		},
		{
			testDescription:     "empty used when default empty and legacy non-empty when required",
			giveFlags:           []*flags.Flag[string]{{Name: "bar", Usage: "this is a flag", Default: "somevalue", Required: true, LegacyEnvVars: []string{"LEGACY_FOO"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"bar": ""},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_BAR",
		},
		{
			testDescription:     "non-empty used when required",
			giveFlags:           []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag", Default: "somevalue", Required: true}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"foo": "hello"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "SKUID_FOO",
		},
		{
			testDescription:     "no default legacy is empty used when required",
			giveFlags:           []*flags.Flag[string]{{Name: "legacy", Usage: "this is a flag", Default: "abcd", Required: true, LegacyEnvVars: []string{"SKUID_BAR"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"legacy": ""},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      true,
			wantChangedByEnvVar: "SKUID_BAR",
		},
	}

	runApplyEnvVarsTests(&suite.Suite, envVars, testCases)
}

func (suite *ApplyEnvVarsTestSuite) TestDoesNotUpdateValue() {
	envVars := map[string]testutil.EnvVar[string]{
		"SKUID_FOO":  {EnvValue: "hello", Value: "hello"},
		"LEGACY_FOO": {EnvValue: "goodbye", Value: "goodbye"},
		"SKUID_BAR":  {EnvValue: "", Value: ""},
	}
	testCases := []ApplyEnvVarsTestCase{
		{
			testDescription:     "passed on command line",
			giveFlags:           []*flags.Flag[string]{{Name: "foo", Usage: "this is a flag"}},
			giveArgs:            []string{"--foo", "bar"},
			wantFlags:           map[string]string{"foo": "bar"},
			wantErrorContains:   nil,
			wantChanged:         true,
			wantDeprecated:      false,
			wantChangedByEnvVar: "",
		},
		{
			testDescription:     "not in env",
			giveFlags:           []*flags.Flag[string]{{Name: "abc", Usage: "this is a flag", Default: "somevalue", LegacyEnvVars: []string{"LEGACY_ABC"}}},
			giveArgs:            []string{},
			wantFlags:           map[string]string{"abc": "somevalue"},
			wantErrorContains:   nil,
			wantChanged:         false,
			wantDeprecated:      false,
			wantChangedByEnvVar: "",
		},
	}
	runApplyEnvVarsTests(&suite.Suite, envVars, testCases)
}

func (suite *ApplyEnvVarsTestSuite) TestFailsToUpdate() {
	t := suite.T()
	envVars := map[string]testutil.EnvVar[string]{
		"SKUID_FOO": {EnvValue: "hello", Value: "hello"},
	}
	testutil.SetupEnv(t, envVars)
	cmd := &cobra.Command{Use: "mycmd"}
	cmd.RunE = testutil.EmptyCobraRun
	var appliedEnvVars []*cmdutil.AppliedEnvVar
	cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
		var err error
		appliedEnvVars, err = cmdutil.ApplyEnvVars(cmd)
		return err
	}
	f := &flags.Flag[bool]{Name: "foo", Usage: "this is a flag", Default: true}
	cmd.Flags().Bool(f.Name, f.Default, f.Usage)
	cmd.Flags().SetAnnotation(f.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
	cmd.Flags().SetAnnotation(f.Name, cmdutil.LegacyEnvVarsAnnotation, f.LegacyEnvVars)

	err := testutil.ExecuteCommand(cmd, []string{}...)

	assert.ErrorContains(t, err, fmt.Sprintf("unable to use value from environment variable %v for flag %q", "SKUID_FOO", f.Name))
	assert.Equal(t, len(appliedEnvVars), 0)
}

func (suite *ApplyEnvVarsTestSuite) TestFailsToUpdateMultipleErrors() {
	t := suite.T()
	envVars := map[string]testutil.EnvVar[string]{
		"SKUID_FOO": {EnvValue: "hello", Value: "hello"},
		"SKUID_BAR": {EnvValue: "goodbye", Value: "goodbye"},
	}
	testutil.SetupEnv(t, envVars)
	cmd := &cobra.Command{Use: "mycmd"}
	cmd.RunE = testutil.EmptyCobraRun
	var appliedEnvVars []*cmdutil.AppliedEnvVar
	cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
		var err error
		appliedEnvVars, err = cmdutil.ApplyEnvVars(cmd)
		return err
	}
	fb := &flags.Flag[bool]{Name: "foo", Usage: "this is a flag", Default: true}
	cmd.Flags().Bool(fb.Name, fb.Default, fb.Usage)
	cmd.Flags().SetAnnotation(fb.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
	cmd.Flags().SetAnnotation(fb.Name, cmdutil.LegacyEnvVarsAnnotation, fb.LegacyEnvVars)
	fi := &flags.Flag[int]{Name: "bar", Usage: "this is a flag", Default: 500}
	cmd.Flags().Int(fi.Name, fi.Default, fi.Usage)
	cmd.Flags().SetAnnotation(fi.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
	cmd.Flags().SetAnnotation(fb.Name, cmdutil.LegacyEnvVarsAnnotation, fi.LegacyEnvVars)

	err := testutil.ExecuteCommand(cmd, []string{}...)

	assert.ErrorContains(t, err, fb.Name)
	assert.ErrorContains(t, err, "SKUID_FOO")
	assert.ErrorContains(t, err, fi.Name)
	assert.ErrorContains(t, err, "SKUID_BAR")
	u, ok := err.(interface {
		Unwrap() []error
	})
	require.True(t, ok)
	assert.Equal(t, 2, len(u.Unwrap()))
	assert.Equal(t, len(appliedEnvVars), 0)
}

func (suite *ApplyEnvVarsTestSuite) TestFlagNotSkuidCliFlag() {
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
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			testutil.SetupEnv(t, envVars)
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = testutil.EmptyCobraRun
			var appliedEnvVars []*cmdutil.AppliedEnvVar
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				var err error
				appliedEnvVars, err = cmdutil.ApplyEnvVars(cmd)
				return err
			}
			tc.setupFlags(cmd)

			err := testutil.ExecuteCommand(cmd, []string{}...)

			if tc.wantErrorContains != nil {
				assert.ErrorContains(t, err, tc.wantErrorContains.Error())
			} else {
				require.NoError(t, err)
			}
			actualFlag := cmd.Flags().Lookup(tc.wantFlag)
			require.NotNil(t, actualFlag)
			assert.False(t, actualFlag.Changed)
			assert.Equal(t, len(appliedEnvVars), 0)
		})
	}
}

func (suite *ApplyEnvVarsTestSuite) TestStringConversion() {
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
	addFlag := func(fs *pflag.FlagSet, name string, p *string) *flags.Flag[string] {
		flag := &flags.Flag[string]{Name: name, Usage: "this is a flag"}
		fs.StringVar(p, flag.Name, flag.Default, flag.Usage)
		return flag
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true)
}

func (suite *ApplyEnvVarsTestSuite) TestStringSliceConversion() {
	envVarsSuccess := map[string]testutil.EnvVar[[]string]{
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
	envVarsError := map[string]testutil.EnvVar[[]string]{}
	addFlag := func(fs *pflag.FlagSet, name string, p *[]string) *flags.Flag[[]string] {
		flag := &flags.Flag[[]string]{Name: name, Usage: "this is a flag"}
		fs.StringSliceVar(p, flag.Name, flag.Default, flag.Usage)
		return flag
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true)
}

func (suite *ApplyEnvVarsTestSuite) TestBoolConversion() {
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
	addFlag := func(fs *pflag.FlagSet, name string, p *bool) *flags.Flag[bool] {
		flag := &flags.Flag[bool]{Name: name, Usage: "this is a flag"}
		fs.BoolVar(p, flag.Name, flag.Default, flag.Usage)
		return flag
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true)
}

func (suite *ApplyEnvVarsTestSuite) TestIntConversion() {
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
	addFlag := func(fs *pflag.FlagSet, name string, p *int) *flags.Flag[int] {
		flag := &flags.Flag[int]{Name: name, Usage: "this is a flag"}
		fs.IntVar(p, flag.Name, flag.Default, flag.Usage)
		return flag
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true)
}

func (suite *ApplyEnvVarsTestSuite) TestPtrTimeConversion() {
	d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	layout := time.RFC3339Nano
	envVarsSuccess := map[string]testutil.EnvVar[*time.Time]{
		"SKUID_GOOD_FOO_DATE_TIME": {EnvValue: d.Format(layout), Value: &d},
	}
	envVarsError := map[string]testutil.EnvVar[*time.Time]{
		"SKUID_BAD_FOO_EMPTY":  {EnvValue: "", Value: nil},
		"SKUID_BAD_FOO_FORMAT": {EnvValue: "2026_01_01", Value: nil},
	}
	addFlag := func(fs *pflag.FlagSet, name string, p **time.Time) *flags.Flag[*time.Time] {
		flag := &flags.Flag[*time.Time]{Name: name, Usage: "this is a flag"}
		v := flags.NewValue(flag.Default, p, func(val string) (*time.Time, error) {
			t, err := time.ParseInLocation(layout, val, time.Local)
			return &t, err
		})
		fs.Var(v, flag.Name, flag.Usage)
		return flag
	}
	runConversionTest(&suite.Suite, envVarsSuccess, addFlag, false)
	runConversionTest(&suite.Suite, envVarsError, addFlag, true)
}

func TestApplyEnvVarsTestSuite(t *testing.T) {
	suite.Run(t, new(ApplyEnvVarsTestSuite))
}

func runApplyEnvVarsTests(suite *suite.Suite, envVars map[string]testutil.EnvVar[string], testCases []ApplyEnvVarsTestCase) {
	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			testutil.SetupEnv(t, envVars)
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = testutil.EmptyCobraRun
			var appliedEnvVars []*cmdutil.AppliedEnvVar
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				var err error
				appliedEnvVars, err = cmdutil.ApplyEnvVars(cmd)
				return err
			}
			for _, f := range tc.giveFlags {
				cmd.Flags().String(f.Name, f.Default, f.Usage)
				cmd.Flags().SetAnnotation(f.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
				cmd.Flags().SetAnnotation(f.Name, cmdutil.LegacyEnvVarsAnnotation, f.LegacyEnvVars)
				if f.Required {
					cmd.MarkFlagRequired(f.Name)
				}
			}

			err := testutil.ExecuteCommand(cmd, tc.giveArgs...)

			if tc.wantErrorContains != nil {
				assert.ErrorContains(t, err, tc.wantErrorContains.Error())
			} else {
				require.NoError(t, err)
			}

			changedCount := 0
			for k, v := range tc.wantFlags {
				flag := cmd.Flags().Lookup(k)
				require.NotNil(t, flag)
				val, err := cmd.Flags().GetString(k)
				require.NoError(t, err)
				assert.Equal(t, v, val)
				assert.Equal(t, flag.Changed, tc.wantChanged)
				i := slices.IndexFunc(appliedEnvVars, func(a *cmdutil.AppliedEnvVar) bool {
					return a.Flag.Name == k
				})
				if tc.wantChangedByEnvVar != "" {
					changedCount++
					require.GreaterOrEqual(t, i, 0)
					aev := appliedEnvVars[i]
					assert.Equal(t, tc.wantChangedByEnvVar, aev.Name)
					assert.Equal(t, tc.wantDeprecated, aev.Deprecated)
				} else {
					require.GreaterOrEqual(t, i, -1)
				}
			}
			assert.Equal(t, len(appliedEnvVars), changedCount)
		})
	}
}

func runConversionTest[T flags.FlagType](suite *suite.Suite, envVars map[string]testutil.EnvVar[T], addFlag func(*pflag.FlagSet, string, *T) *flags.Flag[T], wantError bool) {
	for envKey, envValue := range envVars {
		flagName := strings.TrimPrefix(envKey, "SKUID_")
		flagName = strings.ReplaceAll(flagName, "_", "-")
		flagName = strings.ToLower(flagName)
		testDescription := fmt.Sprintf("convert %v %v", reflect.TypeOf(envValue.Value), envKey)
		suite.Run(testDescription, func() {
			t := suite.T()
			testutil.SetupEnv(t, envVars)
			cmd := &cobra.Command{Use: "mycmd"}
			cmd.RunE = testutil.EmptyCobraRun
			var appliedEnvVars []*cmdutil.AppliedEnvVar
			cmd.PersistentPreRunE = func(c *cobra.Command, a []string) error {
				var err error
				appliedEnvVars, err = cmdutil.ApplyEnvVars(cmd)
				return err
			}
			p := new(T)
			flag := addFlag(cmd.Flags(), flagName, p)
			cmd.Flags().SetAnnotation(flag.Name, cmdutil.FlagSetBySkuidCliAnnotation, []string{"true"})
			cmd.Flags().SetAnnotation(flag.Name, cmdutil.LegacyEnvVarsAnnotation, flag.LegacyEnvVars)
			err := testutil.ExecuteCommand(cmd, []string{}...)
			if wantError {
				assert.ErrorContains(t, err, envKey)
				assert.Equal(t, len(appliedEnvVars), 0)
			} else {
				require.NoError(t, err)
				require.NoError(t, err)
				assert.Equal(t, envValue.Value, *p)
				require.Equal(t, len(appliedEnvVars), 1)
				aev := appliedEnvVars[0]
				assert.Equal(t, envKey, aev.Name)
				assert.Equal(t, flagName, aev.Flag.Name)
				assert.False(t, false)
			}
		})
	}
}
