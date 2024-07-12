package testutil

import (
	"io"
	"io/fs"
	"testing"

	"github.com/skuid/skuid-cli/pkg/util"
	utilmocks "github.com/skuid/skuid-cli/pkg/util/mocks"
	"github.com/stretchr/testify/mock"
)

/*
NOTE - Mock Builders contain "Without*" methods because "github.com/stretchr/testify/mock" does not contain a reliable
way to clear/reset expectations.  It does offer the "Unset" API for a "Call", however it has several issues and could
result in unexpected behavior if used.  In order to support a "default" Mock via the Builder but override certain
expectations, the Builders "Without*" methods should be called.  This will cause the builder to not set any expectations
for the specified method so that they can then be setup within a test.  If/When Issue #1621 (see link below) is addressed,
the "Without*" methods can be eliminated and then each test can unset/clear/reset individual expectations.

See the following issues related to Unset
- https://github.com/stretchr/testify/issues/1621
- https://github.com/stretchr/testify/issues/1599
- https://github.com/stretchr/testify/issues/1542

See the following PRs for activity on Unset
- https://github.com/stretchr/testify/pull/1622
- https://github.com/stretchr/testify/pull/1605
- https://github.com/stretchr/testify/pull/1572
- https://github.com/stretchr/testify/pull/1511
- https://github.com/stretchr/testify/pull/1425
- https://github.com/stretchr/testify/pull/1239
*/

// Default behavior will create a spy on the real implementation of util.FileUtil
type FileUtilBuilder struct {
	skipReadFile     bool
	skipNewZipWriter bool
	newZipWriter     util.ZipWriter
	skipWalkDir      bool
	skipDirExists    bool
}

func (b *FileUtilBuilder) WithoutReadFile() *FileUtilBuilder {
	b.skipReadFile = true
	return b
}

func (b *FileUtilBuilder) WithoutNewZipWriter() *FileUtilBuilder {
	b.skipNewZipWriter = true
	return b
}

func (b *FileUtilBuilder) SetZipWriter(zipWriter util.ZipWriter) *FileUtilBuilder {
	b.newZipWriter = zipWriter
	return b
}

func (b *FileUtilBuilder) WithoutWalkDir() *FileUtilBuilder {
	b.skipWalkDir = true
	return b
}

func (b *FileUtilBuilder) WithoutDirExists() *FileUtilBuilder {
	b.skipDirExists = true
	return b
}

func (b *FileUtilBuilder) Build(t *testing.T) *utilmocks.FileUtil {
	mockFileUtil := utilmocks.NewFileUtil(t)
	realFileUtil := util.NewFileUtil()

	if !b.skipReadFile {
		mockFileUtil.EXPECT().ReadFile(mock.Anything, mock.Anything).RunAndReturn(func(fsys fs.FS, name string) ([]byte, error) {
			return realFileUtil.ReadFile(fsys, name)
		}).Maybe()
	}

	if !b.skipNewZipWriter {
		var zw = b.newZipWriter
		mockFileUtil.EXPECT().NewZipWriter(mock.Anything).RunAndReturn(func(w io.Writer) util.ZipWriter {
			if zw != nil {
				return zw
			} else {
				return realFileUtil.NewZipWriter(w)
			}
		}).Maybe()
	}

	if !b.skipWalkDir {
		mockFileUtil.EXPECT().WalkDir(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
			return realFileUtil.WalkDir(fsys, root, fn)
		}).Maybe()
	}

	if !b.skipDirExists {
		mockFileUtil.EXPECT().DirExists(mock.Anything, mock.Anything).RunAndReturn(func(fsys fs.FS, path string) (bool, error) {
			return realFileUtil.DirExists(fsys, path)
		})
	}

	return mockFileUtil
}

func NewFileUtilBuilder() *FileUtilBuilder {
	return &FileUtilBuilder{}
}
