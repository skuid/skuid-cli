package util

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	jsonpatch "github.com/skuid/json-patch"

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
func SanitizePath(directory string) (friendlyResult string, err error) {
	if directory == "" {
		if friendlyResult, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			return
		}
	} else {
		friendlyResult = directory
	}

	return

}

type WritePayload struct {
	PlanName string
	PlanData []byte
}

func WriteResultsToDisk(targetDirectory string, result WritePayload) (err error) {
	return WriteResultsToDiskInjection(targetDirectory, result, CopyToFile, CreateDirectoryDeep, ioutil.ReadFile)
}

func WriteResultsToDiskInjection(targetDirectory string, result WritePayload, copyToFile FileCreator, createDirectoryDeep DirectoryCreator, ioutilReadFile FileReader) (err error) {
	fields := logrus.Fields{
		"function": "WriteResultsToDiskInjection",
	}
	logging.WithFields(fields)

	// unzip the archive into the output directory
	targetDirFriendly, err := SanitizePath(targetDirectory)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(targetDirFriendly, 0777); err != nil {
		logging.Get().Tracef("Error making target dir: %v", err.Error())
	}

	logging.Get().Tracef("Writing results to %v\n", color.Cyan.Sprint(targetDirFriendly))

	// Store a map of paths that we've already encountered. We'll use this
	// to determine if we need to modify a file or overwrite it.
	pathMap := map[string]bool{}

	tmpFileName, err := CreateTemporaryFile(result.PlanName, result.PlanData)
	if err != nil {
		logging.Get().WithFields(logrus.Fields{
			"fileName": tmpFileName,
		}).
			WithError(err).
			Error("error creating temporary file")
		return err
	}
	// defer os.Remove(tmpFileName)

	// unzip the contents of our temp zip file
	err = UnzipArchive(
		tmpFileName,
		targetDirFriendly,
		pathMap,
		copyToFile,
		createDirectoryDeep,
		ioutilReadFile,
	)

	if err != nil {
		logging.Get().WithFields(logrus.Fields{
			"fileName":                tmpFileName,
			"targetDirectoryFriendly": targetDirFriendly,
			"pathMap":                 pathMap,
		}).WithError(err).Warn("Error with UnzipArchive")
		return err
	}

	logging.Get().Debugf("%v results written to %s\n", color.Magenta.Sprint(result.PlanName), color.Cyan.Sprint(targetDirFriendly))

	return nil
}

const (
	MAX_ATTEMPTS = 5
)

func CreateTemporaryFile(planName string, data []byte) (name string, err error) {
	var tmpfile *os.File
	var n int
	rand.Seed(time.Now().Unix())

	for attempts := 0; attempts < MAX_ATTEMPTS; attempts++ {
		if tmpfile, err = os.CreateTemp("", planName); err == nil {
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

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func UnzipArchive(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) (err error) {
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
			if fileReader, err = sanitizeZip(fileReader); err != nil {
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

func sanitizeZip(reader io.ReadCloser) (newReader io.ReadCloser, err error) {
	var b []byte
	if b, err = ioutil.ReadAll(reader); err != nil {
		logging.Get().Warnf("unable to read all: %v", err)
		return
	}
	// defer reader.Close()

	if b, err = ReSortJson(b); err != nil {
		logging.Get().Warnf("unable to re-sort: %v", err)
		return
	}

	newReader = ioutil.NopCloser(bytes.NewBuffer(b))

	return
}

func CreateDirectoryDeep(path string, fileMode os.FileMode) (err error) {
	if _, err = os.Stat(path); err != nil {
		logging.Get().Tracef("Creating intermediate directory: %v", color.Cyan.Sprint(path))
		err = os.MkdirAll(path, fileMode)
	}
	return
}

type FileCreator func(fileReader io.ReadCloser, path string) error
type DirectoryCreator func(path string, fileMode os.FileMode) error
type FileReader func(path string) ([]byte, error)

func CopyToFile(fileReader io.ReadCloser, path string) (err error) {
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

	return
}

func CombineJSON(newFileReader io.ReadCloser, existingFileReader FileReader, path string) (rc io.ReadCloser, err error) {
	fields := logrus.Fields{
		"function": "CombineJSON",
	}

	logging.WithFields(fields).Tracef("Augmenting File with more JSON Data: %v\n", color.Magenta.Sprint(path))
	existingBytes, err := existingFileReader(path)
	if err != nil {
		logging.Get().Warnf("existingFileReader: %v", err)
		return
	}

	newBytes, err := ioutil.ReadAll(newFileReader)
	if err != nil {
		logging.Get().Warnf("ioutil.ReadAll: %v", err)
		return
	}

	// merge the files together using the json patch library
	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		logging.Get().Warnf("jsonpatch.MergePatch: %v", err)
		return
	}

	// sort all of the keys in the json. custom sort logic.
	// this puts "name" first, then everything alphanumerically
	sorted, err := ReSortJsonIndent(combined, true)
	if err != nil {
		logging.Get().Warnf("ReSortJsonIndent: %v", err)
		return
	}

	var indented bytes.Buffer
	err = json.Indent(&indented, sorted, "", "\t")
	if err != nil {
		logging.Get().Warnf("Indent: %v", err)
		return
	}

	rc = ioutil.NopCloser(bytes.NewReader(indented.Bytes()))

	return
}
