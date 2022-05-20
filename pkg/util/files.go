package util

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
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
	logging.TraceF("Working Directory: %v", wd)

	if strings.Contains(relative, wd) {
		logging.TraceF("Absolute path: %v", relative)
		return relative
	} else {
		logging.TraceLn("Relative Path")
	}

	absolute, _ = filepath.Abs(filepath.Join(wd, relative))
	logging.TraceF("Target Directory: %v", absolute)
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

// DeleteDirectories deletes a list of directories (os.RemoveAll) in a target directory
func DeleteDirectories(targetDirectory string, directories []string) (err error) {
	// Remove all of our metadata directories so we get a clean slate.
	// We may want to improve this later when we do partial retrieves so that
	// we don't clear out the whole directory every time we retrieve.
	for _, dirName := range directories {
		dirPath := filepath.Join(targetDirectory, dirName)

		logging.TraceLn("Deleting Directory: " + color.Red.Sprint(dirPath))

		if err = os.RemoveAll(dirPath); err != nil {
			return
		}
	}
	return
}

func WriteResultsToDisk(targetDirectory string, results [][]byte, noZip bool) (err error) {
	return WriteResultsToDiskInjection(targetDirectory, results, noZip, CopyToFile, CreateDirectoryDeep, ioutil.ReadFile)
}

func WriteResultsToDiskInjection(targetDirectory string, results [][]byte, noZip bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) (err error) {

	// unzip the archive into the output directory
	targetDirFriendly, err := SanitizePath(targetDirectory)
	if err != nil {
		return err
	}

	logging.TraceF("Writing results to %v\n", color.Cyan.Sprint(targetDirFriendly))

	// Store a map of paths that we've already encountered. We'll use this
	// to determine if we need to modify a file or overwrite it.
	pathMap := map[string]bool{}

	for _, result := range results {
		tmpFileName, err := CreateTemporaryFile(result)
		if err != nil {
			return err
		}
		// schedule cleanup of temp file
		defer os.Remove(tmpFileName)

		if noZip {
			logging.TraceF("Moving Temporary File: %v => %v", tmpFileName, targetDirectory)
			err = MoveTemporaryFile(tmpFileName, targetDirectory, pathMap, fileCreator, directoryCreator, existingFileReader)
			if err != nil {
				return err
			}
			continue
		}

		// unzip the contents of our temp zip file
		err = UnzipArchive(tmpFileName, targetDirectory, pathMap, fileCreator, directoryCreator, existingFileReader)
		if err != nil {
			return err
		}
	}

	logging.Printf("Results written to %s\n", color.Cyan.Sprint(targetDirFriendly))

	return nil
}

func CreateTemporaryFile(data []byte) (name string, err error) {
	var tmpfile *os.File
	if tmpfile, err = ioutil.TempFile("", "skuid"); err != nil {
		return
	} else if _, err = tmpfile.Write(data); err != nil {
		return
	} else {
		name = tmpfile.Name()
	}

	logging.TraceLn(color.Yellow.Sprintf("Created Temp File: %v", name))
	return
}

func MoveTemporaryFile(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) (err error) {
	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err = directoryCreator(targetLocation, 0755); err != nil {
			return
		}
	}

	// open the file (to copy later)
	var fi *os.File
	if fi, err = os.Open(sourceFileLocation); err != nil {
		return
	}
	defer fi.Close()

	var fstat os.FileInfo
	if fstat, err = fi.Stat(); err != nil {
		return
	}

	fileReader := ioutil.NopCloser(fi)
	path := filepath.Join(targetLocation, filepath.FromSlash(fi.Name()))

	_, fileAlreadyWritten := pathMap[path]

	if !fileAlreadyWritten {
		pathMap[path] = true
	}

	if fstat.IsDir() {
		return directoryCreator(path, fstat.Mode())
	}

	if fileAlreadyWritten {
		logging.TraceF("Augmenting existing file with more data: %s\n", color.Magenta.Sprint(fi.Name()))
		if filepath.Ext(sourceFileLocation) == ".json" {
			if fileReader, err = CombineJSON(fileReader, existingFileReader, path); err != nil {
				return
			}
		}
	}

	logging.TraceLn("Moving file: " + color.Green.Sprint(fi.Name()))

	if err = fileCreator(fileReader, path); err != nil {
		return
	}

	return
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func UnzipArchive(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) (err error) {
	logging.TraceF("Unzipping Archive: %v => %v", sourceFileLocation, targetLocation)
	var reader *zip.ReadCloser
	if reader, err = zip.OpenReader(sourceFileLocation); err != nil {
		return
	}

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err = directoryCreator(targetLocation, 0755); err != nil {
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
	logging.TraceF("Extracting from Zip: %v", fullPath)

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
		return directoryCreator(fullPath, file.Mode())
	}

	var fileReader io.ReadCloser
	if fileReader, err = file.Open(); err != nil {
		return
	}
	defer fileReader.Close()

	if filepath.Ext(fullPath) == ".json" {
		if fileReader, err = sanitizeZip(fileReader); err != nil {
			return
		}
	}

	if fileAlreadyWritten {
		logging.TraceF("Augmenting existing file with more data: %s\n", color.Magenta.Sprint(file.Name))
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
		logging.TraceLn("Creating intermediate directory: " + color.Cyan.Sprint(path))
		err = os.MkdirAll(path, fileMode)
	}
	return
}

type FileCreator func(fileReader io.ReadCloser, path string) error
type DirectoryCreator func(path string, fileMode os.FileMode) error
type FileReader func(path string) ([]byte, error)

func CopyToFile(fileReader io.ReadCloser, path string) (err error) {
	fileFlags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	fileMode := os.FileMode(0644)

	var targetFile *os.File
	if targetFile, err = os.OpenFile(path, fileFlags, fileMode); err != nil {
		return
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, fileReader); err != nil {
		return
	}

	return
}

func CombineJSON(newFileReader io.ReadCloser, existingFileReader FileReader, path string) (rc io.ReadCloser, err error) {

	logging.TraceF("Augmenting File with more JSON Data: %v\n", color.Magenta.Sprint(path))
	existingBytes, err := existingFileReader(path)
	if err != nil {
		return
	}

	newBytes, err := ioutil.ReadAll(newFileReader)
	if err != nil {
		return
	}

	// merge the files together using the json patch library
	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		return
	}

	// sort all of the keys in the json. custom sort logic.
	// this puts "name" first, then everything alphanumerically
	sorted, err := ReSortJsonIndent(combined, true)
	if err != nil {
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
