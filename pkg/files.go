package pkg

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
	"github.com/skuid/tides/pkg/util"
)

func getFriendlyURL(targetDir string) (string, error) {
	if targetDir == "" {
		friendlyResult, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}
		return friendlyResult, nil
	}
	return targetDir, nil

}

type WriteSettings struct {
	TargetDir          string
	Results            [][]byte
	FileCreator        FileCreator
	DirectoryCreator   DirectoryCreator
	ExistingFileReader FileReader
	NoZip              bool
}

func WriteResultsToDisk2(ws WriteSettings) error {
	return WriteResultsToDisk(ws.TargetDir, ws.Results, ws.FileCreator, ws.DirectoryCreator, ws.ExistingFileReader, ws.NoZip)
}

func WriteResultsToDisk(targetDir string, results [][]byte, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader, noZip bool) error {

	// unzip the archive into the output directory
	targetDirFriendly, err := getFriendlyURL(targetDir)
	if err != nil {
		return err
	}

	logging.Printf("Writing results to %v\n", color.Cyan.Sprint(targetDirFriendly))

	// Remove all of our metadata directories so we get a clean slate.
	// We may want to improve this later when we do partial retrieves so that
	// we don't clear out the whole directory every time we retrieve.
	for _, dirName := range GetMetadataTypeDirNames() {
		dirPath := filepath.Join(targetDir, dirName)

		logging.DebugLn("Deleting Directory: " + color.Red.Sprint(dirPath))

		os.RemoveAll(dirPath)
	}

	// Store a map of paths that we've already encountered. We'll use this
	// to determine if we need to modify a file or overwrite it.
	pathMap := map[string]bool{}

	for _, result := range results {

		tmpFileName, err := createTemporaryFile(result)
		if err != nil {
			return err
		}
		// schedule cleanup of temp file
		defer os.Remove(tmpFileName)

		if noZip {
			err = moveTempFile(tmpFileName, targetDir, pathMap, fileCreator, directoryCreator, existingFileReader)
			if err != nil {
				return err
			}
			continue
		}

		// unzip the contents of our temp zip file
		err = unzip(tmpFileName, targetDir, pathMap, fileCreator, directoryCreator, existingFileReader)
		if err != nil {
			return err
		}
	}

	logging.Printf("Results written to %s\n", color.Cyan.Sprint(targetDirFriendly))

	return nil
}

func createTemporaryFile(data []byte) (name string, err error) {
	tmpfile, err := ioutil.TempFile("", "skuid")
	if err != nil {
		return "", err
	}
	// write to our new file
	if _, err := tmpfile.Write(data); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

func moveTempFile(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {
	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err := directoryCreator(targetLocation, 0755); err != nil {
			return err
		}
	}
	fi, err := os.Open(sourceFileLocation)
	if err != nil {
		return err
	}
	defer fi.Close()
	fstat, err := fi.Stat()
	if err != nil {
		return err
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

		logging.DebugF("Augmenting existing file with more data: %s\n", color.Magenta.Sprint(fi.Name()))

		fileReader, err = combineJSONFile(fileReader, existingFileReader, path)
		if err != nil {
			return err
		}
	}

	logging.DebugLn("Creating file: " + color.Green.Sprint(fi.Name()))

	err = fileCreator(fileReader, path)
	if err != nil {
		return err
	}

	return nil
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func unzip(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {

	reader, err := zip.OpenReader(sourceFileLocation)
	if err != nil {
		return err
	}

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err := directoryCreator(targetLocation, 0755); err != nil {
			return err
		}
	}

	for _, file := range reader.File {
		path := filepath.Join(targetLocation, filepath.FromSlash(file.Name))
		// Check to see if we've already written to this file in this retrieve
		_, fileAlreadyWritten := pathMap[path]
		if !fileAlreadyWritten {
			pathMap[path] = true
		}
		err := readFileFromZipAndWriteToFilesystem(file, path, fileAlreadyWritten, fileCreator, directoryCreator, existingFileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func readFileFromZipAndWriteToFilesystem(file *zip.File, fullPath string, fileAlreadyWritten bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {

	// If this file name contains a /, make sure that we create the directory it belongs in
	if pathParts := strings.Split(fullPath, string(filepath.Separator)); len(pathParts) > 0 {
		// Remove the actual file name from the slice,
		// i.e. pages/MyAwesomePage.xml ---> pages
		pathParts = pathParts[:len(pathParts)-1]
		// and then make dirs for all paths up to that point, i.e. pages, apps
		intermediatePath := filepath.Join(strings.Join(pathParts[:], string(filepath.Separator)))
		if intermediatePath != "" {
			err := directoryCreator(intermediatePath, 0755)
			if err != nil {
				return err
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

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	if fileAlreadyWritten {

		logging.DebugF("Augmenting existing file with more data: %s\n", color.Magenta.Sprint(file.Name))

		fileReader, err = combineJSONFile(fileReader, existingFileReader, fullPath)
		if err != nil {
			return err
		}

	}

	logging.DebugLn("Creating file: " + color.Green.Sprint(file.Name))

	err = fileCreator(fileReader, fullPath)
	if err != nil {
		return err
	}

	return nil
}

func CreateDirectory(path string, fileMode os.FileMode) error {
	if _, err := os.Stat(path); err != nil {

		logging.DebugLn("Creating intermediate directory: " + color.Cyan.Sprint(path))

		return os.MkdirAll(path, fileMode)
	}
	return nil
}

type FileCreator func(fileReader io.ReadCloser, path string) error
type DirectoryCreator func(path string, fileMode os.FileMode) error
type FileReader func(path string) ([]byte, error)

func WriteNewFile(fileReader io.ReadCloser, path string) error {
	targetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer targetFile.Close()
	if _, err := io.Copy(targetFile, fileReader); err != nil {
		return err
	}

	return nil
}

func ReadExistingFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func combineJSONFile(newFileReader io.ReadCloser, existingFileReader FileReader, path string) (rc io.ReadCloser, err error) {
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
	sorted, err := util.ReSortJson(combined)
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
