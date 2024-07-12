package util_test

import (
	"bytes"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/skuid/skuid-cli/pkg/util/mocks"
	"github.com/skuid/skuid-cli/pkg/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FileUtilTestSuite struct {
	suite.Suite
}

func (suite *FileUtilTestSuite) TestReadFile() {
	t := suite.T()

	testCases := []struct {
		testDescription string
		giveFS          fs.FS
		giveName        string
		wantError       bool
		wantResult      []byte
	}{
		{
			testDescription: "reads file that exists",
			giveFS: &fstest.MapFS{
				"a/b/c/d/e.txt": &fstest.MapFile{Data: []byte("hello")},
			},
			giveName:   "a/b/c/d/e.txt",
			wantError:  false,
			wantResult: []byte("hello"),
		},
		{
			testDescription: "fails when file not exist",
			giveFS:          &fstest.MapFS{},
			giveName:        "a/b/c/d/e.txt",
			wantError:       true,
			wantResult:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			result, err := util.NewFileUtil().ReadFile(tc.giveFS, tc.giveName)
			if tc.wantError {
				// intentionally not checking error message
				assert.NotNil(t, err, "Expected ReadFile error to be not nill, got nil")
			} else {
				require.NoError(t, err, "Expected ReadFile err to be nil, got not nil")
			}
			assert.Equal(t, tc.wantResult, result)

		})
	}
}

func (suite *FileUtilTestSuite) TestWalkDir() {
	t := suite.T()

	testCases := []struct {
		testDescription   string
		giveFiles         testutil.TestFiles
		giveRoot          string
		giveAddExpectRoot bool
		wantError         bool
		wantFiles         []string
	}{
		{
			testDescription: "walks all paths from root",
			giveFiles: testutil.TestFiles{
				"a/b/c/d.txt": "",
			},
			giveRoot:  ".",
			wantError: false,
			wantFiles: []string{
				".",
				"a",
				"a/b",
				"a/b/c",
				"a/b/c/d.txt",
			},
		},
		{
			testDescription: "walks all paths from root subfolder",
			giveFiles: testutil.TestFiles{
				"a/b/c/d.txt": "",
			},
			giveRoot:  "a",
			wantError: false,
			wantFiles: []string{
				"a",
				"a/b",
				"a/b/c",
				"a/b/c/d.txt",
			},
		},
		{
			testDescription: "walks all paths from nested subfolder",
			giveFiles: testutil.TestFiles{
				"a/b/c/d.txt": "",
			},
			giveRoot:  "a/b",
			wantError: false,
			wantFiles: []string{
				"a/b",
				"a/b/c",
				"a/b/c/d.txt",
			},
		},
		{
			testDescription: "fails when path not exist",
			giveFiles: testutil.TestFiles{
				"a/b/c/d.txt": "",
			},
			giveRoot:  "z",
			wantError: true,
			wantFiles: []string{
				"z",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			wdf := mocks.NewWalkDirFunc(t)
			fsys := testutil.CreateFS(tc.giveFiles)
			for _, path := range tc.wantFiles {
				wdf.EXPECT().Execute(path, mock.Anything, mock.Anything).RunAndReturn(func(path string, di fs.DirEntry, err error) error {
					return err
				}).Once()
			}
			err := util.NewFileUtil().WalkDir(fsys, tc.giveRoot, wdf.Execute)
			if tc.wantError {
				// intentionally not checking error message
				assert.NotNil(t, err, "Expected WalkDir error to be not nill, got nil")
			} else {
				require.NoError(t, err, "Expected WalkDir err to be nil, got not nil")
			}
		})
	}
}

func (suite *FileUtilTestSuite) TestNewZipWriter() {
	t := suite.T()

	testCases := []struct {
		testDescription string
		giveWriter      io.Writer
	}{
		{
			testDescription: "returns valid ZipWriter",
			giveWriter:      new(bytes.Buffer),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			result := util.NewFileUtil().NewZipWriter(tc.giveWriter)
			assert.NotNil(t, result, "NewZipWriter expected result not nil, got nil")
		})
	}
}

func (suite *FileUtilTestSuite) TestDirExists() {
	t := suite.T()

	testCases := []struct {
		testDescription string
		giveFS          fs.FS
		givePath        string
		wantError       bool
		wantResult      bool
	}{
		{
			testDescription: "directory exists in root",
			giveFS: &fstest.MapFS{
				"test_dir": &fstest.MapFile{Mode: fs.ModeDir},
			},
			givePath:   "test_dir",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "directory exists in subdir",
			giveFS: &fstest.MapFS{
				"a/b/c/d/e": &fstest.MapFile{Mode: fs.ModeDir},
			},
			givePath:   "a/b/c/d/e",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "directory does not exist in root",
			giveFS:          &fstest.MapFS{},
			givePath:        "test_dir",
			wantError:       false,
			wantResult:      false,
		},
		{
			testDescription: "directory does not exist in subdir",
			giveFS:          &fstest.MapFS{},
			givePath:        "a/b/c/d/e",
			wantError:       false,
			wantResult:      false,
		},
		{
			testDescription: "file not detected as a directory in root",
			giveFS: &fstest.MapFS{
				"abcd": &fstest.MapFile{},
			},
			givePath:   "abcd",
			wantError:  false,
			wantResult: false,
		},
		{
			testDescription: "file not detected as a directory in subdir",
			giveFS: &fstest.MapFS{
				"a/b/c/d/e": &fstest.MapFile{},
			},
			givePath:   "a/b/c/d/e",
			wantError:  false,
			wantResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			result, err := util.NewFileUtil().DirExists(tc.giveFS, tc.givePath)
			if tc.wantError {
				// intentionally not checking error message
				assert.NotNil(t, err, "Expected DirExists error to be not nill, got nil")
			} else {
				require.NoError(t, err, "Expected DirExists err to be nil, got not nil")
			}

			assert.Equal(t, tc.wantResult, result)
		})
	}
}

func (suite *FileUtilTestSuite) TestFileExists() {
	t := suite.T()

	testCases := []struct {
		testDescription string
		giveFS          fs.FS
		givePath        string
		wantError       bool
		wantResult      bool
	}{
		{
			testDescription: "file exists in root",
			giveFS: &fstest.MapFS{
				"test_file": &fstest.MapFile{},
			},
			givePath:   "test_file",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "file exists in subdir",
			giveFS: &fstest.MapFS{
				"a/b/c/d/test_file": &fstest.MapFile{},
			},
			givePath:   "a/b/c/d/test_file",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "file does not exist in root",
			giveFS:          &fstest.MapFS{},
			givePath:        "test_file",
			wantError:       false,
			wantResult:      false,
		},
		{
			testDescription: "file does not exist in subdir",
			giveFS:          &fstest.MapFS{},
			givePath:        "a/b/c/d/test_file",
			wantError:       false,
			wantResult:      false,
		},
		{
			testDescription: "directory not detected as a file in root",
			giveFS: &fstest.MapFS{
				"abcd": &fstest.MapFile{Mode: fs.ModeDir},
			},
			givePath:   "abcd",
			wantError:  false,
			wantResult: false,
		},
		{
			testDescription: "directory not detected as a file in subdir",
			giveFS: &fstest.MapFS{
				"a/b/c/d/e": &fstest.MapFile{Mode: fs.ModeDir},
			},
			givePath:   "a/b/c/d/e",
			wantError:  false,
			wantResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			result, err := util.NewFileUtil().FileExists(tc.giveFS, tc.givePath)
			if tc.wantError {
				// intentionally not checking error message
				assert.NotNil(t, err, "Expected FileExists error to be not nill, got nil")
			} else {
				require.NoError(t, err, "Expected FileExists err to be nil, got not nil")
			}

			assert.Equal(t, tc.wantResult, result)
		})
	}
}

func (suite *FileUtilTestSuite) TestPathExists() {
	t := suite.T()

	testCases := []struct {
		testDescription string
		giveFS          fs.FS
		givePath        string
		wantError       bool
		wantResult      bool
	}{
		{
			testDescription: "file path exists in root",
			giveFS: &fstest.MapFS{
				"test_file": &fstest.MapFile{},
			},
			givePath:   "test_file",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "dir path exists in root",
			giveFS: &fstest.MapFS{
				"test_dir": &fstest.MapFile{Mode: fs.ModeDir},
			},
			givePath:   "test_dir",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "file exists in subdir",
			giveFS: &fstest.MapFS{
				"a/b/c/d/test_file": &fstest.MapFile{},
			},
			givePath:   "a/b/c/d/test_file",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "dir exists in subdir",
			giveFS: &fstest.MapFS{
				"a/b/c/d/test_dir": &fstest.MapFile{Mode: fs.ModeDir},
			},
			givePath:   "a/b/c/d/test_dir",
			wantError:  false,
			wantResult: true,
		},
		{
			testDescription: "path does not exist in root",
			giveFS:          &fstest.MapFS{},
			givePath:        "abcd",
			wantError:       false,
			wantResult:      false,
		},
		{
			testDescription: "path does not exist in subdir",
			giveFS:          &fstest.MapFS{},
			givePath:        "a/b/c/d/abcd",
			wantError:       false,
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			result, err := util.NewFileUtil().PathExists(tc.giveFS, tc.givePath)
			if tc.wantError {
				// intentionally not checking error message
				assert.NotNil(t, err, "Expected PathExists error to be not nill, got nil")
			} else {
				require.NoError(t, err, "Expected PathExists err to be nil, got not nil")
			}

			assert.Equal(t, tc.wantResult, result)
		})
	}
}

func TestFileUtilTestSuite(t *testing.T) {
	suite.Run(t, new(FileUtilTestSuite))
}
