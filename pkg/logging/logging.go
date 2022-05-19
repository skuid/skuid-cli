package logging

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/flags"
)

const (
	SEPARATOR_LENGTH = 50
)

var (
	logger             = logrus.New()
	fileLoggingEnabled bool
	loggingDirectory   string
	fileStringFormat   = func() (ret string) {
		ret = time.RFC822
		ret = strings.ReplaceAll(ret, " ", "-")
		ret = strings.ReplaceAll(ret, ":", "-")
		return
	}()
	logStringFmt = fmt.Sprintf("[%v]: %%v\n", time.RFC1123Z)
)

func init() {
	color.Enable = false
}

func InitializeLogging(cmd *cobra.Command, _ []string) (err error) {
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	// log.Out = os.Stdout

	// You could set this to any `io.Writer` such as a file
	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err == nil {
	//  log.Out = file
	// } else {
	//  log.Info("Failed to log to file, using default stderr")
	// }

	if fileLoggingEnabled, err = cmd.Flags().GetBool(flags.FileLogging.Name); err != nil {
		return
	}

	if loggingDirectory, err = cmd.Flags().GetString(flags.FileLoggingDirectory.Name); err != nil {
		return
	}

	// try to open a file for this run off this
	if fileLoggingEnabled {
		// 	var wd string
		// if wd, err = os.Getwd(); err != nil {
		// 	return
		// }

		// ls.FileName = fmt.Sprintf(fileStringFormat+".log", ls.RunStart)

		// var fp string
		// if strings.Contains(ls.FileLocation, wd) {
		// 	fp = path.Join(ls.FileLocation, ls.FileName)
		// } else {
		// 	fp = path.Join(wd, ls.FileLocation, ls.FileName)
		// }

		// if ls.File, err = os.Open(fp); err != nil {
		// 	return
		// }
		// if err = fsLogSystem.OpenFile(); err != nil {
		// 	return
		// } else {
		// 	fsLogSystem.Write()
		// }
	}

	return
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

func FatalF(msg string, args ...interface{}) {
	log.Fatalf(color.Red.Sprint(msg), args)
}
