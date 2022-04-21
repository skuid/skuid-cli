package main

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TestFile contains the information needed to create a test file
type TestFile struct {
	Name string
	Body string
}

// CreateTestZip allows for creating temporary in memory zip files
func CreateTestZip(files []TestFile) (io.ReadCloser, error) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	zipWriter := zip.NewWriter(buf)

	for _, file := range files {
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}
		_, err = zipFile.Write([]byte(file.Body))
		if err != nil {
			return nil, err
		}
	}

	// Make sure to check the error on Close.
	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(buf), nil
}

// Archive compresses a file/directory to a writer
//
// If the path ends with a separator, then the contents of the folder at that path
// are at the root level of the archive, otherwise, the root of the archive contains
// the folder as its only item (with contents inside).
//
// If progress is not nil, it is called for each file added to the archive.
func Archive(inFilePath string, writer io.Writer, metadataFilter *Metadata) error {
	return archiveWithFilter(inFilePath, writer, func(relativePath string) bool {
		// if there was a metadata filter, apply it.
		if metadataFilter != nil {
			if !(*metadataFilter).FilterMetadataItem(relativePath) {
				// If our file does not meet our filter criteria, just skip this file
				return false
			}
		}
		return true
	})
}

// ArchivePartial compresses all files in a file/directory matching a relative prefix to a writer
//
// If the path ends with a separator, then the contents of the folder at that path
// are at the root level of the archive, otherwise, the root of the archive contains
// the folder as its only item (with contents inside).
//
// If progress is not nil, it is called for each file added to the archive.
func ArchivePartial(inFilePath string, writer io.Writer, basePrefix string) error {
	return archiveWithFilter(inFilePath, writer, func(relativePath string) bool {
		return strings.HasPrefix(relativePath, basePrefix)
	})
}

func archiveWithFilter(inFilePath string, writer io.Writer, filter func(string) bool) error {
	zipWriter := zip.NewWriter(writer)

	basePath := filepath.Dir(inFilePath)

	err := filepath.Walk(inFilePath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() {
			return err
		}

		relativeFilePath, err := filepath.Rel(basePath, filePath)
		if err != nil {
			return err
		}

		archivePath := path.Join(filepath.SplitList(relativeFilePath)...)

		if strings.HasPrefix(archivePath, ".") {
			return nil
		}

		if !filter(relativeFilePath) {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		zipFileWriter, err := zipWriter.Create(archivePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFileWriter, file)
		return err
	})
	if err != nil {
		return err
	}

	return zipWriter.Close()
}
