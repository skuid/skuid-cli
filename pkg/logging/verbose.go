package logging

import "time"

var (
	// let's use a file global variable
	// do not access this
	isVerbose = false
)

func SetVerbose(verbosity bool) {
	isVerbose = verbosity
}

// VerboseSection creates some visual space for a title in verbose mode
func VerboseSection(description string) {
	if isVerbose {
		PrintSeparator()
		Println(description)
	}
}

// VerboseLn will only call Println if ArgVerbose is true.
func VerboseLn(args ...interface{}) {
	if isVerbose {
		Println(args...)
	}
}

// VerboseLn will call SuccessWithTime if ArgVerbose is true.
// Otherwise, it just prints a success message
func VerboseSuccess(msg string, t time.Time) {
	if isVerbose {
		SuccessWithTime(msg, t)
	}
	// } else {
	// 	Println(msg + ".")
	// }
}

// VerboseCommand only prints the command if ArgVerbose is true.
func VerboseCommand(commandName string) {
	if isVerbose {
		PrintCommand(commandName)
	}
}

// VerboseError logs an error if ArgVerbose is true.
func VerboseError(description string, err error) {
	if isVerbose {
		PrintError(description, err)
	}
}

// VerboseF is a gated printF
func VerboseF(msg string, args ...interface{}) {
	if isVerbose {
		Printf(msg, args...)
	}
}

// VerboseSeparator is a gated separator
func VerboseSeparator() {
	if isVerbose {
		PrintSeparator()
	}
}
