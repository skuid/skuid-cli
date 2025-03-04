//go:build !test

package util

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/skuid/skuid-cli/pkg/logging"
)

var (
	pathMap = make(map[string]bool, 0)
)

func ResetPathMap() {
	pathMap = make(map[string]bool)
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func UnzipArchive(sourceFileLocation, targetLocation string, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) (err error) {
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

		// Sanitize all metadata .json files that aren't included as files on the site itself
		if !strings.Contains(path, "files/") {
			if filepath.Ext(path) == ".json" {
				if fileReader, err = SanitizeZip(fileReader); err != nil {
					logging.Get().Warnf("Error Sanitizing Zip: %v", err)
					return
				}
			}
		}

		if fileAlreadyWritten {
			if fileReader, err = CombineJSON(fileReader, existingFileReader, path); err != nil {
				logging.Get().Warnf("Error Combining JSON: %v", err)
				return
			}
		}

		err = fileCreator(fileReader, path)
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

	if b, err = ReSortJson(b); err != nil {
		logging.Get().Warnf("unable to re-sort: %v", err)
		return
	}

	newReader = io.NopCloser(bytes.NewBuffer(b))

	return
}
