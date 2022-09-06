package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/skuid/tides/pkg/logging"
)

// FromWindowsPath takes a `path` string and replaces double back-ticks
// with whatever the system filepath separator is.
func FromWindowsPath(path string) string {
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

// GetAbsolutePath gets the absolute path for the directory from the relative path
func GetAbsolutePath(relative string) (absolute string) {
	wd, _ := os.Getwd()
	logging.WithFields(logrus.Fields{
		"function": "GetAbsolutePath",
	})
	logging.Get().Tracef("Working Directory: %v", wd)

	if strings.Contains(relative, wd) {
		logging.Get().Tracef("Absolute path: %v", relative)
		return relative
	} else {
		logging.Get().Trace("Relative Path")
	}

	absolute, _ = filepath.Abs(filepath.Join(wd, relative))
	logging.Get().Tracef("Target Directory: %v", absolute)
	return
}

// SanitizePath returns the "friendly" path for the given string.
// This really only occurs if the directory is non-existent...
func SanitizePath(directory string) (sanitizedResult string, err error) {
	if directory == "" {
		return filepath.Abs(filepath.Dir(os.Args[0]))
	} else {
		sanitizedResult = directory
	}
	return
}
