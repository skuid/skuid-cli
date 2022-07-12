//go:build !test

package util

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/skuid/tides/pkg/logging"
)

// add thread protection
var (
	unzipMutex sync.Mutex
)

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func UnzipArchive(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) (err error) {
	unzipMutex.Lock()
	defer unzipMutex.Unlock()

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

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err = directoryCreator(targetLocation, 0755); err != nil {
			logging.Get().Warnf("directoryCreator: %v", err)
			return
		}
	}

	for _, file := range reader.File {
		path := filepath.Join(targetLocation, filepath.FromSlash(file.Name))
		// Check to see if we've already written to this file in this retrieve
		_, fileAlreadyWritten := pathMap[path]
		if !fileAlreadyWritten {
			pathMap[path] = true
		}

		logging.Get().Tracef("Extracting from Zip: %v", color.Blue.Sprint(path))

		// If this file name contains a /, make sure that we create the directory it belongs in
		if pathParts := strings.Split(path, string(filepath.Separator)); len(pathParts) > 0 {
			// Remove the actual file name from the slice,
			// i.e. pages/MyAwesomePage.xml ---> pages
			pathParts = pathParts[:len(pathParts)-1]
			// and then make dirs for all paths up to that point, i.e. pages, apps
			if intermediatePath := filepath.Join(strings.Join(pathParts[:], string(filepath.Separator))); intermediatePath != "" {
				if err = directoryCreator(intermediatePath, 0755); err != nil {
					logging.Get().Warnf("Unable to create intermediary directory: %v", err)
					return
				}
			} else {
				// If we don't have an intermediate path, skip out.
				// Currently Skuid CLI does not create any files in the base directory
				return nil
			}
		}

		if file.FileInfo().IsDir() {
			logging.Get().Tracef("Creating Directory: %v", color.Blue.Sprint(file.Name))
			return directoryCreator(path, file.Mode())
		}

		var fileReader io.ReadCloser
		if fileReader, err = file.Open(); err != nil {
			logging.Get().Warnf("Error opening file: %v", err)
			return
		}
		defer fileReader.Close()

		if filepath.Ext(path) == ".json" {
			if fileReader, err = SanitizeZip(fileReader); err != nil {
				logging.Get().Warnf("Error Sanitizing Zip: %v", err)
				return
			}
		}

		if fileAlreadyWritten {
			if fileReader, err = CombineJSON(fileReader, existingFileReader, path); err != nil {
				logging.Get().Warnf("Error Combining JSON: %v", err)
				return
			}
		}

		if err = fileCreator(fileReader, path); err != nil {
			logging.Get().Warnf("Error with file creator: %v", err)
			return
		}
	}

	return
}

func SanitizeZip(reader io.ReadCloser) (newReader io.ReadCloser, err error) {
	var b []byte
	if b, err = ioutil.ReadAll(reader); err != nil {
		logging.Get().Warnf("unable to read all: %v", err)
		return
	}
	defer reader.Close()

	if b, err = ReSortJson(b); err != nil {
		logging.Get().Warnf("unable to re-sort: %v", err)
		return
	}

	newReader = ioutil.NopCloser(bytes.NewBuffer(b))

	return
}
