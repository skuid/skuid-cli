package util

import (
	"io/ioutil"
	"os"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/skuid/skuid-cli/pkg/logging"
)

type WritePayload struct {
	PlanName string
	PlanData []byte
}

func WriteResultsToDisk(targetDirectory string, result WritePayload) (err error) {
	return WriteResults(targetDirectory, result, CopyToFile, CreateDirectoryDeep, ioutil.ReadFile)
}

func WriteResults(targetDirectory string, result WritePayload, copyToFile FileCreator, createDirectoryDeep DirectoryCreator, ioutilReadFile FileReader) (err error) {
	fields := logrus.Fields{
		"function": "WriteResultsToDiskInjection",
	}
	logging.WithFields(fields)

	// unzip the archive into the output directory
	targetDirFriendly, err := SanitizePath(targetDirectory)
	if err != nil {
		return err
	}

	if targetDirectory != "" {
		if err := createDirectoryDeep(targetDirectory, 0755); err != nil {
			logging.Get().Tracef("Error making target dir: %v", err.Error())
		}
	}

	logging.Get().Tracef("Writing results to %v\n", color.Cyan.Sprint(targetDirFriendly))

	tmpFileName, err := CreateTemporaryFile(result.PlanName, result.PlanData)
	if err != nil {
		logging.Get().WithFields(logrus.Fields{
			"fileName": tmpFileName,
		}).
			WithError(err).
			Error("error creating temporary file")
		return err
	}
	defer os.Remove(tmpFileName)

	// unzip the contents of our temp zip file
	err = UnzipArchive(
		tmpFileName,
		targetDirectory,
		copyToFile,
		createDirectoryDeep,
		ioutilReadFile,
	)

	if err != nil {
		logging.Get().WithFields(logrus.Fields{
			"fileName":                tmpFileName,
			"targetDirectoryFriendly": targetDirFriendly,
		}).WithError(err).Warn("Error with UnzipArchive")
		return err
	}

	logging.Get().Debugf("%v results written to %s\n", color.Magenta.Sprint(result.PlanName), color.Cyan.Sprint(targetDirFriendly))

	return nil
}
