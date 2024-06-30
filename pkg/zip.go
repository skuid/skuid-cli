package pkg

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
	"golang.org/x/sync/errgroup"
)

// Archive compresses a file/directory to a writer
func Archive(inFilePath string, filter *NlxMetadata) (result []byte, err error) {
	result, _, err = ArchiveWithFilterFunc(inFilePath, func(relativePath string) bool {
		if filter != nil {
			keep := filter.FilterItem(relativePath)
			if !keep {
				return false
			}
		}
		return true
	})
	return
}

// ArchiveFiles compresses all files in a directory that are present in the slices files to a writer returning the number of files written
func ArchiveFiles(inFilePath string, files []string) ([]byte, int, error) {
	return ArchiveWithFilterFunc(inFilePath, func(relativePath string) bool {
		// NOTE - As of golang v1.21, slices package includes a Contains method (slices.Contains(files, relativePath))
		// however in order to support >= v1.20, unable to use it.
		// TODO: If/When skuid-cli states an official minimum supported go version and if/when that version
		// is >= v1.21, the slices Contains can be called directly instead of using custom util StringSliceContainsKey
		return util.StringSliceContainsKey(files, relativePath)
	})
}

func ArchiveMetadata(inFilePath string, excludedMetadataDirs []string) (result []byte, err error) {
	hasExcludedDirs := len(excludedMetadataDirs) > 0
	result, _, err = ArchiveWithFilterFunc(inFilePath, func(relativePath string) bool {
		if !hasExcludedDirs {
			return true
		}
		// NOTE - As of golang v1.21, slices package includes a Contains method (slices.Contains(files, relativePath))
		// however in order to support >= v1.20, unable to use it.
		// TODO: If/When skuid-cli states an official minimum supported go version and if/when that version
		// is >= v1.21, the slices Contains can be called directly instead of using custom util StringSliceContainsKey
		metadataType, _ := GetEntityDetails(relativePath)
		return !util.StringSliceContainsKey(excludedMetadataDirs, metadataType)
	})
	return
}

type archiveSuccess struct {
	Bytes    []byte
	FilePath string
}

func ArchiveWithFilterFunc(inFilePath string, filterKeep func(string) bool) (result []byte, fileCount int, err error) {
	if !filepath.IsAbs(inFilePath) {
		err = fmt.Errorf("inFilePath must be an absolute path")
		return
	}

	inFileStat, err := os.Stat(inFilePath)
	if err != nil {
		return nil, 0, err
	}
	if !inFileStat.IsDir() {
		msg := fmt.Sprintf("Requested folder %s is not a directory", inFilePath)
		logging.Get().Warnf(msg)
		return nil, 0, errors.New(msg)
	}

	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)

	// https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline
	g, ctx := errgroup.WithContext(context.Background())
	ch := make(chan archiveSuccess)

	err = filepath.Walk(inFilePath, func(filePath string, fileInfo os.FileInfo, e error) (err error) {
		if e != nil {
			return e
		}

		if fileInfo.IsDir() {
			logging.Get().Debugf("Zipping: %v", color.Cyan.Sprint(filePath))
			return
		}

		// we only want the immediate directory and the filename for the archive path
		var archivePath string
		if archivePath, err = filepath.Rel(inFilePath, filePath); err != nil {
			logging.Get().Warnf("ArchivePath Error: %v", err)
			return
		}

		if strings.HasPrefix(archivePath, ".") {
			logging.Get().Debugf(color.Gray.Sprintf("Ignoring hidden file: %v", filePath))
			return
		}

		if !filterKeep(archivePath) {
			logging.Get().Debugf(color.Gray.Sprintf("Ignoring filtered file: %v", filePath))
			return
		}

		// spin off a thread archiving each file
		g.Go(func() (err error) {
			logging.Get().Tracef("Processing: %v => %v", color.Green.Sprint(filePath), color.Yellow.Sprint(archivePath))
			var fileBytes []byte
			fileBytes, err = os.ReadFile(filePath)
			if err != nil {
				logging.Get().Warnf("Error Processing %v: %v", filePath, err)
				return
			}
			success := archiveSuccess{
				Bytes:    fileBytes,
				FilePath: archivePath,
			}
			select {
			case ch <- success:
			case <-ctx.Done():
				return ctx.Err()
			}
			return
		})
		return
	})

	go func() {
		err = g.Wait()
		close(ch) // after all workers in group are done, we can close channel to begin range
		if err != nil {
			logging.Get().WithError(err).Fatal("failed during ArchiveWithFilterFunc")
		}
	}()

	fileCount = 0
	for success := range ch {
		var zipFileWriter io.Writer
		logging.Get().Tracef("Finished Processing %v", color.Green.Sprint(success.FilePath))
		zipFileWriter, err = zipWriter.Create(success.FilePath)
		if err != nil {
			logging.Get().Errorf("Error processing %v: %v", success.FilePath, err)
			return
		}
		_, err = zipFileWriter.Write(success.Bytes)
		if err != nil {
			logging.Get().Errorf("Error writing %v: %v", success.FilePath, err)
			return
		}
		fileCount++
	}

	_ = zipWriter.Close()
	result, err = io.ReadAll(buffer)

	return
}
