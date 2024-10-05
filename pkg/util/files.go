package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/skuid/skuid-cli/pkg/logging"
)

const (
	MAX_ATTEMPTS = 2
)

// ZipFile used for testing
type ZipFile interface {
	io.Writer
}

// WalkDirFunc used for testing
type WalkDirFunc func(path string, d fs.DirEntry, err error) error

type ZipWriter interface {
	Create(name string) (io.Writer, error)
	Close() error
}

type FileUtil interface {
	ReadFile(fsys fs.FS, name string) ([]byte, error)
	WalkDir(fsys fs.FS, root string, fn fs.WalkDirFunc) error
	NewZipWriter(w io.Writer) ZipWriter
	DirExists(fsys fs.FS, path string) (bool, error)
	FileExists(fsys fs.FS, path string) (bool, error)
	PathExists(fsys fs.FS, path string) (bool, error)
}

type fileUtil struct{}

func (z *fileUtil) ReadFile(fsys fs.FS, name string) ([]byte, error) {
	return fs.ReadFile(fsys, name)
}

func (z *fileUtil) WalkDir(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(fsys, root, fn)
}

func (z *fileUtil) NewZipWriter(w io.Writer) ZipWriter {
	return zip.NewWriter(w)
}

// Determines if the path specified exists and is a directory
// returns:
//
//	true, nil if exists and its a directory
//	false, nil if does not exist or if exists but its not a directory
//	false, error if error encountered evaluating - note that the path may exist and it may be a directory or a file
func (z *fileUtil) DirExists(fsys fs.FS, path string) (bool, error) {
	exists, fi, err := pathExists(fsys, path)
	if err == nil {
		return exists && fi.IsDir(), nil
	}

	return exists, err
}

// Determines if the path specified exists and is a file
// returns:
//
//	true, nil if exists and its a file
//	false, nil if does not exist or if exists but its not a file
//	false, error if error encountered evaluating - note that the path may exist and it may be a directory or a file
func (z *fileUtil) FileExists(fsys fs.FS, path string) (bool, error) {
	exists, fi, err := pathExists(fsys, path)
	if err == nil {
		return exists && !fi.IsDir(), nil
	}

	return exists, err
}

// Determines if the path specified exists
// returns:
//
//	true, nil if exists
//	false, nil if does not exist
//	false, error if error encountered evaluating - note that the path may exist
func (z *fileUtil) PathExists(fsys fs.FS, path string) (bool, error) {
	exists, _, err := pathExists(fsys, path)
	if err == nil {
		return exists, nil
	}

	return exists, err
}

func NewFileUtil() FileUtil {
	return new(fileUtil)
}

type FileCreator func(fileReader io.ReadCloser, path string) error

type DirectoryCreator func(path string, fileMode os.FileMode) error

type FileReader func(path string) ([]byte, error)

var cleanFileNameRE = regexp.MustCompile(`[^a-zA-Z0-9.-_]`)
var replaceFileNameRE = regexp.MustCompile(`[_]+`)

func CleanFileName(name string) string {
	return strings.Trim(replaceFileNameRE.ReplaceAllString(cleanFileNameRE.ReplaceAllString(name, "_"), "_"), "_")
}

func CreateTemporaryFile(name string, ext string, data []byte) (string, error) {
	pattern := "skuid-cli-" + CleanFileName(name) + "-*"
	if ext != "" {
		pattern += ext
	}
	fields := logging.Fields{
		"name":    name,
		"ext":     ext,
		"pattern": pattern,
		"dataLen": len(data),
	}
	logger := logging.WithName("util.CreateTemporaryFile", fields)
	logger.Tracef("Creating temp file with pattern %v", logging.QuoteText(pattern))

	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("could not create temp file with pattern %v: %w", logging.QuoteText(pattern), err)
	}
	defer tmpFile.Close()

	logger.Tracef("Created temp file %v", logging.QuoteText(tmpFile.Name()))
	if len(data) > 0 {
		if n, err := tmpFile.Write(data); err != nil {
			return "", fmt.Errorf("unable to write to temp file %v: %w", logging.QuoteText(tmpFile.Name()), err)
		} else if n == 0 {
			return "", fmt.Errorf("did not write anything to temp file %v: %w", logging.QuoteText(tmpFile.Name()), err)
		}
	}

	return tmpFile.Name(), nil
}

func CreateDirectoryDeep(path string, fileMode os.FileMode) error {
	logging.WithName("util.CreateDirectoryDeep", logging.Fields{"path": path}).Tracef("Ensuring directory exists at %v", logging.ColorResource.QuoteText(path))
	if err := os.MkdirAll(path, fileMode); err != nil {
		return fmt.Errorf("unable to create directory for path %v: %w", logging.QuoteText(path), err)
	}

	return nil
}

func CopyToFile(fileReader io.ReadCloser, path string) error {
	logging.WithName("util.CopyToFile", logging.Fields{"path": path}).Tracef("Creating and writing to file %v", logging.ColorResource.QuoteText(path))
	targetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file %v: %w", logging.QuoteText(path), err)
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, fileReader); err != nil {
		return fmt.Errorf("unable to copy file %v: %w", logging.QuoteText(path), err)
	}

	return nil
}

func pathExists(fsys fs.FS, path string) (bool, fs.FileInfo, error) {
	fi, err := fs.Stat(fsys, path)
	if err == nil {
		return true, fi, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil, nil
	}

	return false, nil, err
}
