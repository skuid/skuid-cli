package logging

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

const (
	SEPARATOR_LENGTH = 50
)

// Wrapping the singleton to support testing - without this, any tests written to test the "Set*" methods
// would use the global singleton logging package methods which would change the actual logging
// occuring during the test.  For example, testing SetFileLogging would change logging to go to a file
// during the test itself.  By abstracting via interface, we can unit test methods that require
// LogInformer without changing the actual logging occuring during the test.
// TODO: This should be expanded to expose all logging methods as a long which can be passed down
// from factory to avoid global singleton for logging.
// See: https://github.com/skuid/skuid-cli/issues/166
// And the following related issues:
// https://github.com/skuid/skuid-cli/issues/169
// https://github.com/skuid/skuid-cli/issues/174
// https://github.com/skuid/skuid-cli/issues/175
// https://github.com/skuid/skuid-cli/issues/176
// https://github.com/skuid/skuid-cli/issues/180
// https://github.com/skuid/skuid-cli/issues/181
type LogInformer interface {
	SetVerbose()
	SetTrace()
	SetDiagnostic()
	SetFileLogging(loggingDirectory string) (err error)
}

type LogConfig struct{}

func (l *LogConfig) SetVerbose() {
	SetVerbose()
}
func (l *LogConfig) SetTrace() {
	SetTrace()
}
func (l *LogConfig) SetDiagnostic() {
	SetDiagnostic()
}
func (l *LogConfig) SetFileLogging(loggingDirectory string) (err error) {
	return SetFileLogging(loggingDirectory)
}

func NewLogConfig() LogInformer {
	return &LogConfig{}
}

var (
	safe                 sync.Mutex
	loggerSingleton      logrus.Ext1FieldLogger
	fileLogging          bool
	fileLoggingDirectory string
	logFile              *os.File
	fieldLogging         bool
	LineSeparator        = strings.Repeat("-", SEPARATOR_LENGTH)
	StarSeparator        = strings.Repeat("*", SEPARATOR_LENGTH)

	fileStringFormat = func() (ret string) {
		ret = time.RFC3339
		ret = strings.ReplaceAll(ret, " ", "")
		ret = strings.ReplaceAll(ret, ":", "")
		return
	}()
)

func Level(logger logrus.Ext1FieldLogger) logrus.Level {
	if l, ok := (logger).(*logrus.Logger); ok {
		return l.Level
	} else {
		return logrus.PanicLevel
	}
}

func SetVerbose() logrus.Ext1FieldLogger {
	loggerSingleton = Get()
	l, _ := loggerSingleton.(*logrus.Logger)
	l.SetLevel(logrus.DebugLevel)
	return loggerSingleton
}

func SetTrace() logrus.Ext1FieldLogger {
	loggerSingleton = Get()
	l, _ := loggerSingleton.(*logrus.Logger)
	l.SetLevel(logrus.TraceLevel)
	return loggerSingleton
}

func SetDiagnostic() logrus.Ext1FieldLogger {
	loggerSingleton = Get()
	l, _ := loggerSingleton.(*logrus.Logger)
	l.SetLevel(logrus.TraceLevel)
	SetFieldLogging(true)
	return loggerSingleton
}

func SetFieldLogging(b bool) {
	fieldLogging = b
}

func GetFieldLogging() bool {
	return fieldLogging
}

func SetFileLogging(loggingDirectory string) error {
	l, ok := Get().(*logrus.Logger)
	if !ok {
		return fmt.Errorf("unable to obtain logger")
	}

	// MkdirAll will do nothing if the directory already exists, create it with specified
	// permissions if not exist and return error otherwise (e.g., the path points to a file)
	if err := os.MkdirAll(loggingDirectory, 0777); err != nil {
		if errors.Is(err, syscall.ENOTDIR) {
			return fmt.Errorf("must specify a directory for logging directory: %v", loggingDirectory)
		}
		return err
	}

	file, err := os.CreateTemp(loggingDirectory, fmt.Sprintf("skuid-cli-%v-*.log", time.Now().Format(fileStringFormat)))
	if err != nil {
		return err
	}
	// CreateTemp applies 0600
	if err := os.Chmod(file.Name(), 0666); err != nil {
		return err
	}

	resetFileLogging()
	l.SetOutput(file)
	l.SetFormatter(&logrus.TextFormatter{})
	color.Enable = false
	logFile = file
	fileLogging = true
	fileLoggingDirectory = loggingDirectory

	return nil
}

func GetFileLogging() (string, bool) {
	return fileLoggingDirectory, fileLogging
}

func WithFields(fields logrus.Fields) logrus.Ext1FieldLogger {
	loggerSingleton = Get()
	if fileLogging || Level(loggerSingleton) == logrus.TraceLevel && fieldLogging {
		return loggerSingleton.WithFields(fields)
	}
	return loggerSingleton
}

func WithField(field string, value interface{}) logrus.Ext1FieldLogger {
	loggerSingleton = Get()
	if fileLogging || Level(loggerSingleton) == logrus.TraceLevel && fieldLogging {
		return loggerSingleton.WithField(field, value)
	}
	return loggerSingleton
}

func Reset() logrus.Ext1FieldLogger {
	safe.Lock()
	loggerSingleton = nil
	resetFileLogging()
	fieldLogging = false
	safe.Unlock()
	return Get()
}

func DisableLogging() logrus.Ext1FieldLogger {
	loggerSingleton = Get()
	safe.Lock()
	l, _ := loggerSingleton.(*logrus.Logger)
	l.Out = nil
	l.SetLevel(logrus.PanicLevel)
	safe.Unlock()
	return loggerSingleton
}

func Get() logrus.Ext1FieldLogger {
	safe.Lock()
	defer safe.Unlock()
	if loggerSingleton != nil {
		return loggerSingleton
	}

	loggerSingleton = logrus.New()

	l, _ := loggerSingleton.(*logrus.Logger)

	l.SetLevel(logrus.InfoLevel)

	// Log as JSON instead of the default ASCII formatter.
	l.SetFormatter(&logrus.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	l.SetOutput(os.Stdout)

	// could have been changed by previous call to SetFileLogging
	color.Enable = true

	return loggerSingleton
}

func resetFileLogging() {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			Get().Warnf("unable to close the current log file: %v", err)
		}
		logFile = nil
	}
	fileLoggingDirectory = ""
	fileLogging = false
}
