package util

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"

	"github.com/skuid/tides/pkg/logging"
)

const (
	MAX_ATTEMPTS = 5
)

type FileCreator func(fileReader io.ReadCloser, path string) error

type DirectoryCreator func(path string, fileMode os.FileMode) error

type FileReader func(path string) ([]byte, error)

func CreateTemporaryFile(planName string, data []byte) (name string, err error) {
	var tmpfile *os.File
	var n int
	rand.Seed(time.Now().Unix())

	for attempts := 0; attempts < MAX_ATTEMPTS; attempts++ {
		if tmpfile, err = os.CreateTemp("", strings.ReplaceAll(planName, " ", "-")); err == nil {
			break
		}
		time.NewTimer(time.Second * time.Duration(attempts))
	}

	if err != nil {
		logging.Get().WithError(err).Warn("Couldn't create tempfile")
		return
	}

	logging.Get().Tracef("created temp file: %v", color.Green.Sprintf(tmpfile.Name()))

	if n, err = tmpfile.Write(data); err != nil {
		logging.Get().WithError(err).Warn("Couldn't write to temp file")
		return
	} else if n == 0 {
		err = fmt.Errorf("didn't write anything")
		logging.Get().WithError(err).Warn("wrote nothing to tempfile")
		return
	} else {
		name = tmpfile.Name()
	}

	logging.Get().WithField("tempFileName", name).Tracef("Created Temp File")
	return
}

func CreateDirectoryDeep(path string, fileMode os.FileMode) (err error) {
	if _, err = os.Stat(path); err != nil {
		logging.Get().Tracef("Creating intermediate directory: %v", color.Cyan.Sprint(path))
		err = os.MkdirAll(path, fileMode)
	}
	return
}

func CopyToFile(fileReader io.ReadCloser, path string) (err error) {
	logging.Get().Tracef("%v: %v", color.Yellow.Sprint("Creating File"), path)

	var targetFile *os.File
	if targetFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		logging.Get().WithError(err).Warn("unable to open file in copytofile")
		return
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, fileReader); err != nil {
		logging.Get().WithError(err).Error("unable to copy to target")
		return
	}

	logging.Get().Tracef("%v: %v", color.Yellow.Sprint("Copied to File"), path)

	return
}
