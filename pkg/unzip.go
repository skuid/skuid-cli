package pkg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	pathMap = make(map[string]bool, 0)
)

type WritePayload struct {
	PlanName string
	PlanData []byte
}

func WriteResultsToDisk(targetDirectory string, result WritePayload) (err error) {
	return WriteResults(targetDirectory, result, util.CopyToFile, util.CreateDirectoryDeep, os.ReadFile)
}

func WriteResults(targetDirectory string, result WritePayload, copyToFile util.FileCreator, createDirectoryDeep util.DirectoryCreator, ioutilReadFile util.FileReader) (err error) {
	if !filepath.IsAbs(targetDirectory) {
		err = fmt.Errorf("targetDirectory must be an absolute path")
		return
	}

	fields := logrus.Fields{
		"function": "WriteResultsToDiskInjection",
	}
	logging.WithFields(fields)

	if err := createDirectoryDeep(targetDirectory, 0755); err != nil {
		logging.Get().Tracef("Error making target dir: %v", err.Error())
	}

	logging.Get().Tracef("Writing results to %v\n", color.Cyan.Sprint(targetDirectory))

	tmpFileName, err := util.CreateTemporaryFile(result.PlanName, result.PlanData)
	if err != nil {
		logging.Get().WithFields(logrus.Fields{
			"fileName": tmpFileName,
		}).
			WithError(err).
			Error("error creating temporary file")
		return err
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpFileName)

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
			"fileName":        tmpFileName,
			"targetDirectory": targetDirectory,
		}).WithError(err).Warn("Error with UnzipArchive")
		return err
	}

	logging.Get().Debugf("%v results written to %s\n", color.Magenta.Sprint(result.PlanName), color.Cyan.Sprint(targetDirectory))

	return nil
}

func ResetPathMap() {
	pathMap = make(map[string]bool)
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func UnzipArchive(sourceFileLocation, targetLocation string, fileCreator util.FileCreator, directoryCreator util.DirectoryCreator, existingFileReader util.FileReader) (err error) {
	fields := logrus.Fields{
		"function":           "UnzipArchive",
		"sourceFileLocation": sourceFileLocation,
		"targetLocation":     targetLocation,
	}
	logging.WithFields(fields).Tracef("Unzipping Archive: %v => %v", color.Green.Sprint(sourceFileLocation), color.Blue.Sprint(targetLocation))
	var reader *zip.ReadCloser
	if reader, err = zip.OpenReader(sourceFileLocation); err != nil {
		logging.Get().WithError(err).Warn("unable to unzip archive")
		return
	}
	defer reader.Close()

	if err = directoryCreator(targetLocation, 0755); err != nil {
		logging.Get().Warnf("directoryCreator: %v", err)
		return
	}

	for _, file := range reader.File {
		archivePath := filepath.FromSlash(file.Name)

		// Skuid Review Required - This code is unmodified but it seems very odd.  Why would we ever have a file in the archive twice?  What are the valid
		// scenarios where this could happen?
		// Check to see if we've already written to this file in this retrieve
		_, fileAlreadyWritten := pathMap[archivePath]
		if !fileAlreadyWritten {
			pathMap[archivePath] = true
		}

		logging.Get().Tracef("Extracting from Zip: %v", color.Blue.Sprint(archivePath))

		entityFile, entityFileErr := metadata.NewMetadataEntityFile(archivePath)
		if entityFileErr != nil {
			// Skuid Review Required - This code is unmodified from a behavior perspective but it raises questions around leaving the local file system in an
			// unexpected state.  The server gave us a file and we can't interpret it so something more than logging and continuing seems the more prudent approach.
			// TODO: We need to do something more than just "skip" and log an error - not sure what can be done but we received an invalid file
			// here and it warrants more than just "ok, thanks but no thanks, moving on".  Possibly the approach to unzipping changes where the payload
			// is inspected prior to any local file changes (e.g, don't even delete the existing dirs until payload is validated) and only if valid, are local
			// file changes made
			logging.Get().Warnf("unable to write file to disk, could not resolve metadata information for %q in archive %q: %v", archivePath, sourceFileLocation, entityFileErr)
			continue
		}

		filePath := filepath.Join(targetLocation, entityFile.Path)
		fileDir := filepath.Dir(filePath)
		if err = directoryCreator(fileDir, 0755); err != nil {
			logging.Get().Warnf("Unable to create %v directory for file %v in archive %v: ", fileDir, entityFile.Path, sourceFileLocation)
			return
		}

		// TODO: Skuid Review Required - See https://github.com/skuid/skuid-cli/issues/145
		// This code existed in 0.6.7 but I don't think it's a valid scenario as archives cannot have directories.  Does this need to be here,
		// can there be a Dir in an archive and if there is one, why do we return and not process the remainder of the archive?  Leaving this
		// code for now since I don't fully understand it and don't think it will ever get hit since archives cannot contain directories (at
		// least from what I am aware of).
		if file.FileInfo().IsDir() {
			logging.Get().Tracef("Creating Directory: %v", color.Blue.Sprint(file.Name))
			return directoryCreator(filePath, file.Mode())
		}

		var fileReader io.ReadCloser
		if fileReader, err = file.Open(); err != nil {
			logging.Get().Warnf("Error opening file: %v", err)
			return
		}

		// Skuid Review Required - This code has been modified to applying "SanitizeZip" to all entity definition files.  The previous
		// code did not apply this to all definition files which I believe was/is the intent of what is being done here.  The question
		// is to confirm what the intent was/is and if the change made is correct to account for the issue described in
		// https://github.com/skuid/skuid-cli/issues/198?
		// Sanitize all metadata definition
		if entityFile.IsEntityDefinitionFile {
			if fileReader, err = SanitizeZip(fileReader); err != nil {
				logging.Get().Warnf("Error Sanitizing Zip: %v", err)
				return
			}
		}

		if fileAlreadyWritten {
			if fileReader, err = util.CombineJSON(fileReader, existingFileReader, filePath); err != nil {
				logging.Get().Warnf("Error Combining JSON: %v", err)
				return
			}
		}

		err = fileCreator(fileReader, filePath)
		if err != nil {
			logging.Get().Warnf("Error with file creator: %v", err)
			return
		}
		err = fileReader.Close()
		if err != nil {
			return err
		}
	}

	return
}

func SanitizeZip(reader io.ReadCloser) (newReader io.ReadCloser, err error) {
	var b []byte
	if b, err = io.ReadAll(reader); err != nil {
		logging.Get().Warnf("unable to read all: %v", err)
		return
	}
	defer reader.Close()

	if b, err = util.ReSortJson(b); err != nil {
		logging.Get().Warnf("unable to re-sort: %v", err)
		return
	}

	newReader = io.NopCloser(bytes.NewBuffer(b))

	return
}
