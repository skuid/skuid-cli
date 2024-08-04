package cmdutil_test

import (
	"testing"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ToCommandTestSuite struct {
	suite.Suite
}

func (suite *ToCommandTestSuite) TestSetsDefaults() {
	testCases := []struct {
		testDescription string
		giveTemplate    *cmdutil.CmdTemplate
		giveRun         cmdutil.CobraRunE
		wantUse         string
		wantShort       string
		wantLong        string
		wantExample     string
	}{
		{
			testDescription: "minimum requird values",
			giveTemplate:    &cmdutil.CmdTemplate{Use: "mycmd", Short: "mycmdshort", Long: "mycmdlong"},
			giveRun:         nil,
			wantUse:         "mycmd",
			wantShort:       "mycmdshort",
			wantLong:        "mycmdlong",
			wantExample:     "",
		},
		{
			testDescription: "everything has a value",
			giveTemplate:    &cmdutil.CmdTemplate{Use: "mycmd", Short: "mycmdshort", Long: "mycmdlong", Example: "mycmdexample"},
			giveRun:         testutil.EmptyCobraRun,
			wantUse:         "mycmd",
			wantShort:       "mycmdshort",
			wantLong:        "mycmdlong",
			wantExample:     "mycmdexample",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			actualCmd := tc.giveTemplate.ToCommand(tc.giveRun)
			assert.NotNil(t, actualCmd)
			assert.True(t, actualCmd.SilenceUsage)
			assert.Equal(t, tc.wantUse, actualCmd.Use)
			assert.Equal(t, tc.wantShort, actualCmd.Short)
			assert.Equal(t, tc.wantLong, actualCmd.Long)
			assert.Equal(t, tc.wantExample, actualCmd.Example)
			assert.Equal(t, tc.giveRun == nil, actualCmd.RunE == nil)
		})
	}
}

func (suite *ToCommandTestSuite) TestInvalidTemplate() {
	testCases := []struct {
		testDescription string
		giveTemplate    *cmdutil.CmdTemplate
		giveRun         cmdutil.CobraRunE
		wantError       string
	}{
		{
			testDescription: "use empty",
			giveTemplate:    &cmdutil.CmdTemplate{Short: "foobar", Long: "foobar"},
			wantError:       "must provide a value for Use",
		},
		{
			testDescription: "short empty",
			giveTemplate:    &cmdutil.CmdTemplate{Use: "foobar", Long: "foobar"},
			wantError:       "must provide a value for Short",
		},

		{
			testDescription: "long empty",
			giveTemplate:    &cmdutil.CmdTemplate{Use: "foobar", Short: "foobar"},
			wantError:       "must provide a value for Long",
		},
		{
			testDescription: "example empty with run",
			giveTemplate:    &cmdutil.CmdTemplate{Use: "mycmd", Short: "foobar", Long: "foobar"},
			giveRun:         testutil.EmptyCobraRun,
			wantError:       "must provide a value for Example",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			assert.PanicsWithError(t, tc.wantError, func() {
				tc.giveTemplate.ToCommand(tc.giveRun)
			})
		})
	}
}

func TestToCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ToCommandTestSuite))
}
