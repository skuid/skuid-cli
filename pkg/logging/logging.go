package logging

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

const (
	SEPARATOR_LENGTH = 50
)

var (
	safe            sync.Mutex
	loggerSingleton Logger
	fileLogging     bool
	LineSeparator   = strings.Repeat("-", SEPARATOR_LENGTH)
	StarSeparator   = strings.Repeat("*", SEPARATOR_LENGTH)

	fileStringFormat = func() (ret string) {
		ret = time.RFC3339
		ret = strings.ReplaceAll(ret, " ", "")
		ret = strings.ReplaceAll(ret, ":", "")
		return
	}()
)

type Logger interface {
	logrus.Ext1FieldLogger
}

func SetVerbose() Logger {
	loggerSingleton = Get()
	loggerSingleton.Info("Setting verbose logging level.")
	l, _ := loggerSingleton.(*logrus.Logger)
	l.SetLevel(logrus.DebugLevel)
	return loggerSingleton
}

func SetTrace() Logger {
	loggerSingleton = Get()
	loggerSingleton.Info("Setting trace logging level.")
	l, _ := loggerSingleton.(*logrus.Logger)
	l.SetLevel(logrus.TraceLevel)
	return loggerSingleton
}

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
		l, _ := Get().(*logrus.Logger)
		l.SetOutput(file)
		l.SetFormatter(&logrus.TextFormatter{})
		color.Enable = false
		fileLogging = true
	}

	return
}

func WithFields(fields logrus.Fields) Logger {
	loggerSingleton = Get()
	if fileLogging {
		loggerSingleton = loggerSingleton.WithFields(fields)
	}
	return loggerSingleton
}

func WithField(field string, value interface{}) Logger {
	loggerSingleton = Get()
	if fileLogging {
		loggerSingleton = loggerSingleton.WithField(field, value)
	}
	return loggerSingleton
}

func Reset() Logger {
	safe.Lock()
	loggerSingleton = nil
	safe.Unlock()
	return Get()
}

func DisableLogging() Logger {
	loggerSingleton = Get()
	safe.Lock()
	l, _ := loggerSingleton.(*logrus.Logger)
	l.Out = nil
	l.SetLevel(logrus.PanicLevel)
	safe.Unlock()
	return loggerSingleton
}

func Get() Logger {
	safe.Lock()
	defer safe.Unlock()
	if loggerSingleton != nil {
		return loggerSingleton
	}

	loggerSingleton = logrus.StandardLogger()

	l, _ := loggerSingleton.(*logrus.Logger)

	// Log as JSON instead of the default ASCII formatter.
	l.SetFormatter(&logrus.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	l.SetOutput(os.Stdout)

	return loggerSingleton
}
