package pkg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	"golang.org/x/sync/errgroup"

	"github.com/skuid/tides/pkg/logging"
)

// Archive compresses a file/directory to a writer
func Archive(inFilePath string, filter *NlxMetadata) (result []byte, err error) {
	return ArchiveWithFilterFunc(inFilePath, func(relativePath string) bool {
		if filter != nil {
			return !filter.FilterItem(relativePath)
		}
		return true
	})
}

// ArchivePartial compresses all files in a file/directory matching a relative prefix to a writer
func ArchivePartial(inFilePath string, basePrefix string) ([]byte, error) {
	return ArchiveWithFilterFunc(inFilePath, func(relativePath string) bool {
		return strings.HasPrefix(relativePath, basePrefix)
	})
}

type archiveSuccess struct {
	Bytes    []byte
	FilePath string
}

func ArchiveWithFilterFunc(inFilePath string, filter func(string) bool) (result []byte, err error) {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)
	basePath := filepath.Dir(inFilePath)

	// halves the time of archival
	eg := &errgroup.Group{}
	ch := make(chan archiveSuccess)

	if err = filepath.Walk(inFilePath, func(filePath string, fileInfo os.FileInfo, e error) (err error) {
		if e != nil {
			return e
		}

		if fileInfo.IsDir() {
			logging.Get().Debugf("Zipping: %v", color.Cyan.Sprint(filePath))
			return
		}

		var relativeFilePath string
		if relativeFilePath, err = filepath.Rel(basePath, filePath); err != nil {
			logging.Get().Warnf("Relative Filepath Error: %v", err)
			return
		}

		encapsulatingDirectory, fileName := filepath.Split(filePath)
		encapsulatingFolder := filepath.Base(encapsulatingDirectory)

		// we only want the immediate directory and the filename for the archive path
		// so we are going to truncate the archive path
		archivePath := path.Join(encapsulatingFolder, fileName)

		if strings.HasPrefix(archivePath, "..") {
			fmt.Println("DOT")
		}
		fmt.Println("======================")
		fmt.Println(encapsulatingFolder)
		if (strings.HasPrefix(archivePath, ".") && strings.HasSuffix(archivePath, ".")) || !filter(relativeFilePath) {
			// todo: fix this; it's not properly filtering off of the low level directory and the filename
			logging.Get().Debugf(color.Gray.Sprintf("Ignoring: %v", filePath))
			fmt.Println(filter(relativeFilePath))
			fmt.Println(strings.HasPrefix(archivePath, "."))
			fmt.Println(archivePath)
			fmt.Println("======================")
			return
		}
		fmt.Println("======================")

		// spin off a thread archiving the file
		eg.Go(func() error {
			logging.Get().Tracef("Processing: %v => %v", color.Green.Sprint(filePath), color.Yellow.Sprint(archivePath))
			if bytes, err := ioutil.ReadFile(filePath); err != nil {
				logging.Get().Warnf("Error Processing %v: %v", filePath, err)
				return err
			} else {
				ch <- archiveSuccess{
					Bytes:    bytes,
					FilePath: archivePath,
				}
			}
			return nil
		})

		return err
	}); err != nil {
		return
	}

	go func() {
		err = eg.Wait()
		if err != nil {
			logging.Get().WithError(err).Fatal("failed during ArchiveWithFilterFunc")
		}
		close(ch)
	}()

	for success := range ch {
		logging.Get().Tracef("Finished Processing %v", color.Green.Sprint(success.FilePath))
		if zipFileWriter, e := zipWriter.Create(success.FilePath); err != nil {
			err = e
			logging.Get().Warnf("Error processing %v: %v", success.FilePath, err)
			return
		} else if _, e := zipFileWriter.Write(success.Bytes); e != nil {
			logging.Get().Warnf("Error writing %v: %v", success.FilePath, err)
			err = e
			return
		}
	}

	zipWriter.Close()
	result, err = ioutil.ReadAll(buffer)

	return
}
