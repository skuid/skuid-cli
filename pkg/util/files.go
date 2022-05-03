package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/skuid/tides/pkg/logging"
)

func FromWindowsPath(path string) string {
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

func GetAbs(directory string) (path string) {
	wd, _ := os.Getwd()
	logging.DebugF("Working Directory: %v", wd)

	if strings.Contains(directory, wd) {
		logging.DebugF("Absolute path: %v", directory)
		return directory
	} else {
		logging.DebugLn("Relative Path")
	}

	path, _ = filepath.Abs(filepath.Join(wd, directory))
	logging.DebugF("Target Directory: %v", path)
	return
}
