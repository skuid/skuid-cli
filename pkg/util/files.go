package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"

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

func CreateTemporaryFile(planName string, data []byte) (name string, err error) {
	var tmpfile *os.File
	var n int
	rand.Seed(time.Now().Unix())

	for attempts := 0; attempts < MAX_ATTEMPTS; attempts++ {
		if tmpfile, err = os.CreateTemp("", strings.ReplaceAll(planName, " ", "-")); err == nil {
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

func CreateDirectoryDeep(path string, fileMode os.FileMode) (err error) {
	if _, err = os.Stat(path); err != nil {
		logging.Get().Tracef("Creating intermediate directory: %v", color.Cyan.Sprint(path))
		err = os.MkdirAll(path, fileMode)
	}
	return
}

func CopyToFile(fileReader io.ReadCloser, path string) (err error) {
	logging.Get().Tracef("%v: %v", color.Yellow.Sprint("Creating File"), path)

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

	logging.Get().Tracef("%v: %v", color.Yellow.Sprint("Copied to File"), path)

	return
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
