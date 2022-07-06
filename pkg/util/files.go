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
	log := logging.WithFields(logrus.Fields{
		"function": "GetAbsolutePath",
	})
	log.Tracef("Working Directory: %v", wd)

	if strings.Contains(relative, wd) {
		log.Tracef("Absolute path: %v", relative)
		return relative
	} else {
		log.Trace("Relative Path")
	}

	absolute, _ = filepath.Abs(filepath.Join(wd, relative))
	log.Tracef("Target Directory: %v", absolute)
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
	log := logging.WithFields(fields)

	// unzip the archive into the output directory
	targetDirFriendly, err := SanitizePath(targetDirectory)
	if err != nil {
		return err
	}

	log.Tracef("Writing results to %v\n", color.Cyan.Sprint(targetDirFriendly))

	if err := os.MkdirAll(targetDirFriendly, 0777); err != nil {
		log.Tracef("Error making target dir: %v", err.Error())
	}

	// Store a map of paths that we've already encountered. We'll use this
	// to determine if we need to modify a file or overwrite it.
	pathMap := map[string]bool{}

	tmpFileName, err := CreateTemporaryFile(result.PlanName, result.PlanData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"fileName": tmpFileName,
		}).
			WithError(err).
			Error("error creating temporary file")
		return err
	}
	defer os.Remove(tmpFileName)

	// unzip the contents of our temp zip file
	if err = UnzipArchive(
		tmpFileName,
		targetDirectory,
		pathMap,
		copyToFile,
		createDirectoryDeep,
		ioutilReadFile,
	); err != nil {
		log.WithFields(logrus.Fields{
			"fileName":        tmpFileName,
			"targetDirectory": targetDirectory,
			"pathMap":         pathMap,
		}).WithError(err).
			Warn("Error with UnzipArchive")
		return err
	}

	log.Debugf("%v results written to %s\n", color.Magenta.Sprint(result.PlanName), color.Cyan.Sprint(targetDirFriendly))

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
	log := logging.WithFields(fields)
	log.Tracef("Unzipping Archive: %v => %v", color.Green.Sprint(sourceFileLocation), color.Blue.Sprint(targetLocation))
	var reader *zip.ReadCloser
	if reader, err = zip.OpenReader(sourceFileLocation); err != nil {
		log.WithError(err).Warn("unable to unzip archive")
		return
	}

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err = directoryCreator(targetLocation, 0755); err != nil {
			log.Warnf("directoryCreator: %v", err)
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

		if err = readFileFromZipAndWriteToFilesystem(file, path, fileAlreadyWritten, fileCreator, directoryCreator, existingFileReader); err != nil {
			log.Warnf("readFileFromZipAndWriteToFilesystem: %v", err)
			return
		}
	}

	return
}

func sanitizeZip(reader io.ReadCloser) (newReader io.ReadCloser, err error) {
	var b []byte
	if b, err = ioutil.ReadAll(reader); err != nil {
		return
	}
	reader.Close()

	if b, err = ReSortJson(b); err != nil {
		return
	}

	newReader = ioutil.NopCloser(bytes.NewBuffer(b))

	return
}

func readFileFromZipAndWriteToFilesystem(
	file *zip.File,
	fullPath string,
	fileAlreadyWritten bool,
	fileCreator FileCreator,
	directoryCreator DirectoryCreator,
	existingFileReader FileReader,
) (err error) {
	fields := logrus.Fields{
		"func":     "readFileFromZipAndWriteToFilesystem",
		"fullPath": fullPath,
	}
	log := logging.WithFields(fields)
	log.Tracef("Extracting from Zip: %v", color.Blue.Sprint(fullPath))

	// If this file name contains a /, make sure that we create the directory it belongs in
	if pathParts := strings.Split(fullPath, string(filepath.Separator)); len(pathParts) > 0 {
		// Remove the actual file name from the slice,
		// i.e. pages/MyAwesomePage.xml ---> pages
		pathParts = pathParts[:len(pathParts)-1]
		// and then make dirs for all paths up to that point, i.e. pages, apps
		if intermediatePath := filepath.Join(strings.Join(pathParts[:], string(filepath.Separator))); intermediatePath != "" {

			if err = directoryCreator(intermediatePath, 0755); err != nil {
				return
			}
		} else {
			// If we don't have an intermediate path, skip out.
			// Currently Skuid CLI does not create any files in the base directory
			return nil
		}
	}

	if file.FileInfo().IsDir() {
		log.Trace("Creating Directory.")
		return directoryCreator(fullPath, file.Mode())
	}

	var fileReader io.ReadCloser
	if fileReader, err = file.Open(); err != nil {
		return
	}
	defer fileReader.Close()

	if filepath.Ext(fullPath) == ".json" {
		log.Trace("Sanitizing Zip.")
		if fileReader, err = sanitizeZip(fileReader); err != nil {
			log.Tracef("Error: %v", err)
			return
		}
	}

	if fileAlreadyWritten {
		log.Tracef("Augmenting existing file with more data: %s\n", color.Magenta.Sprint(file.Name))
		if fileReader, err = CombineJSON(fileReader, existingFileReader, fullPath); err != nil {
			return
		}
	}

	if err = fileCreator(fileReader, fullPath); err != nil {
		return
	}

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
	log := logging.WithFields(fields)
	log.Tracef("Augmenting File with more JSON Data: %v\n", color.Magenta.Sprint(path))
	existingBytes, err := existingFileReader(path)
	if err != nil {
		log.Tracef("existingFileReader: %v", err)
		return
	}

	newBytes, err := ioutil.ReadAll(newFileReader)
	if err != nil {
		log.Tracef("ioutil.ReadAll: %v")
		return
	}

	// merge the files together using the json patch library
	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		log.Tracef("jsonpatch.MergePatch: %v", err)
		return
	}

	// sort all of the keys in the json. custom sort logic.
	// this puts "name" first, then everything alphanumerically
	sorted, err := ReSortJsonIndent(combined, true)
	if err != nil {
		log.Tracef("ReSortJsonIndent: %v", err)
		return
	}

	var indented bytes.Buffer
	err = json.Indent(&indented, sorted, "", "\t")
	if err != nil {
		return
	}

	rc = ioutil.NopCloser(bytes.NewReader(indented.Bytes()))

	return
}
