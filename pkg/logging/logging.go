package logging

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gookit/color"

	"github.com/skuid/tides/pkg/constants"
)

const (
	SEPARATOR_LENGTH = 50
)

func init() {
	color.Enable = true
}

// Println redirects through the color package
// to print a gray colored line
func Println(args ...interface{}) {
	fmt.Println(args...)
}

// Printf redirects through the color package
// to print a gray colored message
func Printf(formatString string, args ...interface{}) {
	fmt.Printf(formatString, args...)
}

// PrintSeparator creates some visual space
func PrintSeparator() {
	fmt.Println(strings.Repeat("-", SEPARATOR_LENGTH))
}

// PrintErrorSeparator creates some visual space
func PrintErrorSeparator() {
	color.Red.Println(strings.Repeat("*", SEPARATOR_LENGTH))
}

// Printcommand is a command that outputs separated
// version and command information based on what we're running
func PrintCommand(commandName string) {
	PrintSeparator()

	vName := func() string {
		if constants.VERSION_NAME != "unknown" {
			return color.Blue.Sprint(constants.VERSION_NAME)
		} else {
			return color.Red.Sprint(constants.VERSION_NAME)
		}
	}()

	Println(color.Cyan.Sprintf("Skuid Tides Version:\t%v", vName))
	Println(color.Cyan.Sprintf("Running Command:  \t%v", color.Green.Sprint(commandName)))

	PrintSeparator()
}

// PrintError formats an error with a description and message
func PrintError(description string, err error) {
	PrintErrorSeparator()
	color.Red.Println(description)
	color.Red.Println(err.Error())
	PrintErrorSeparator()
}

// SuccessWithTime Returns a formatted string with a duration since a start time
func SuccessWithTime(description string, timeStart time.Time) {
	color.Gray.Printf("%-25s\t%v\n", fmt.Sprintf("%v:", description), color.Green.Sprint(time.Since(timeStart)))
}

func Fatal(err error) {
	log.Fatal(color.Red.Sprintf("ERROR: %v", err))
}
