package cmdutil_test

import (
	"errors"
	"testing"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommandError(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveName        string
		giveError       error
		wantName        string
		wantPanic       bool
		wantError       error
	}{
		{
			testDescription: "panics when name empty and err nil",
			giveName:        "",
			giveError:       nil,
			wantName:        "",
			wantPanic:       true,
			wantError:       nil,
		},
		{
			testDescription: "panics when name not empty and err is nil",
			giveName:        "mycmd",
			giveError:       nil,
			wantName:        "mycmd",
			wantPanic:       true,
			wantError:       nil,
		},
		{
			testDescription: "return error when name empty and err not nil",
			giveName:        "",
			giveError:       assert.AnError,
			wantName:        "",
			wantError:       assert.AnError,
		},
		{
			testDescription: "return error when name not empty and err not nil",
			giveName:        "mycmd",
			giveError:       assert.AnError,
			wantName:        "mycmd",
			wantError:       assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			e := cmdutil.NewCommandError(tc.giveName, tc.giveError)
			require.NotNil(t, e)
			var asCmdError *cmdutil.CommandError
			require.ErrorAs(t, e, &asCmdError)
			assert.Equal(t, tc.giveName, asCmdError.Name)
			assert.Equal(t, tc.wantError, asCmdError.Unwrap())
			if tc.wantPanic {
				assert.Panics(t, func() {
					assert.Equal(t, tc.wantError.Error(), asCmdError.Error())
				})
			} else {
				assert.Equal(t, tc.wantError.Error(), asCmdError.Error())
			}
			var isCmdError *cmdutil.CommandError
			assert.False(t, errors.Is(e, isCmdError))
		})
	}
}

func TestCheckError(t *testing.T) {
	testCases := []struct {
		testDescription   string
		giveCmd           *cobra.Command
		giveError         error
		giveRecovered     any
		wantPanic         bool
		wantErrorContains []string
	}{
		{
			testDescription:   "return nil when all params nil",
			giveCmd:           nil,
			giveError:         nil,
			giveRecovered:     nil,
			wantErrorContains: nil,
		},
		{
			testDescription:   "return err when not recovered",
			giveCmd:           nil,
			giveError:         assert.AnError,
			giveRecovered:     nil,
			wantErrorContains: []string{assert.AnError.Error()},
		},
		{
			testDescription:   "panics when recovered not nil and cmd nil",
			giveCmd:           nil,
			giveError:         nil,
			giveRecovered:     "foobar",
			wantPanic:         true,
			wantErrorContains: nil,
		},
		{
			testDescription:   "panics when recovered not nil and err not nil and cmd nil",
			giveCmd:           nil,
			giveError:         assert.AnError,
			giveRecovered:     "foobar",
			wantPanic:         true,
			wantErrorContains: nil,
		},
		{
			testDescription:   "return nil when err nil and recovered nil",
			giveCmd:           &cobra.Command{Use: "mycmd"},
			giveError:         nil,
			giveRecovered:     nil,
			wantErrorContains: nil,
		},
		{
			testDescription:   "return err when recovered nil",
			giveCmd:           &cobra.Command{Use: "mycmd"},
			giveError:         assert.AnError,
			giveRecovered:     nil,
			wantErrorContains: []string{assert.AnError.Error()},
		},
		{
			testDescription:   "return recovered err when recovered not nil and err nil",
			giveCmd:           &cobra.Command{Use: "mycmd"},
			giveError:         nil,
			giveRecovered:     "foobar",
			wantErrorContains: []string{"foobar"},
		},
		{
			testDescription:   "return recovered err with err when recovered not nil and err not nil",
			giveCmd:           &cobra.Command{Use: "mycmd"},
			giveError:         assert.AnError,
			giveRecovered:     "foobar",
			wantErrorContains: []string{assert.AnError.Error(), "foobar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			var e error
			if tc.wantPanic {
				assert.Panics(t, func() {
					e = cmdutil.CheckError(tc.giveCmd, tc.giveError, tc.giveRecovered)
				})
			} else {
				e = cmdutil.CheckError(tc.giveCmd, tc.giveError, tc.giveRecovered)
			}
			if len(tc.wantErrorContains) > 0 {
				for _, m := range tc.wantErrorContains {
					assert.ErrorContains(t, e, m)
				}
			} else {
				assert.Nil(t, e)
			}
		})
	}
}

func TestHandleCommandResult(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveCommand     *cobra.Command
		giveError       error
		giveLogger      *logging.Logger
		giveMessage     string
		wantPanic       bool
		wantError       error
	}{
		{
			testDescription: "return nil and not log",
			giveCommand:     nil,
			giveError:       nil,
			giveLogger:      nil,
			giveMessage:     "",
			wantError:       nil,
		},
		{
			testDescription: "panics when error and cmd nil",
			giveCommand:     nil,
			giveError:       assert.AnError,
			giveLogger:      nil,
			giveMessage:     "some message",
			wantPanic:       true,
		},
		{
			testDescription: "return error and not log when error with empty message",
			giveCommand:     &cobra.Command{Use: "mycmd"},
			giveError:       assert.AnError,
			giveLogger:      nil,
			giveMessage:     "",
			wantError:       assert.AnError,
		},
		{
			testDescription: "return error and not log when error with non-empty message",
			giveCommand:     &cobra.Command{Use: "mycmd"},
			giveError:       assert.AnError,
			giveLogger:      nil,
			giveMessage:     "some message",
			wantError:       assert.AnError,
		},
		{
			testDescription: "does not log when no error with no message",
			giveCommand:     &cobra.Command{Use: "mycmd"},
			giveError:       nil,
			giveLogger:      nil,
			giveMessage:     "",
			wantError:       nil,
		},
		{
			testDescription: "panics when no logger when no error and message",
			giveCommand:     &cobra.Command{Use: "mycmd"},
			giveError:       nil,
			giveLogger:      nil,
			giveMessage:     "some message",
			wantPanic:       true,
		},
		{
			testDescription: "logs when no error with message",
			giveCommand:     &cobra.Command{Use: "mycmd"},
			giveError:       nil,
			giveLogger:      logging.WithName("TestHandleCommandResult"),
			giveMessage:     "some message",
			wantError:       nil,
		},
	}

	for _, tc := range testCases {
		if tc.wantPanic {
			assert.Panics(t, func() {
				cmdutil.HandleCommandResult(tc.giveCommand, tc.giveLogger, tc.giveError, tc.giveMessage)
			})
		} else {
			e := cmdutil.HandleCommandResult(tc.giveCommand, tc.giveLogger, tc.giveError, tc.giveMessage)
			if tc.wantError == nil {
				assert.Nil(t, e)
			} else {
				var cmdError *cmdutil.CommandError
				require.ErrorAs(t, e, &cmdError)
				assert.Equal(t, tc.giveCommand.Name(), cmdError.Name)
				assert.Equal(t, tc.wantError, cmdError.Unwrap())
			}
		}
	}
}
