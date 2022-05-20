package logging

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/skuid/tides/pkg/constants"
)

const (
	SEPARATOR_LENGTH = 50
)

var (
	Logger        *logrus.Logger
	LineSeparator = strings.Repeat("-", SEPARATOR_LENGTH)
	StarSeparator = strings.Repeat("*", SEPARATOR_LENGTH)

	fileStringFormat = func() (ret string) {
		ret = time.RFC3339
		ret = strings.ReplaceAll(ret, " ", "")
		ret = strings.ReplaceAll(ret, ":", "")
		return
	}()
)

func SetFileLogging(loggingDirectory string) (err error) {
	var wd string
	if wd, err = os.Getwd(); err != nil {
		return
	}

	var dir string
	if strings.Contains(loggingDirectory, wd) {
		dir = loggingDirectory
	} else {
		dir = path.Join(wd, loggingDirectory)
	}

	if stat, e := os.Stat(dir); e != nil {
		if err = os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		err = fmt.Errorf("Directory required at loc: %v", dir)
		return
	}

	logFileName := time.Now().Format(fileStringFormat) + ".log"

	if _, err = os.Create(path.Join(dir, logFileName)); err != nil {
		return
	}

	if file, err := os.OpenFile(path.Join(dir, logFileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		return err
	} else {
		Logger.SetOutput(file)
		Logger.SetFormatter(&logrus.TextFormatter{})
		color.Enable = false
	}

	return
}

func init() {
	Logger = logrus.New()

	// Log as JSON instead of the default ASCII formatter.
	Logger.SetFormatter(&logrus.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	Logger.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	Logger.SetLevel(logrus.TraceLevel)
}

// Println redirects through the color package
// to print a gray colored line
func Println(args ...interface{}) {
	Logger.Log(logrus.InfoLevel, args...)
}

// Printf redirects through the color package
// to print a gray colored message
func Printf(formatString string, args ...interface{}) {
	Logger.Logf(logrus.InfoLevel, formatString, args...)
}

// PrintSeparator creates some visual space
func PrintSeparator() {
	Logger.Logln(logrus.InfoLevel, LineSeparator)
}

// PrintErrorSeparator creates some visual space
func PrintErrorSeparator() {
	Logger.Logln(logrus.ErrorLevel, StarSeparator)
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
	Logger.Logln(logrus.ErrorLevel, description)
	Logger.Logln(logrus.ErrorLevel, err.Error())
	PrintErrorSeparator()
}

// SuccessWithTime Returns a formatted string with a duration since a start time
func SuccessWithTime(description string, timeStart time.Time) {
	Logger.Logf(logrus.InfoLevel, "%-25s\t%v\n", fmt.Sprintf("%v:", description), color.Green.Sprint(time.Since(timeStart)))
}

func Fatal(err error) {
	Logger.Fatal(color.Red.Sprintf("ERROR: %v", err))
}

func Fatalf(msg string, args ...interface{}) {
	Logger.Fatalf(color.Red.Sprint(msg), args)
}

func SetVerbose(verbosity bool) {
	if verbosity {
		Logger.SetLevel(logrus.WarnLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

func SetDebug(debugging bool) {
	if debugging {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

// VerboseSection creates some visual space for a title in verbose mode
func VerboseSection(description string) {
	Logger.Debugln(LineSeparator)
	Logger.Debugln(description)
}

// VerboseLn will only call Println if ArgVerbose is true.
func VerboseLn(args ...interface{}) {
	Logger.Debugln(args...)
}

// VerboseLn will call SuccessWithTime if ArgVerbose is true.
// Otherwise, it just prints a success message
func VerboseSuccess(msg string, t time.Time) {
	Logger.WithField("duration", time.Since(t)).Debug(msg)
}

// DebugCommand only prints the command if ArgVerbose is true.
func DebugCommand(commandName string) {
	Logger.Debug(commandName)
}

// DebugError logs an error if ArgVerbose is true.
func DebugError(description string, err error) {
	Logger.WithError(err).Debug(description)
}

func Debug(msg string) {
	Logger.Debug(msg)
}

// Debugf is a gated printF
func Debugf(msg string, args ...interface{}) {
	Logger.Debugf(msg, args...)
}

// VerboseSeparator is a gated separator
func VerboseSeparator() {
	Logger.Debugln(LineSeparator)
}

func TraceF(msg string, args ...interface{}) {
	Logger.Tracef(msg, args...)
}

func TraceLn(msg string) {
	Logger.Traceln(msg)
}

func TraceErr(msg string, err error) {
	Logger.WithError(err).Trace(msg)
}
