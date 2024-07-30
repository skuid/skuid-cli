package cmdutil_test

import (
	"fmt"
	"testing"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	cmdutil_mocks "github.com/skuid/skuid-cli/pkg/cmdutil/mocks"
	"github.com/skuid/skuid-cli/pkg/flags"
	logging_mocks "github.com/skuid/skuid-cli/pkg/logging/mocks"
	"github.com/skuid/skuid-cli/pkg/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ToCommandTestSuite struct {
	suite.Suite
}

func (suite *ToCommandTestSuite) TestSetsDefaults() {
	testCases := []struct {
		testDescription string
		giveTemplate    *cmdutil.CmdTemplate
		giveRun         cmdutil.CommandFunc
		wantUse         string
		wantShort       string
		wantLong        string
		wantExample     string
		wantFlags       []string
		wantCommands    []string
		wantRunNil      bool
	}{
		{
			testDescription: "empty defaults",
			giveTemplate:    &cmdutil.CmdTemplate{Use: "mycmd"},
			giveRun:         nil,
			wantUse:         "mycmd",
			wantShort:       "",
			wantLong:        "",
			wantExample:     "",
			wantFlags:       nil,
			wantCommands:    nil,
			wantRunNil:      true,
		},
		{
			testDescription: "everything has a value",
			giveTemplate: &cmdutil.CmdTemplate{Use: "mycmd", Short: "mycmdshort", Long: "mycmdlong", Example: "mycmdexample",
				Flags: &cmdutil.CommandFlags{
					String: []*flags.Flag[string]{{Name: "myflags", Usage: "this is my usage s"}},
					Bool:   []*flags.Flag[bool]{{Name: "myflagb", Usage: "this is my usage b"}},
				},
				MutuallyExclusiveFlags: [][]string{{"myflags", "myflagb"}},
				Commands:               []*cobra.Command{{Use: "cmd1"}, {Use: "cmd2"}},
			},
			giveRun:      emptyRun,
			wantUse:      "mycmd",
			wantShort:    "mycmdshort",
			wantLong:     "mycmdlong",
			wantExample:  "mycmdexample",
			wantFlags:    []string{"myflags", "myflagb"},
			wantCommands: []string{"cmd1", "cmd2"},
			wantRunNil:   false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			mockLogConfig := logging_mocks.NewLogInformer(t)
			mockCommander := cmdutil_mocks.NewCommandInformer(t)
			if len(tc.wantFlags) > 0 {
				mockCommander.EXPECT().AddFlags(mock.Anything, mock.Anything).Once()
			}
			if len(tc.giveTemplate.MutuallyExclusiveFlags) > 0 {
				mockCommander.EXPECT().MarkFlagsMutuallyExclusive(mock.Anything, mock.Anything).Once()
			}

			factory := &cmdutil.Factory{
				LogConfig: mockLogConfig,
				Commander: mockCommander,
			}
			actualCmd := tc.giveTemplate.ToCommand(factory, nil, nil, tc.giveRun)
			assert.NotNil(t, actualCmd)
			assert.True(t, actualCmd.SilenceUsage)
			assert.Equal(t, tc.wantUse, actualCmd.Use)
			assert.Equal(t, tc.wantShort, actualCmd.Short)
			assert.Equal(t, tc.wantLong, actualCmd.Long)
			assert.Equal(t, tc.wantExample, actualCmd.Example)

			if len(tc.wantFlags) > 0 {
				mockCommander.AssertCalled(t, "AddFlags", actualCmd, tc.giveTemplate.Flags)
			}

			if len(tc.giveTemplate.MutuallyExclusiveFlags) > 0 {
				mockCommander.AssertCalled(t, "MarkFlagsMutuallyExclusive", actualCmd, tc.giveTemplate.MutuallyExclusiveFlags)
			}

			for _, sub := range actualCmd.Commands() {
				assert.Contains(t, tc.wantCommands, sub.Name())
			}
			assert.Equal(t, len(tc.wantCommands), len(actualCmd.Commands()))
			assert.Equal(t, tc.wantRunNil, actualCmd.RunE == nil)
		})
	}
}

func (suite *ToCommandTestSuite) TestHooksCalled() {
	t := suite.T()

	mockLogConfig := logging_mocks.NewLogInformer(t)
	mockCommander := cmdutil_mocks.NewCommandInformer(t)
	mockPPreRun := cmdutil_mocks.NewCommandFunc(t)
	mockPreRun := cmdutil_mocks.NewCommandFunc(t)
	mockRun := cmdutil_mocks.NewCommandFunc(t)
	template := &cmdutil.CmdTemplate{
		Use: "mycmd",
	}
	factory := &cmdutil.Factory{
		LogConfig: mockLogConfig,
		Commander: mockCommander,
	}
	expectedArgs := []string{"one", "two"}
	expectedCmd := template.ToCommand(factory, mockPPreRun.Execute, mockPreRun.Execute, mockRun.Execute)

	// testify checks equality on pointers based on their referenced values, not the actual pointer value
	factoryMatcher := mock.MatchedBy(testutil.SameMatcher(factory))
	cmdMatcher := mock.MatchedBy(testutil.SameMatcher(expectedCmd))
	pPreRunCall := mockPPreRun.EXPECT().Execute(factoryMatcher, cmdMatcher, expectedArgs).Return(nil).Once()
	preRunCall := mockPreRun.EXPECT().Execute(factoryMatcher, cmdMatcher, expectedArgs).Return(nil).Once().NotBefore(pPreRunCall)
	mockRun.EXPECT().Execute(factoryMatcher, cmdMatcher, expectedArgs).Return(nil).Once().NotBefore(preRunCall)
	err := executeCommand(expectedCmd, expectedArgs...)
	assert.NoError(t, err)
}

func (suite *ToCommandTestSuite) TestHookErrors() {
	testCases := []struct {
		testDescription string
		giveSetup       func(pPreRun *cmdutil_mocks.CommandFunc, preRun *cmdutil_mocks.CommandFunc, run *cmdutil_mocks.CommandFunc)
		wantError       error
		wantRunCalled   int
	}{
		{
			testDescription: "persistent pre run error",
			giveSetup: func(pPreRun *cmdutil_mocks.CommandFunc, preRun *cmdutil_mocks.CommandFunc, run *cmdutil_mocks.CommandFunc) {
				pPreRun.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
				run.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantError:     assert.AnError,
			wantRunCalled: 0,
		},
		{
			testDescription: "pre run error",
			giveSetup: func(pPreRun *cmdutil_mocks.CommandFunc, preRun *cmdutil_mocks.CommandFunc, run *cmdutil_mocks.CommandFunc) {
				pPreRun.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				preRun.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
				run.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantError:     assert.AnError,
			wantRunCalled: 0,
		},
		{
			testDescription: "run error",
			giveSetup: func(pPreRun *cmdutil_mocks.CommandFunc, preRun *cmdutil_mocks.CommandFunc, run *cmdutil_mocks.CommandFunc) {
				pPreRun.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				preRun.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				run.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantError:     assert.AnError,
			wantRunCalled: 1,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			mockLogConfig := logging_mocks.NewLogInformer(t)
			mockCommander := cmdutil_mocks.NewCommandInformer(t)
			mockPPreRun := cmdutil_mocks.NewCommandFunc(t)
			mockPreRun := cmdutil_mocks.NewCommandFunc(t)
			mockRun := cmdutil_mocks.NewCommandFunc(t)
			template := &cmdutil.CmdTemplate{
				Use: "mycmd",
			}
			factory := &cmdutil.Factory{
				LogConfig: mockLogConfig,
				Commander: mockCommander,
			}
			tc.giveSetup(mockPPreRun, mockPreRun, mockRun)
			expectedCmd := template.ToCommand(factory, mockPPreRun.Execute, mockPreRun.Execute, mockRun.Execute)
			err := executeCommand(expectedCmd, []string{}...)
			assert.Same(t, tc.wantError, err)
			// run must be defined on the command for it to run when executed
			// so we define run with Maybe() and then check # times it was called
			mockRun.AssertNumberOfCalls(t, "Execute", tc.wantRunCalled)
		})
	}
}

func (suite *ToCommandTestSuite) TestFlagsConfigured() {
	t := suite.T()

	mockLogConfig := logging_mocks.NewLogInformer(t)
	mockCommander := cmdutil_mocks.NewCommandInformer(t)
	template := &cmdutil.CmdTemplate{
		Use: "mycmd",
		Flags: &cmdutil.CommandFlags{
			String: []*flags.Flag[string]{{Name: "myflags", Usage: "this is my usage s"}},
			Bool:   []*flags.Flag[bool]{{Name: "myflagb", Usage: "this is my usage b"}},
		},
	}
	factory := &cmdutil.Factory{
		LogConfig: mockLogConfig,
		Commander: mockCommander,
	}

	cmdMatcher := mock.AnythingOfType("*cobra.Command")
	acfCall := mockCommander.EXPECT().AddFlags(cmdMatcher, template.Flags).Once()
	mfmeCall := mockCommander.EXPECT().MarkFlagsMutuallyExclusive(cmdMatcher, template.MutuallyExclusiveFlags).Once().NotBefore(acfCall)
	aevCall := mockCommander.EXPECT().ApplyEnvVars(cmdMatcher).Return(nil).Once().NotBefore((mfmeCall))
	mockCommander.EXPECT().SetupLogging(cmdMatcher, mockLogConfig).Return(nil).Once().NotBefore(aevCall)
	actualCmd := template.ToCommand(factory, nil, nil, emptyRun)
	err := executeCommand(actualCmd, []string{}...)
	require.NoError(t, err)
}

func (suite *ToCommandTestSuite) TestFlagsNil() {
	t := suite.T()

	template := &cmdutil.CmdTemplate{
		Use:   "mycmd",
		Flags: nil,
	}
	factory := &cmdutil.Factory{}
	actualCmd := template.ToCommand(factory, nil, nil, emptyRun)
	err := executeCommand(actualCmd, []string{}...)
	require.NoError(t, err)
}

func (suite *ToCommandTestSuite) TestFlagsConfigureError() {
	testCases := []struct {
		testDescription string
		giveSetup       func(m *cmdutil_mocks.CommandInformer)
		wantError       error
	}{
		{
			testDescription: "applyenvvars error",
			giveSetup: func(m *cmdutil_mocks.CommandInformer) {
				m.EXPECT().AddFlags(mock.Anything, mock.Anything).Maybe()
				m.EXPECT().MarkFlagsMutuallyExclusive(mock.Anything, mock.Anything).Maybe()
				m.EXPECT().ApplyEnvVars(mock.Anything).Return(assert.AnError).Once()
			},
			wantError: assert.AnError,
		},
		{
			testDescription: "setuplogging error",
			giveSetup: func(m *cmdutil_mocks.CommandInformer) {
				m.EXPECT().AddFlags(mock.Anything, mock.Anything).Maybe()
				m.EXPECT().MarkFlagsMutuallyExclusive(mock.Anything, mock.Anything).Maybe()
				m.EXPECT().ApplyEnvVars(mock.Anything).Return(nil).Maybe()
				m.EXPECT().SetupLogging(mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()

			mockCommander := cmdutil_mocks.NewCommandInformer(t)
			template := &cmdutil.CmdTemplate{
				Use: "mycmd",
				Flags: &cmdutil.CommandFlags{
					String: []*flags.Flag[string]{{Name: "myflags", Usage: "this is my usage s"}},
					Bool:   []*flags.Flag[bool]{{Name: "myflagb", Usage: "this is my usage b"}},
				},
				MutuallyExclusiveFlags: [][]string{{"flag1", "flag2"}},
			}
			factory := &cmdutil.Factory{
				Commander: mockCommander,
			}

			tc.giveSetup(mockCommander)
			actualCmd := template.ToCommand(factory, nil, nil, emptyRun)
			err := executeCommand(actualCmd, []string{}...)
			assert.Same(t, tc.wantError, err)
		})
	}
}

func (suite *ToCommandTestSuite) TestInvalidTemplate() {
	testCases := []struct {
		testDescription string
		giveTemplate    *cmdutil.CmdTemplate
		wantError       string
	}{
		{
			testDescription: "use empty",
			giveTemplate:    &cmdutil.CmdTemplate{},
			wantError:       "must provide a value for Use",
		},
		{
			testDescription: "mutually exlusive without flags",
			giveTemplate: &cmdutil.CmdTemplate{
				Use:                    "mycmd",
				MutuallyExclusiveFlags: [][]string{{"foo", "bar"}},
			},
			wantError: fmt.Sprintf("command %q must contain flags to have flags marked mutually exclusive", "mycmd"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			assert.PanicsWithError(t, tc.wantError, func() {
				tc.giveTemplate.ToCommand(&cmdutil.Factory{}, nil, nil, nil)
			})
		})
	}
}

func TestToCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ToCommandTestSuite))
}
