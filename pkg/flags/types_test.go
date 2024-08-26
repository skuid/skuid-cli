package flags_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

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

func (suite *ValueTestSuite) TestType() {
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
			testDescription: "[]string",
			giveValue:       &flags.Value[[]string]{},
			wantTypeName:    "[]string",
		},
		{
			testDescription: "pointer",
			giveValue:       &flags.Value[*string]{},
			wantTypeName:    "string",
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

type SliceValueTestSuite struct {
	suite.Suite
}

func (suite *SliceValueTestSuite) TestNewSliceValue() {
	t := suite.T()
	flagName := "myflag"
	expectedDefaultValue := []int{100}
	multiplier := 500

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	parse := func(val string) (int, error) {
		if iVal, err := strconv.Atoi(val); err != nil {
			return 0, err
		} else {
			return iVal * multiplier, nil
		}
	}
	p := new([]int)
	v := flags.NewSliceValue(expectedDefaultValue, p, parse)
	fs.VarP(v, flagName, "", "")

	actualValue, err := fs.GetIntSlice(flagName)
	require.NoError(t, err)
	assert.Equal(t, expectedDefaultValue, actualValue)
	fs.Set(flagName, "2")
	actualValue, err = fs.GetIntSlice(flagName)
	require.NoError(t, err)
	assert.Equal(t, []int{multiplier * 2}, actualValue)
}

func (suite *SliceValueTestSuite) TestNewSliceValueWithRedact() {
	t := suite.T()
	flagName := "myflag"
	expectedDefaultValue := []string{"default"}
	redactedValue := "!!!!"
	expectedRedactedValue := []string{redactedValue}
	unredactedValue := "foobar"

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	parse := func(val string) (string, error) {
		return fmt.Sprintf("%v::%v", val, val), nil
	}
	redact := func(string) string { return redactedValue }
	p := new([]string)
	v := flags.NewSliceValueWithRedact(expectedDefaultValue, p, parse, redact)
	fs.VarPF(v, flagName, "", "")

	actualValue, err := fs.GetStringSlice(flagName)
	require.NoError(t, err)
	assert.Equal(t, expectedRedactedValue, actualValue)

	actualFlag := fs.Lookup(flagName)
	require.NotNil(t, actualFlag)
	assert.Equal(t, "["+strings.Join(expectedRedactedValue, ", ")+"]", actualFlag.Value.String())
	rs := actualFlag.Value.(*flags.SliceValue[string])
	require.NotNil(t, rs)
	assert.Equal(t, "["+strings.Join(expectedDefaultValue, ", ")+"]", rs.Unredacted().String())

	fs.Set(flagName, unredactedValue)
	assert.Equal(t, "["+strings.Join(expectedRedactedValue, ", ")+"]", actualFlag.Value.String())
	require.NotNil(t, rs)
	assert.Equal(t, fmt.Sprintf("[%v::%v]", unredactedValue, unredactedValue), rs.Unredacted().String())
}

func (suite *SliceValueTestSuite) TestType() {
	testCases := []struct {
		testDescription string
		giveValue       pflag.Value
		wantTypeName    string
	}{
		{
			testDescription: "string",
			giveValue:       &flags.SliceValue[string]{},
			wantTypeName:    "stringSlice",
		},
		{
			testDescription: "int",
			giveValue:       &flags.SliceValue[int]{},
			wantTypeName:    "intSlice",
		},
		{
			testDescription: "bool",
			giveValue:       &flags.SliceValue[bool]{},
			wantTypeName:    "boolSlice",
		},
		{
			testDescription: "[]string",
			giveValue:       &flags.SliceValue[[]string]{},
			wantTypeName:    "[][]string",
		},
		{
			testDescription: "pointer",
			giveValue:       &flags.SliceValue[*string]{},
			wantTypeName:    "[]*string",
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

func TestSliceValueTestSuite(t *testing.T) {
	suite.Run(t, new(SliceValueTestSuite))
}
