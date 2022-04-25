package main

import (
	"strings"
	"time"

	"github.com/gookit/color"
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
	color.Gray.Println(args...)
}

// Printf redirects through the color package
// to print a gray colored message
func Printf(formatString string, args ...interface{}) {
	color.Gray.Printf(formatString, args...)
}

// Separator creates some visual space
func Separator() {
	Println(strings.Repeat("-", SEPARATOR_LENGTH))
}

// ErrorSeparator creates some visual space
func ErrorSeparator() {
	color.Red.Println(strings.Repeat("*", SEPARATOR_LENGTH))
}

// Printcommand is a command that outputs separated
// version and command information based on what we're running
func PrintCommand(commandName string) {
	Separator()

	vName := func() string {
		if VERSION_NAME != "unknown" {
			return color.Blue.Sprint(VERSION_NAME)
		} else {
			return color.Red.Sprint(VERSION_NAME)
		}
	}()

	Println(color.Blue.Sprintf("Skuid Tides Version:\t%v", vName))
	Println(color.Yellow.Sprintf("Running Command:  \t%v", color.Green.Sprint(commandName)))

	Separator()
}

// PrintError formats an error with a description and message
func PrintError(description string, err error) {
	ErrorSeparator()
	color.Red.Printf("%v: %v\n", description, err.Error())
	ErrorSeparator()
}

// SuccessWithTime Returns a formatted string with a duration since a start time
func SuccessWithTime(description string, timeStart time.Time) {
	color.Gray.Printf("%-30s:\t(%v)\n", description, color.Green.Sprint(time.Since(timeStart)))
}

// VerboseSection creates some visual space for a title in verbose mode
func VerboseSection(description string) {
	if ArgVerbose {
		Separator()
		Println(description)
	}
}

// VerboseLn will only call Println if ArgVerbose is true.
func VerboseLn(args ...interface{}) {
	if ArgVerbose {
		Println(args...)
	}
}

// VerboseLn will call SuccessWithTime if ArgVerbose is true.
// Otherwise, it just prints a success message
func VerboseSuccess(msg string, t time.Time) {
	if ArgVerbose {
		SuccessWithTime(msg, t)
	} else {
		Println(msg + ".")
	}
}

// VerboseCommand only prints the command if ArgVerbose is true.
func VerboseCommand(commandName string) {
	if ArgVerbose {
		PrintCommand(commandName)
	}
}

// VerboseError logs an error if ArgVerbose is true.
func VerboseError(description string, err error) {
	if ArgVerbose {
		PrintError(description, err)
	}
}
