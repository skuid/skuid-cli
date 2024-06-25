package util

import (
	"path/filepath"
	"strings"
)

// FromWindowsPath takes a `path` string and replaces double back-ticks
// with whatever the system filepath separator is.
func FromWindowsPath(path string) string {
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

// GetAbsolutePath gets the absolute path for the directory from the relative path
// if path is not absolute (includes blank), it will be joined with the current
// working directory to turn it in to an absolute path
func GetAbsolutePath(path string) (absolute string, err error) {
	return filepath.Abs(path)
}

// SanitizePath returns the absolute path path for the given directory string.
// If directory is not absolute (includes blank), it will be joined with the
// current working directory to turn it in to an absolute path
func SanitizePath(directory string) (sanitizedResult string, err error) {
	return GetAbsolutePath(directory)
}
