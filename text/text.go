package text

import (
	"fmt"
	"strings"
	"time"

	"github.com/skuid/skuid-cli/version"
)

// RunCommand outputs text to be shown at the beginning of each command
func RunCommand(commandName string) string {
	return joinError([]string{
		Separator(),
		fmt.Sprintf("Skuid CLI Version %v", version.Name),
		fmt.Sprintf("Running Command: %v", commandName),
		Separator(),
	})
}

// Separator creates some visual space
func Separator() string {
	return "------------------------------------------"
}

// ErrorSeparator creates some visual space
func ErrorSeparator() string {
	return "******************************************"
}

// PrettyError formats an error with a description and message
func PrettyError(description string, err error) string {
	return joinError([]string{
		"",
		ErrorSeparator(),
		fmt.Sprintf("%v:", description),
		err.Error(),
		Separator(),
	})
}

func joinError(errors []string) string {
	return strings.Join(errors, "\n")
}

// AugmentError adds a description in the line above the error
func AugmentError(description string, err error) error {
	return fmt.Errorf("%v\n%v", description, err.Error())
}

// SuccessWithTime Returns a formatted string with a duration since a start time
func SuccessWithTime(description string, timeStart time.Time) string {
	return fmt.Sprintf("%s: (%v)", description, time.Since(timeStart))
}

// VerboseSection creates some visual space for a title in verbose mode
func VerboseSection(description string) string {
	return strings.Join([]string{
		Separator(),
		description,
	}, "\n")
}
