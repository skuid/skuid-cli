package logging

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
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
