package cmd_test

import (
	"testing"

	"github.com/sirupsen/logrus"
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
		Level:            logrus.TraceLevel,
		Diag:             true,
		FileLogging:      true,
		FileLoggingDir:   "mylogs",
		NoConsoleLogging: true,
	}
	envVars := map[string]testutil.EnvVar[string]{
		cmdutil.EnvVarName(flags.LogLevel.Name):         {EnvValue: "trace", Value: "trace"},
		cmdutil.EnvVarName(flags.Diag.Name):             {EnvValue: "1", Value: "1"},
		cmdutil.EnvVarName(flags.FileLogging.Name):      {EnvValue: "1", Value: "1"},
		cmdutil.EnvVarName(flags.LogDirectory.Name):     {EnvValue: "mylogs", Value: "mylogs"},
		cmdutil.EnvVarName(flags.NoConsoleLogging.Name): {EnvValue: "1", Value: "1"},
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

func TestSupportsDeprecatedLoggingFlags(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveArgs        []string
		giveLevel       logrus.Level
		giveDiag        bool
		wantError       []string
	}{
		{
			testDescription: "verbose specified",
			giveArgs:        []string{"--verbose"},
			giveLevel:       logrus.DebugLevel,
			giveDiag:        false,
		},
		{
			testDescription: "trace specified",
			giveArgs:        []string{"--trace"},
			giveLevel:       logrus.TraceLevel,
			giveDiag:        false,
		},
		{
			testDescription: "diagnostic specified",
			giveArgs:        []string{"--diagnostic"},
			giveLevel:       logrus.TraceLevel,
			giveDiag:        true,
		},
		{
			testDescription: "verbose & trace specified",
			giveArgs:        []string{"--verbose", "--trace"},
			giveLevel:       logrus.TraceLevel,
			giveDiag:        false,
		},
		{
			testDescription: "verbose & diagnostic specified",
			giveArgs:        []string{"--verbose", "--diagnostic"},
			giveLevel:       logrus.TraceLevel,
			giveDiag:        true,
		},
		{
			testDescription: "trace & diagnostic specified",
			giveArgs:        []string{"--trace", "--diagnostic"},
			giveLevel:       logrus.TraceLevel,
			giveDiag:        true,
		},
		{
			testDescription: "verbose & trace & diagnosticspecified",
			giveArgs:        []string{"--verbose", "--trace", "--diagnostic"},
			giveLevel:       logrus.TraceLevel,
			giveDiag:        true,
		},
		{
			testDescription: "log-level & verbose",
			giveArgs:        []string{"--log-level", "warn", "--verbose"},
			giveLevel:       logrus.WarnLevel,
			giveDiag:        false,
			wantError:       []string{"log-level", "verbose"},
		},
		{
			testDescription: "log-level & trace",
			giveArgs:        []string{"--log-level", "warn", "--trace"},
			giveLevel:       logrus.WarnLevel,
			giveDiag:        false,
			wantError:       []string{"log-level", "trace"},
		},
		{
			testDescription: "log-level & diagnostic",
			giveArgs:        []string{"--log-level", "warn", "--diagnostic"},
			giveLevel:       logrus.WarnLevel,
			giveDiag:        false,
			wantError:       []string{"log-level", "diagnostic"},
		},
		{
			testDescription: "diag & verbose",
			giveArgs:        []string{"--diag", "--verbose"},
			giveLevel:       logrus.InfoLevel,
			giveDiag:        true,
			wantError:       []string{"diag", "verbose"},
		},
		{
			testDescription: "diag & trace",
			giveArgs:        []string{"--diag", "--trace"},
			giveLevel:       logrus.InfoLevel,
			giveDiag:        true,
			wantError:       []string{"diag", "trace"},
		},
		{
			testDescription: "diag & diagnostic",
			giveArgs:        []string{"--diag", "--diagnostic"},
			giveLevel:       logrus.InfoLevel,
			giveDiag:        true,
			wantError:       []string{"diag", "diagnostic"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			mockLI := mocks.NewLogInformer(t)
			loggingOptions := &logging.LoggingOptions{
				Level:          tc.giveLevel,
				Diag:           tc.giveDiag,
				FileLoggingDir: flags.LogDirectory.Default,
			}
			mockLI.EXPECT().Setup(loggingOptions).Return(nil).Once()
			factory := &cmdutil.Factory{
				LogConfig: mockLI,
			}
			c := cmd.NewCmdRoot(factory)
			// force it to run
			c.RunE = testutil.EmptyCobraRun

			err := testutil.ExecuteCommand(c, tc.giveArgs...)
			if len(tc.wantError) > 0 {
				require.Error(t, err)
				for _, m := range tc.wantError {
					assert.ErrorContains(t, err, m)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
