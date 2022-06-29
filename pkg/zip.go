package pkg

import (
	"archive/zip"
	"bytes"
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
			logging.Get().Trace(color.Magenta.Sprint(filePath))
			return
		}

		var relativeFilePath string
		if relativeFilePath, err = filepath.Rel(basePath, filePath); err != nil {
			logging.Get().Tracef("Relative Filepath Error: %v", err)
			return
		}

		encapsulatingDirectory, fileName := filepath.Split(filePath)
		encapsulatingFolder := filepath.Base(encapsulatingDirectory)

		// we only want the immediate directory and the filename for the archive path
		// so we are going to truncate the archive path
		archivePath := path.Join(encapsulatingFolder, fileName)

		if strings.HasPrefix(archivePath, ".") || !filter(relativeFilePath) {
			logging.Get().Tracef(color.Gray.Sprintf("Ignoring: %v", filePath))
			return
		} else {
			logging.Get().Infof("Processing: %v => %v", color.Green.Sprint(filePath), color.Yellow.Sprint(archivePath))
		}

		// spin off a thread archiving the file
		eg.Go(func() error {
			if bytes, err := ioutil.ReadFile(filePath); err != nil {
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
		if zipFileWriter, e := zipWriter.Create(success.FilePath); err != nil {
			err = e
			return
		} else if _, e := zipFileWriter.Write(success.Bytes); e != nil {
			err = e
			return
		}
		logging.Get().Trace(color.Green.Sprintf("Finished Processing %v", success.FilePath))
	}

	zipWriter.Close()
	result, err = ioutil.ReadAll(buffer)

	return
}
