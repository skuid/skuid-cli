package pkg_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/skuid/skuid-cli/pkg"
	pkgmocks "github.com/skuid/skuid-cli/pkg/mocks"
	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/skuid/skuid-cli/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const validJSON = `{
	"good": "json"
}`
const validXML = `
	<skuid__page unsavedchangeswarning="yes">
	</skuid__page>
`
const validJS = `function testme() { console.log('hello'); }`
const validTXT = `hello there`
const validPNG = `
	data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACEAAAAgCAYAAAHrvyFcAAACTklEQVR4nNRXTShEURR+S0tLS0tLS0tLS0sLCzZSVjZSFmykbGajlM0sZSFlIyGSSEhKFEKiSZNG8hdp+I7Ode575753n5mXcerMe/fec793fu+5E6we3ZWHZ8/KIDwxDvASdC7SJFMwNn9Ok8wkRCtiksaMx5hBBAsTraM7xLRVYhCzCuDe/BFta8/t21+XQoX7VxLC0xJ6efuwob+4ZWTb2kh2hG2RNhm7pG3SRmOn8YLybglo5BRIRIgIIFyqwNzurRk09K1Zi7mFy29fYCHsh/qelR8/uJy0cFCk59ZpKfppV+yS3FI1EGS35IvisxsEyQmq616yxqomVDcxINpYcn79OlpNkcoSTy1CtMa2NvavO8E05vBaIEyUGDGbtbIw9cIMH8VFJiwfuL4ik4sTTouiEwDvOE1wsrDjUgFoVDGARqkAkM7hFM/OBBlOfq9RAOS2nEBJD86c+ANMbxXUvOfsYwBNhhonxTKmgI5vHgnMWfrh3ubLADYAXDy+m9ErTCaGk0U7gaSD0XglqQdtpfXB5IudWgk6mEUkXXOZKoEgNg1s0JnOearNZapETYQjC/qfnmgb37PizhdgrnG+aWSqBD4ilcA6E+Z/pYQsrz9TAuWVJNgxcWCA8ZeimkpQM4Bw19RhorDkOCUwxiWB8ycOx9xU8YOzHIeNrxJQmhNRU8KH0V84qa3mI/+9pWFYNLl8VW4e2vSSh8GyiUU64P7lA3W63yjjw/J24VSCCXH1tSyJ0VNgnIucSkgqPb3T/4Y07salTB5kcfQJAAD//xMfR9wAAAAGSURBVAMACOfMIINaPS4AAAAASUVORK5CYII=
`

type ArchiveTestDetails struct {
	giveFS            fs.FS
	giveFileUtil      util.FileUtil
	giveArchiveFilter pkg.ArchiveFilter
	wantError         error
	wantResult        testutil.TestFiles
}

type ArchiveTestSuite struct {
	suite.Suite
}

func (suite *ArchiveTestSuite) TestNoMetadataDirs() {
	t := suite.T()
	fsysNoDirs := &fstest.MapFS{}
	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsysNoDirs,
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: nil,
		wantError:         pkg.ErrArchiveNoMetadataTypeDirs,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestNoMetadataFiles() {
	t := suite.T()
	fsysNoFiles := &fstest.MapFS{
		"pages": {Mode: fs.ModeDir},
		"files": {Mode: fs.ModeDir},
	}
	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsysNoFiles,
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: nil,
		wantError:         pkg.ErrArchiveNoFiles,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestCreatesArchive() {
	t := suite.T()
	validFiles := testutil.TestFiles{
		"apps/my_app.json":                 validJSON,
		"pages/page1.xml":                  validXML,
		"pages/page1.json":                 validJSON,
		"files/file1.js":                   validJS,
		"files/file1.js.skuid.json":        validJSON,
		"files/file2.txt":                  validTXT,
		"files/file2.txt.skuid.json":       validJSON,
		"site/site.json":                   validJSON,
		"site/logo/my_logo.png":            validPNG,
		"site/logo/my_logo.png.skuid_json": validJSON,
	}
	fsys := testutil.CreateFS(validFiles)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: nil,
		wantError:         nil,
		wantResult:        validFiles,
	})
}

func (suite *ArchiveTestSuite) TestFilterCalledOnValidFilenamesForValidMetadataTypes() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	filter := pkgmocks.NewArchiveFilter(t)
	for path := range validMetadataTypeFiles {
		filter.EXPECT().Execute(path).Return(true)
	}

	runArchiveTest(suite.T(), ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: filter.Execute,
		wantError:         nil,
		wantResult:        validMetadataTypeFiles,
	})

	// note that since we did not set an expectation for anything other than expected paths, if
	// Execute is called for something else, the test will FailNow by testify since it doesn't
	// have an expectation configured for anything but the paths we expect. testify doesn't
	// have explicit support for a "Times(0)/Never" so we can't add a mock.Anything
	// expectation so the below is for clarity and just in case testify changes behavior unexpectedly.
	filter.AssertNumberOfCalls(t, "Execute", len(validMetadataTypeFiles))
}

// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
func (suite *ArchiveTestSuite) TestFilterCalledOnInvalidFilenamesForValidMetadataTypes() {
	t := suite.T()
	invalidFiles := testutil.TestFiles{
		"pages/page1":           "",
		"pages/page1.txt":       "",
		"pages/page1.jpg":       "",
		"apps/my_app":           "",
		"apps/my_app.xml":       "",
		"datasources/my_ds":     "",
		"datasources/my_ds.png": "",
	}
	fsys := testutil.CreateFS(invalidFiles)
	filter := pkgmocks.NewArchiveFilter(t)
	for path := range invalidFiles {
		filter.EXPECT().Execute(path).Return(true).Once()
	}
	runArchiveTest(suite.T(), ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: filter.Execute,
		wantError:         nil,
		wantResult:        invalidFiles,
	})

	// note that since we did not set an expectation for anything other than expected paths, if
	// Execute is called for something else, the test will FailNow by testify since it doesn't
	// have an expectation configured for anything but the paths we expect. testify doesn't
	// have explicit support for a "Times(0)/Never" so we can't add a mock.Anything
	// expectation so the below is for clarity and just in case testify changes behavior unexpectedly.
	filter.AssertNumberOfCalls(t, "Execute", len(invalidFiles))
}

func (suite *ArchiveTestSuite) TestFilterNotCalledOnFilenamesForInvalidMetadataTypes() {
	t := suite.T()
	invalidMetadataTypeFiles := createInvalidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(invalidMetadataTypeFiles)
	filter := pkgmocks.NewArchiveFilter(t)

	runArchiveTest(suite.T(), ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: filter.Execute,
		wantError:         pkg.ErrArchiveNoMetadataTypeDirs,
		wantResult:        nil,
	})

	// note that since we did not set an expectation, if Execute is called, the test will FailNow
	// by testify since it doesn't have an expectation configured. testify doesn't have explicit
	// support for a "Times(0)/Never" so we can't add a mock.Anything expectation so the below is
	// for clarity and just in case testify changes behavior unexpectedly.
	filter.AssertNumberOfCalls(t, "Execute", 0)
}

func (suite *ArchiveTestSuite) TestArchiveFilterEliminatesAllFiles() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            testutil.CreateFS(validMetadataTypeFiles),
		giveFileUtil:      util.NewFileUtil(),
		giveArchiveFilter: func(path string) bool { return false },
		wantError:         pkg.ErrArchiveNoFiles,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestZipWriterCreateError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("Create failed")
	mockZipWriter := testutil.NewZipWriterBuilder().WithoutCreate().Build(t)
	mockZipWriter.EXPECT().Create(mock.Anything).Return(nil, expectedErr)
	mockFileUtil := testutil.NewFileUtilBuilder().SetZipWriter(mockZipWriter).Build(t)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestZipWriterCloseError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("Close failed")
	mockZipWriter := testutil.NewZipWriterBuilder().WithoutClose().Build(t)
	mockZipWriter.EXPECT().Close().Return(expectedErr)
	mockFileUtil := testutil.NewFileUtilBuilder().SetZipWriter(mockZipWriter).Build(t)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestReadFileError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("ReadFile failed")
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutReadFile().Build(t)
	mockFileUtil.EXPECT().ReadFile(fsys, mock.Anything).Return(nil, expectedErr)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestWalkDirError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("WalkDir failed")
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutWalkDir().Build(t)
	mockFileUtil.EXPECT().WalkDir(mock.Anything, mock.Anything, mock.Anything).Return(expectedErr)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestWalkDirFuncError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("WalkDirFuncParam has error")
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutWalkDir().Build(t)
	mockFileUtil.EXPECT().WalkDir(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
		return fs.WalkDir(fsys, root, func(path string, di fs.DirEntry, e error) error {
			return fn(path, di, expectedErr)
		})
	})

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
func (suite *ArchiveTestSuite) TestWalkDirOnlyMetadataTypeDirsThatExist() {
	t := suite.T()

	validMetadataTypeFiles := testutil.TestFiles{
		"pages/page1.xml": validXML,
	}
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutWalkDir().Build(t)
	mockFileUtil.EXPECT().WalkDir(fsys, "pages", mock.Anything).RunAndReturn(func(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
		return fs.WalkDir(fsys, root, fn)
	}).Once()

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         nil,
		wantResult:        validMetadataTypeFiles,
	})

	// note that since we did not set an expectation for anything other than expected directories, if
	// WalkDir is called for something else, the test will FailNow by testify since it doesn't
	// have an expectation configured for anything but the directories we expect. testify doesn't
	// have explicit support for a "Times(0)/Never" so we can't add a mock.Anything
	// expectation so the below is for clarity and just in case testify changes behavior unexpectedly.
	mockFileUtil.AssertNumberOfCalls(t, "WalkDir", 1)
}

// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
func (suite *ArchiveTestSuite) TestWalkDirNotCalledOnInvalidMetadataTypeDirs() {
	t := suite.T()
	invalidMetadataTypeFiles := testutil.TestFiles{
		"pages":                   "", // NOTE - This is a File not a directory
		"apps":                    "", // NOTE - This is a File not a directory
		"rootfile.json":           "",
		"pagesfake/page1.xml":     "",
		"invalidfolder/app.json":  "",
		"anotherfolder/test.json": "",
		"site2/site.json":         "",
	}

	fsys := testutil.CreateFS(invalidMetadataTypeFiles)
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutWalkDir().Build(t)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         pkg.ErrArchiveNoMetadataTypeDirs,
		wantResult:        nil,
	})

	// note that since we did not set an expectation, if WalkDir is called, the test will FailNow
	// by testify since it doesn't have an expectation configured. testify doesn't have explicit
	// support for a "Times(0)/Never" so we can't add a mock.Anything expectation so the below is
	// for clarity and just in case testify changes behavior unexpectedly.
	mockFileUtil.AssertNumberOfCalls(t, "WalkDir", 0)
}

func (suite *ArchiveTestSuite) TestDirExistsError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("DirExists failed")
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutDirExists().Build(t)
	mockFileUtil.EXPECT().DirExists(fsys, mock.Anything).Return(false, expectedErr)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

func (suite *ArchiveTestSuite) TestWriteToZipError() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	expectedErr := fmt.Errorf("Write failed")
	mockZipWriter := testutil.NewZipWriterBuilder().WithoutCreate().Build(t)
	mockZipFile := testutil.NewZipFileBuilder().WithoutWrite().Build(t)
	mockZipFile.EXPECT().Write(mock.Anything).Return(0, expectedErr)
	mockZipWriter.EXPECT().Create(mock.Anything).RunAndReturn(func(name string) (io.Writer, error) {
		return mockZipFile, nil
	})
	mockFileUtil := testutil.NewFileUtilBuilder().SetZipWriter(mockZipWriter).Build(t)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:            fsys,
		giveFileUtil:      mockFileUtil,
		giveArchiveFilter: nil,
		wantError:         expectedErr,
		wantResult:        nil,
	})
}

func TestArchiveTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveTestSuite))
}

func TestMetadataArchiveFilterTestSuite(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveNlxMetadata *pkg.NlxMetadata
		giveItems       []string
		wantResult      bool
	}{
		{
			testDescription: "does not contain when metadata nil",
			giveNlxMetadata: nil,
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "pages/my_page", "files/my_file.txt", "files/my_file.txt.skuid.json", "nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      false,
		},
		{
			testDescription: "does not contain when metadata empty",
			giveNlxMetadata: &pkg.NlxMetadata{},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "pages/my_page", "files/my_file.txt", "files/my_file.txt.skuid.json", "nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      false,
		},
		{
			testDescription: "contains when item by filename in metadata",
			giveNlxMetadata: &pkg.NlxMetadata{
				Pages: []string{"my_page"},
				Files: []string{"my_file.txt"},
			},
			giveItems:  []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult: true,
		},
		{
			// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
			testDescription: "contains when item by name in metadata",
			giveNlxMetadata: &pkg.NlxMetadata{
				Apps:        []string{"my_app"},
				Pages:       []string{"my_page"},
				DataSources: []string{"my_datasource"},
			},
			giveItems:  []string{"apps/my_app", "pages/my_page", "datasources/my_datasource"},
			wantResult: true,
		},
		{
			testDescription: "does not contain when item by filename not in metadata",
			giveNlxMetadata: &pkg.NlxMetadata{
				Pages: []string{"my_page"},
				Files: []string{"my_file.txt"},
			},
			giveItems:  []string{"pages/fake_page.json", "pages/fake_page.xml", "files/fake_file.txt", "files/fake_file.json"},
			wantResult: false,
		},
		{
			// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
			testDescription: "does not contain when item by name not in metadata",
			giveNlxMetadata: &pkg.NlxMetadata{
				Apps:        []string{"my_app"},
				Pages:       []string{"my_page"},
				DataSources: []string{"my_datasource"},
			},
			giveItems:  []string{"apps/fake_app", "pages/fake_page", "datasources/fake_datasource"},
			wantResult: false,
		},
		{
			testDescription: "does not contain when invalid metadata type",
			giveNlxMetadata: &pkg.NlxMetadata{
				Apps:               []string{"nonmetadatatype"},
				AuthProviders:      []string{"nonmetadatatype"},
				ComponentPacks:     []string{"nonmetadatatype"},
				DataServices:       []string{"nonmetadatatype"},
				DataSources:        []string{"nonmetadatatype"},
				DesignSystems:      []string{"nonmetadatatype"},
				Variables:          []string{"nonmetadatatype"},
				Files:              []string{"nonmetadatatype"},
				Pages:              []string{"nonmetadatatype"},
				PermissionSets:     []string{"nonmetadatatype"},
				SitePermissionSets: []string{"nonmetadatatype"},
				SessionVariables:   []string{"nonmetadatatype"},
				Site:               []string{"nonmetadatatype"},
				Themes:             []string{"nonmetadatatype"},
			},
			giveItems:  []string{"nometadatatype.json", "nometadatatype.txt", "invalidtype/nonmetadatatype.json", "invalidtype/invalidfolder/nonmetadatatype.json"},
			wantResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			filter := pkg.MetadataArchiveFilter(tc.giveNlxMetadata)
			runArchiveFilterTest(t, filter, tc.giveItems, tc.wantResult)
		})
	}
}

func TestFileNameArchiveFilterTestSuite(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveFiles       []string
		giveItems       []string
		wantResult      bool
	}{
		{
			testDescription: "does not contain when files nil",
			giveFiles:       nil,
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "pages/my_page", "files/my_file.txt", "files/my_file.txt.skuid.json", "nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      false,
		},
		{
			testDescription: "does not contain when files empty",
			giveFiles:       []string{},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "pages/my_page", "files/my_file.txt", "files/my_file.txt.skuid.json", "nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      false,
		},
		{
			testDescription: "contains when in files",
			giveFiles:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult:      true,
		},
		{
			testDescription: "does not contain when valid metadata type not in files",
			giveFiles:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			giveItems:       []string{"pages/fake_page.json", "pages/fake_page.xml", "files/fake_file.txt", "files/fake_file.json"},
			wantResult:      false,
		},
		{
			testDescription: "does not contain when invalid metadata type not in files",
			giveFiles:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			giveItems:       []string{"nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			filter := pkg.FileNameArchiveFilter(tc.giveFiles)
			runArchiveFilterTest(t, filter, tc.giveItems, tc.wantResult)
		})
	}
}

func TestMetadataDirArchiveFilterFilterTestSuite(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveDirs        []string
		giveItems       []string
		wantResult      bool
	}{
		{
			testDescription: "contains when dirs nil",
			giveDirs:        nil,
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "pages/my_page", "files/my_file.txt", "files/my_file.txt.skuid.json", "nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      true,
		},
		{
			testDescription: "contains when dirs empty",
			giveDirs:        []string{},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "pages/my_page", "files/my_file.txt", "files/my_file.txt.skuid.json", "nometadatatype.json", "nometadatatype.txt", "invalidtype/my_file.json"},
			wantResult:      true,
		},
		{
			testDescription: "contains when valid metadata type item by filename not in dirs",
			giveDirs:        []string{"apps", "datasources"},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult:      true,
		},
		{
			// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
			testDescription: "contains when invalid metadata type item by filename not in dirs",
			giveDirs:        []string{"apps", "datasources"},
			giveItems:       []string{"nometadatatype.json", "nometadatatype.txt", "invalidtype/nonmetadatatype.json", "invalidtype/invalidfolder/nonmetadatatype.json"},
			wantResult:      true,
		},
		{
			// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
			testDescription: "contains when valid metadata type item by name not in dirs",
			giveDirs:        []string{"files", "designsystems"},
			giveItems:       []string{"apps/my_app", "pages/my_page", "datasources/my_datasource"},
			wantResult:      true,
		},
		{
			// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
			testDescription: "contains when invalid metadata type item by name not in dirs",
			giveDirs:        []string{"files", "designsystems"},
			giveItems:       []string{"nometadatatype", "invalidtype/nonmetadatatype", "invalidtype/invalidfolder/nonmetadatatype"},
			wantResult:      true,
		},
		{
			testDescription: "does not contain when item by filename in dirs",
			giveDirs:        []string{"pages", "files"},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult:      false,
		},
		{
			// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
			testDescription: "does not contain when item by name in dirs",
			giveDirs:        []string{"apps", "datasources", "pages"},
			giveItems:       []string{"apps/my_app", "pages/my_page", "datasources/my_datasource"},
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			filter := pkg.MetadataDirArchiveFilter(tc.giveDirs)
			runArchiveFilterTest(t, filter, tc.giveItems, tc.wantResult)
		})
	}
}

func runArchiveTest(t *testing.T, atd ArchiveTestDetails) ([]byte, int, error) {
	result, actualFileCount, err := pkg.Archive(atd.giveFS, atd.giveFileUtil, atd.giveArchiveFilter)

	if atd.wantError != nil {
		assert.EqualError(t, err, atd.wantError.Error())
	} else {
		require.NoError(t, err, "Expected Archive err to be nil, but got not nil")
	}

	assert.Equal(t, len(atd.wantResult), actualFileCount)
	assert.Equal(t, atd.wantResult == nil, result == nil)

	if atd.wantResult != nil {
		actualFiles := readArchive(t, result)
		assert.Equal(t, atd.wantResult, actualFiles)
	}

	return result, actualFileCount, err
}

func runArchiveFilterTest(t *testing.T, filter pkg.ArchiveFilter, items []string, expected bool) {
	for _, item := range items {
		assert.Equal(t, expected, filter(item))
	}
}

func readArchive(t *testing.T, actualResult []byte) testutil.TestFiles {
	require.NotNil(t, actualResult, "Expected actualResult not to be nil")

	zipReader, err := zip.NewReader(bytes.NewReader(actualResult), int64(len(actualResult)))
	require.NoError(t, err, "Expected NewReader err to be nil, but got not nil")

	actualFiles := testutil.TestFiles{}
	for _, zipFile := range zipReader.File {
		func() {
			rc, err := zipFile.Open()
			require.NoError(t, err, "Expected Open err to be nil, but got not nil")

			defer rc.Close()

			var b bytes.Buffer
			_, err = io.Copy(&b, rc)
			require.NoError(t, err, "Expected Copy err to be nil, but got not nil")

			actualFiles[zipFile.Name] = b.String()
		}()
	}

	return actualFiles
}

// files that have a valid metadata type and an expected filename for their metadata type
// go doesn't support const maps so creating closure (vs. var) to ensure not modified by a test
func createValidMetadataTypeFilesFixture() testutil.TestFiles {
	return testutil.TestFiles{
		"apps/my_app.json":                 validJSON,
		"pages/page1.xml":                  validXML,
		"pages/page1.json":                 validJSON,
		"files/file1.js":                   validJS,
		"files/file1.js.skuid.json":        validJSON,
		"files/file2.txt":                  validTXT,
		"files/file2.txt.skuid.json":       validJSON,
		"site/site.json":                   validJSON,
		"site/logo/my_logo.png":            validPNG,
		"site/logo/my_logo.png.skuid_json": validJSON,
	}
}

// files that do not have a valid metadata type
// go doesn't support const maps so creating closure (vs. var) to ensure not modified by a test
func createInvalidMetadataTypeFilesFixture() testutil.TestFiles {
	return testutil.TestFiles{
		"invalid.txt":                   "",
		"invalid.json":                  "",
		"invalidfolder/fake_app.xml":    "",
		"pagesfake/fake_page.xml":       "",
		"sitefake/fake_page.xml":        "",
		"datasourcesfake/fake_page.xml": "",
	}
}
