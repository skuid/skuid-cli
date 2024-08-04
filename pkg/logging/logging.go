package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	clierrors "github.com/skuid/skuid-cli/pkg/errors"
)

type LoggingOptions struct {
	Verbose        bool
	Trace          bool
	Diagnostic     bool
	FileLogging    bool
	FileLoggingDir string
	NoConsole      bool
}

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
	Setup(*LoggingOptions) error
	Teardown()
}

type LogConfig struct{}

func (l *LogConfig) Setup(opts *LoggingOptions) error {
	return Setup(opts)
}

func (l *LogConfig) Teardown() {
	Teardown()
}

func NewLogConfig() LogInformer {
	return &LogConfig{}
}

var (
	safe             sync.Mutex
	loggerSingleton  *logrus.Logger
	fieldLogging     bool
	logFile          *os.File
	fileStringFormat = func() (ret string) {
		ret = time.RFC3339
		ret = strings.ReplaceAll(ret, " ", "")
		ret = strings.ReplaceAll(ret, ":", "")
		return
	}()
)

func WithFields(fields logrus.Fields) *logrus.Entry {
	loggerSingleton = Get()
	if fieldLogging {
		return loggerSingleton.WithFields(fields)
	}
	return logrus.NewEntry(loggerSingleton)
}

func WithField(field string, value interface{}) *logrus.Entry {
	loggerSingleton = Get()
	if fieldLogging {
		return loggerSingleton.WithField(field, value)
	}
	return logrus.NewEntry(loggerSingleton)
}

func Get() *logrus.Logger {
	safe.Lock()
	defer safe.Unlock()

	// We want to enforce that no logging occurs prior to logging being initialized to ensure that all
	// logs go to a single location during execution (see
	// https://github.com/skuid/skuid-cli/issues/180 &
	// https://github.com/techfg/skuid-cli/commit/5a85e16589469829a4b52973aeb7795b3a1a0048#diff-dc4c8db195eb54b1b5300cbe47798f318b8c7444bac9475c7668c65d15a665baR32).
	// However, due to the use of a global singleton for logging  (see https://github.com/skuid/skuid-cli/issues/174),
	// while it is safe to assert here during normal execution (because Init will be called in PersistentPreRun
	// before anything else occurs and any log statements emitted), during testing, Init doesn't get called so
	// when code attempts to log a message it will panic.  Until the global singleton is eliminated and migrated to an
	// injected logger (e.g., from Factory and passed down the execution chain - see https://github.com/techfg/skuid-cli/blob/tfgbeta/pkg/cmdutil/factory.go#L14),
	// there are three options:
	//    1. If loggerSingleton is null here, call Init with default options - not desired behavior for normal execution
	//       as we want to enforce that no logging occurs until Init is complete
	//    2. Add a TestMain to every testing package to call Init with default options - not ideal and a lot of code to
	//       add maintain that isn't obvious to the casual observer
	//    3. Detect that we are under test and if logging isn't intialized, initialze with defaults - While not ideal,
	//       this is effectively what the prior version of this code did, it would just always initialize with defaults
	//       instead of only during testing.
	// TODO: This workaround is not ideal and the entire logging package needs to be revisited with proper
	//       DI throughout the entire codebase.  Until then, this workaround achieves desired outcome
	//       of maintaining all current tests while enforcing that during normal execution, Init must
	//       be called prior to any log messages being emitted. Once DI is implemented, the entire logging package
	//       goes away so this will be eliminated.
	if testing.Testing() && loggerSingleton == nil {
		if err := setupLog(&LoggingOptions{}); err != nil {
			panic(err)
		}
	} else {
		// should never occur in production
		// this enforces that during normal execution, no logging occurs prior to logging being initialized
		// as we want to ensure that all logs go to a single location and initialization is required to
		// configure the target location (stdout, file) with the specified logging level.
		clierrors.MustConditionf(loggerSingleton != nil, "logger not initialized")
	}

	return loggerSingleton
}

func Setup(options *LoggingOptions) error {
	safe.Lock()
	defer safe.Unlock()

	return setupLog(options)
}

func Teardown() {
	safe.Lock()
	defer safe.Unlock()

	if logFile != nil {
		// intentionally ignoring any error since there is nothing we can do
		// and we should only be tearing down when we are exiting anyway
		_ = logFile.Close()
		logFile = nil
	}
	loggerSingleton = nil
	fieldLogging = false
}

func setupLog(options *LoggingOptions) error {
	// should never occur in production
	clierrors.MustConditionf(loggerSingleton == nil, "logger already initialized")

	logLevel := logrus.InfoLevel
	if options.Diagnostic || options.Trace {
		logLevel = logrus.TraceLevel
	} else if options.Verbose {
		logLevel = logrus.DebugLevel
	}

	var output io.Writer = os.Stdout
	if options.NoConsole {
		output = io.Discard
	}

	if options.FileLogging {
		var err error
		logFile, err = createLogFile(options.FileLoggingDir)
		if err != nil {
			return err
		}
	}

	loggerSingleton = logrus.New()
	loggerSingleton.SetLevel(logLevel)
	loggerSingleton.SetFormatter(&logrus.TextFormatter{})
	loggerSingleton.SetOutput(output)
	if options.FileLogging {
		loggerSingleton.AddHook(lfshook.NewHook(logFile, &logrus.TextFormatter{DisableColors: true}))
	}
	// force disabling colors applied to log messages (not the entire message itself, only the colors that we explicitly applied to the text
	// inside the message) to avoid ANSI characters in the log file.  We lose colors in console by doing this (we really only want to disable
	// in the file) but the only alternative is to have a custom formatters that strips the ansi codes during the file log.  The colors applied
	// to the entire log message itself by logrus will still appear in the terminal and the lfshook will disable the logrus colors going to the
	// file.  In short, we only lose colors that we applied directly to the text in the message (e.g. color.Green.Sprintf(...)).
	color.Enable = !options.FileLogging
	fieldLogging = options.Diagnostic
	return nil
}

func createLogFile(loggingDirectory string) (*os.File, error) {
	// MkdirAll will do nothing if the directory already exists, create it with specified
	// permissions if not exist and return error otherwise (e.g., the path points to a file)
	if err := os.MkdirAll(loggingDirectory, 0777); err != nil {
		if errors.Is(err, syscall.ENOTDIR) {
			return nil, fmt.Errorf("must specify a directory for logging directory: %v", loggingDirectory)
		}
		return nil, err
	}

	file, err := os.CreateTemp(loggingDirectory, fmt.Sprintf("skuid-cli-%v-*.log", time.Now().Format(fileStringFormat)))
	if err != nil {
		return nil, err
	}
	// CreateTemp applies 0600
	if err := os.Chmod(file.Name(), 0666); err != nil {
		return nil, err
	}

	return file, nil
}
