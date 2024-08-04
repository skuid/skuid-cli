package cmd_test

import (
	"testing"

	"github.com/skuid/skuid-cli/cmd"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/logging/mocks"
	"github.com/skuid/skuid-cli/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRoot(t *testing.T) {
	factory := &cmdutil.Factory{}
	c := cmd.NewCmdRoot(factory)
	assert.NotNil(t, c)
}

func TestSetupSuccess(t *testing.T) {
	loggingOptions := &logging.LoggingOptions{
		Verbose:        true,
		Trace:          true,
		Diagnostic:     true,
		FileLogging:    true,
		FileLoggingDir: "mylogs",
		NoConsole:      true,
	}
	envVars := map[string]testutil.EnvVar[string]{
		cmdutil.EnvVarName(flags.Verbose.Name):      {EnvValue: "1", Value: "1"},
		cmdutil.EnvVarName(flags.Trace.Name):        {EnvValue: "1", Value: "1"},
		cmdutil.EnvVarName(flags.Diagnostic.Name):   {EnvValue: "1", Value: "1"},
		cmdutil.EnvVarName(flags.FileLogging.Name):  {EnvValue: "1", Value: "1"},
		cmdutil.EnvVarName(flags.LogDirectory.Name): {EnvValue: "mylogs", Value: "mylogs"},
		cmdutil.EnvVarName(flags.NoConsole.Name):    {EnvValue: "1", Value: "1"},
	}
	testutil.SetupEnv(t, envVars)
	mockLI := mocks.NewLogInformer(t)
	// we expect the logging options to match the environment variable values
	// which ensures they were applied prior to Setup being called
	mockLI.EXPECT().Setup(loggingOptions).Return(nil).Once()
	factory := &cmdutil.Factory{
		LogConfig: mockLI,
	}
	c := cmd.NewCmdRoot(factory)
	// force it to run
	c.RunE = testutil.EmptyCobraRun

	err := testutil.ExecuteCommand(c, []string{}...)
	require.NoError(t, err)
}

func TestSetupError(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveEnvVars     map[string]testutil.EnvVar[string]
		giveSetup       func(li *mocks.LogInformer)
	}{
		{
			testDescription: "applyEnvVars error",
			giveEnvVars: map[string]testutil.EnvVar[string]{
				cmdutil.EnvVarName(flags.Verbose.Name): {EnvValue: "blahblahblah", Value: ""},
			},
			giveSetup: func(li *mocks.LogInformer) {
				// should not be called
			},
		},
		{
			testDescription: "log setup error",
			giveEnvVars: map[string]testutil.EnvVar[string]{
				cmdutil.EnvVarName(flags.Verbose.Name): {EnvValue: "1", Value: "1"},
			},
			giveSetup: func(li *mocks.LogInformer) {
				li.EXPECT().Setup(mock.Anything).Return(assert.AnError).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			testutil.SetupEnv(t, tc.giveEnvVars)
			mockLI := mocks.NewLogInformer(t)
			tc.giveSetup(mockLI)
			factory := &cmdutil.Factory{
				LogConfig: mockLI,
			}
			c := cmd.NewCmdRoot(factory)
			c.RunE = testutil.EmptyCobraRun
			err := testutil.ExecuteCommand(c, []string{}...)
			require.Error(t, err)
		})
	}
}
