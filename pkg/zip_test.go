package pkg_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"path"
	"testing"
	"testing/fstest"

	"github.com/bobg/go-generics/v4/slices"
	"github.com/orsinium-labs/enum"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	pkgmocks "github.com/skuid/skuid-cli/pkg/mocks"
	"github.com/skuid/skuid-cli/pkg/util"
	"github.com/skuid/skuid-cli/pkg/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	validXML = `
		<skuid__page unsavedchangeswarning="yes">
		</skuid__page>
	`
	validJS  = `function testme() { console.log('hello'); }`
	validTXT = `hello there`
	validPNG = `
		data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACEAAAAgCAYAAAHrvyFcAAACTklEQVR4nNRXTShEURR+S0tLS0tLS0tLS0sLCzZSVjZSFmykbGajlM0sZSFlIyGSSEhKFEKiSZNG8hdp+I7Ode575753n5mXcerMe/fec793fu+5E6we3ZWHZ8/KIDwxDvASdC7SJFMwNn9Ok8wkRCtiksaMx5hBBAsTraM7xLRVYhCzCuDe/BFta8/t21+XQoX7VxLC0xJ6efuwob+4ZWTb2kh2hG2RNhm7pG3SRmOn8YLybglo5BRIRIgIIFyqwNzurRk09K1Zi7mFy29fYCHsh/qelR8/uJy0cFCk59ZpKfppV+yS3FI1EGS35IvisxsEyQmq616yxqomVDcxINpYcn79OlpNkcoSTy1CtMa2NvavO8E05vBaIEyUGDGbtbIw9cIMH8VFJiwfuL4ik4sTTouiEwDvOE1wsrDjUgFoVDGARqkAkM7hFM/OBBlOfq9RAOS2nEBJD86c+ANMbxXUvOfsYwBNhhonxTKmgI5vHgnMWfrh3ubLADYAXDy+m9ErTCaGk0U7gaSD0XglqQdtpfXB5IudWgk6mEUkXXOZKoEgNg1s0JnOearNZapETYQjC/qfnmgb37PizhdgrnG+aWSqBD4ilcA6E+Z/pYQsrz9TAuWVJNgxcWCA8ZeimkpQM4Bw19RhorDkOCUwxiWB8ycOx9xU8YOzHIeNrxJQmhNRU8KH0V84qa3mI/+9pWFYNLl8VW4e2vSSh8GyiUU64P7lA3W63yjjw/J24VSCCXH1tSyJ0VNgnIucSkgqPb3T/4Y07salTB5kcfQJAAD//xMfR9wAAAAGSURBVAMACOfMIINaPS4AAAAASUVORK5CYII=
	`
)

type ArchiveTestDetails struct {
	giveFS             fs.FS
	giveFileUtil       util.FileUtil
	giveArchiveFilter  pkg.ArchiveFilter
	wantError          error
	wantResultFiles    testutil.TestFiles
	wantResultEntities []metadata.MetadataEntity
}

type EntityValidationTestCase struct {
	giveMetadataType metadata.MetadataType
	giveFiles        []string
}

type ArchiveTestSuite struct {
	suite.Suite
}

func (suite *ArchiveTestSuite) TestNoMetadataDirs() {
	t := suite.T()
	fsysNoDirs := &fstest.MapFS{}
	runArchiveTest(t, ArchiveTestDetails{
		giveFS:             fsysNoDirs,
		giveFileUtil:       util.NewFileUtil(),
		giveArchiveFilter:  nil,
		wantError:          nil,
		wantResultFiles:    nil,
		wantResultEntities: nil,
	})
}

func (suite *ArchiveTestSuite) TestNoMetadataFiles() {
	t := suite.T()
	fsysNoFiles := &fstest.MapFS{
		"pages": {Mode: fs.ModeDir},
		"files": {Mode: fs.ModeDir},
	}
	runArchiveTest(t, ArchiveTestDetails{
		giveFS:             fsysNoFiles,
		giveFileUtil:       util.NewFileUtil(),
		giveArchiveFilter:  nil,
		wantError:          nil,
		wantResultFiles:    nil,
		wantResultEntities: nil,
	})
}

func (suite *ArchiveTestSuite) TestCreatesArchive() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:             fsys,
		giveFileUtil:       util.NewFileUtil(),
		giveArchiveFilter:  nil,
		wantError:          nil,
		wantResultFiles:    validMetadataTypeFiles,
		wantResultEntities: createEntitiesFromFiles(t, validMetadataTypeFiles),
	})
}

func (suite *ArchiveTestSuite) TestFilterCalledOnValidFilenamesForValidMetadataTypes() {
	t := suite.T()
	validMetadataTypeFiles := createValidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	filter := pkgmocks.NewArchiveFilter(t)
	for path := range validMetadataTypeFiles {
		item, err := metadata.NewMetadataEntityFile(path)
		require.NoError(t, err)
		filter.EXPECT().Execute(*item).Return(true)
	}

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:             fsys,
		giveFileUtil:       util.NewFileUtil(),
		giveArchiveFilter:  filter.Execute,
		wantError:          nil,
		wantResultFiles:    validMetadataTypeFiles,
		wantResultEntities: createEntitiesFromFiles(t, validMetadataTypeFiles),
	})

	// note that since we did not set an expectation for anything other than expected paths, if
	// Execute is called for something else, the test will FailNow by testify since it doesn't
	// have an expectation configured for anything but the paths we expect. testify doesn't
	// have explicit support for a "Times(0)/Never" so we can't add a mock.Anything
	// expectation so the below is for clarity and just in case testify changes behavior unexpectedly.
	filter.AssertNumberOfCalls(t, "Execute", len(validMetadataTypeFiles))
}

// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
func (suite *ArchiveTestSuite) TestFilterNotCalledOnInvalidFilenamesForValidMetadataTypes() {
	invalidFiles := testutil.TestFiles{
		"pages/page1":           "",
		"pages/page1.txt":       "",
		"pages/page1.jpg":       "",
		"apps/my_app":           "",
		"apps/my_app.xml":       "",
		"datasources/my_ds":     "",
		"datasources/my_ds.png": "",
	}
	for k, v := range invalidFiles {
		suite.Run(k, func() {
			t := suite.T()
			testFiles := testutil.TestFiles{k: v}
			fsys := testutil.CreateFS(testFiles)
			filter := pkgmocks.NewArchiveFilter(t)
			runArchiveTest(t, ArchiveTestDetails{
				giveFS:             fsys,
				giveFileUtil:       util.NewFileUtil(),
				giveArchiveFilter:  filter.Execute,
				wantError:          fmt.Errorf("%v", k),
				wantResultFiles:    nil,
				wantResultEntities: nil,
			})
			// note that since we did not set an expectation for anything other than expected paths, if
			// Execute is called for something else, the test will FailNow by testify since it doesn't
			// have an expectation configured for anything but the paths we expect. testify doesn't
			// have explicit support for a "Times(0)/Never" so we can't add a mock.Anything
			// expectation so the below is for clarity and just in case testify changes behavior unexpectedly.
			filter.AssertNumberOfCalls(t, "Execute", 0)
		})
	}
}

func (suite *ArchiveTestSuite) TestFilterNotCalledOnFilenamesForInvalidMetadataTypes() {
	t := suite.T()
	invalidMetadataTypeFiles := createInvalidMetadataTypeFilesFixture()
	fsys := testutil.CreateFS(invalidMetadataTypeFiles)
	filter := pkgmocks.NewArchiveFilter(t)

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:             fsys,
		giveFileUtil:       util.NewFileUtil(),
		giveArchiveFilter:  filter.Execute,
		wantError:          nil,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             testutil.CreateFS(validMetadataTypeFiles),
		giveFileUtil:       util.NewFileUtil(),
		giveArchiveFilter:  func(item metadata.MetadataEntityFile) bool { return false },
		wantError:          nil,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
	})
}

// TODO: Skuid Review Required - See comment above ArchiveFilter type in pkg/zip.go
func (suite *ArchiveTestSuite) TestWalkDirOnlyMetadataTypeDirsThatExist() {
	t := suite.T()

	validMetadataTypeFiles := testutil.TestFiles{
		"pages/page1.xml":  validXML,
		"pages/page1.json": `{"name": "page1"}`,
	}
	fsys := testutil.CreateFS(validMetadataTypeFiles)
	mockFileUtil := testutil.NewFileUtilBuilder().WithoutWalkDir().Build(t)
	mockFileUtil.EXPECT().WalkDir(fsys, "pages", mock.Anything).RunAndReturn(func(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
		return fs.WalkDir(fsys, root, fn)
	}).Once()

	runArchiveTest(t, ArchiveTestDetails{
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          nil,
		wantResultFiles:    validMetadataTypeFiles,
		wantResultEntities: createEntitiesFromFiles(t, validMetadataTypeFiles),
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          nil,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
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
		giveFS:             fsys,
		giveFileUtil:       mockFileUtil,
		giveArchiveFilter:  nil,
		wantError:          expectedErr,
		wantResultFiles:    nil,
		wantResultEntities: nil,
	})
}

func (suite *ArchiveTestSuite) TestEntityValidationFailure() {
	var testCases []EntityValidationTestCase

	// for types that require more than one file, a validation error will be encountered
	// all types will have a validation error encountered due to missing entity name in definition file
	for _, mdt := range metadata.MetadataTypes.Members() {
		switch mdt {
		case metadata.MetadataTypeSite:
			testCases = append(testCases,
				EntityValidationTestCase{mdt, []string{"site.json"}},
				EntityValidationTestCase{mdt, []string{"favicon/my_icon.ico.skuid.json"}},
				EntityValidationTestCase{mdt, []string{"logo/my_logo.png.skuid.json"}},
			)
		case metadata.MetadataTypeComponentPacks:
			// unable to validate component packs because the only validation performed is that there is at least 1 file
			// and archive only operates on the presence of a file.
			continue
		default:
			testCases = append(testCases, EntityValidationTestCase{mdt, []string{"my_mdt.json"}})
		}
	}

	for _, tc := range testCases {
		suite.Run(tc.giveMetadataType.Name(), func() {
			t := suite.T()
			testFiles := make(testutil.TestFiles)
			for _, fp := range tc.giveFiles {
				testFiles[path.Join(tc.giveMetadataType.DirName(), fp)] = "" // force validation failure due to empty entity name in file
			}
			fsys := testutil.CreateFS(testFiles)

			runArchiveTest(t, ArchiveTestDetails{
				giveFS:             fsys,
				giveFileUtil:       util.NewFileUtil(),
				giveArchiveFilter:  nil,
				wantError:          fmt.Errorf("one or more entities found in %v were invalid", logging.QuoteText(fmt.Sprint(fsys))),
				wantResultFiles:    nil,
				wantResultEntities: nil,
			})
		})
	}
}

func TestArchiveTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveTestSuite))
}

func TestMetadataArchiveFilter(t *testing.T) {
	validFileNames := createValidMetadataTypeFilesNamesFixture()
	testCases := []struct {
		testDescription string
		giveNlxMetadata *metadata.NlxMetadata
		giveItems       []string
		wantResult      bool
	}{
		{
			testDescription: "does not contain when metadata nil",
			giveNlxMetadata: nil,
			giveItems:       validFileNames,
			wantResult:      false,
		},
		{
			testDescription: "does not contain when metadata empty",
			giveNlxMetadata: &metadata.NlxMetadata{},
			giveItems:       validFileNames,
			wantResult:      false,
		},
		{
			testDescription: "contains when item by filename in metadata",
			giveNlxMetadata: &metadata.NlxMetadata{
				Pages: []string{"my_page"},
				Files: []string{"my_file.txt"},
			},
			giveItems:  []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult: true,
		},
		{
			testDescription: "does not contain when item by filename not in metadata",
			giveNlxMetadata: &metadata.NlxMetadata{
				Pages: []string{"my_page"},
				Files: []string{"my_file.txt"},
			},
			giveItems:  []string{"pages/fake_page.json", "pages/fake_page.xml", "files/fake_file.txt", "files/fake_file.json"},
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

func TestMetadataEntityFileArchiveFilter(t *testing.T) {
	validFileNames := createValidMetadataTypeFilesNamesFixture()
	testCases := []struct {
		testDescription string
		giveFiles       []string
		giveItems       []string
		wantResult      bool
	}{
		{
			testDescription: "does not contain when files nil",
			giveFiles:       nil,
			giveItems:       validFileNames,
			wantResult:      false,
		},
		{
			testDescription: "does not contain when files empty",
			giveFiles:       []string{},
			giveItems:       validFileNames,
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
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			var entities []metadata.MetadataEntity
			for _, f := range tc.giveFiles {
				item, err := metadata.NewMetadataEntityFile(f)
				require.NoError(t, err)
				entities = append(entities, item.Entity)
			}
			filter := pkg.MetadataEntityArchiveFilter(entities)
			runArchiveFilterTest(t, filter, tc.giveItems, tc.wantResult)
		})
	}
}

func TestMetadataTypeArchiveFilter(t *testing.T) {
	validFileNames := createValidMetadataTypeFilesNamesFixture()
	testCases := []struct {
		testDescription string
		giveDirs        []string
		giveItems       []string
		wantResult      bool
	}{
		{
			testDescription: "contains when dirs nil",
			giveDirs:        nil,
			giveItems:       validFileNames,
			wantResult:      true,
		},
		{
			testDescription: "contains when dirs empty",
			giveDirs:        []string{},
			giveItems:       validFileNames,
			wantResult:      true,
		},
		{
			testDescription: "contains when valid metadata type item by filename not in dirs",
			giveDirs:        []string{"apps", "datasources"},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult:      true,
		},
		{
			testDescription: "does not contain when item by filename in dirs",
			giveDirs:        []string{"pages", "files"},
			giveItems:       []string{"pages/my_page.json", "pages/my_page.xml", "files/my_file.txt", "files/my_file.txt.skuid.json"},
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			var metadataTypes []metadata.MetadataType
			for _, d := range tc.giveDirs {
				mdt := enum.Parse(metadata.MetadataTypes, metadata.MetadataTypeValue(d))
				require.NotNil(t, mdt)
				metadataTypes = append(metadataTypes, *mdt)
			}
			filter := pkg.MetadataTypeArchiveFilter(metadataTypes)
			runArchiveFilterTest(t, filter, tc.giveItems, tc.wantResult)
		})
	}
}

func runArchiveTest(t *testing.T, atd ArchiveTestDetails) {
	result, archivedFilePaths, archivedEntities, err := pkg.Archive(atd.giveFS, atd.giveFileUtil, atd.giveArchiveFilter)

	if atd.wantError != nil {
		assert.ErrorContains(t, err, atd.wantError.Error())
	} else {
		require.NoError(t, err, "Expected Archive err to be nil, but got not nil")
	}

	assert.ElementsMatch(t, slices.Collect(maps.Keys(atd.wantResultFiles)), archivedFilePaths)
	assert.ElementsMatch(t, atd.wantResultEntities, archivedEntities)

	if atd.wantResultFiles != nil {
		actualFiles := readArchive(t, result)
		assert.Equal(t, atd.wantResultFiles, actualFiles)
	}
}

func runArchiveFilterTest(t *testing.T, filter pkg.ArchiveFilter, items []string, expected bool) {
	for _, item := range items {
		metadataItem, err := metadata.NewMetadataEntityFile(item)
		require.NoError(t, err)
		assert.Equal(t, expected, filter(*metadataItem))
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
		"apps/my_app.json":                                 `{"name": "my_app"}`,
		"authproviders/my_ap.json":                         `{"name": "my_ap"}`,
		"componentpacks/mycomponents/custom_runtime.json":  ``,
		"componentpacks/mycomponents/custom_builders.json": ``,
		"componentpacks/mycomponents/js/custom.js":         ``,
		"componentpacks/mycomponents/css/custom.css":       ``,
		// Skuid Review Required - Unable to determine how to create dataservices, there is no "add" option on the Data Services tab
		// within "Data Sources" in settings and the "DefaultSiteCertificate" that is there doesn't download with a retrieve. What is the structure
		// and how to create and should DefaultSiteCertificate be downloaded on a retrieve?
		"dataservices/my_ds.json":  `{"name": "my_ds"}`,
		"datasources/my_ds.json":   `{"name": "my_ds"}`,
		"designsystems/my_ds.json": `{"objectData": {"name": "my_ds"}}`,
		// Skuid Review Required - Unable to determine how to create variables, what is the structure and how to create?
		"variables/my_variables.json":    `{"name": "my_variables"}`,
		"files/file1.js":                 validJS,
		"files/file1.js.skuid.json":      `{"name": "file1.js"}`,
		"files/file2.txt":                validTXT,
		"files/file2.txt.skuid.json":     `{"name": "file2.txt"}`,
		"pages/page1.xml":                validXML,
		"pages/page1.json":               `{"name": "page1"}`,
		"permissionsets/my_ps.json":      `{"name": "my_ps"}`,
		"sitepermissionsets/editor.json": `{"name": "editor"}`,
		// Skuid Review Required - Unable to determine how to create sessionvariables, what is the structure and how to create?
		"sessionvariables/my_sv.json":        `{"name": "my_sv"}`,
		"site/site.json":                     `{"name": "my site name"}`,
		"site/logo/foobar.png":               validPNG,
		"site/logo/foobar.png.skuid.json":    `{"name": "foobar.png"}`,
		"site/logo/foobar2.gif":              validPNG,
		"site/logo/foobar2.gif.skuid.json":   `{"name": "foobar2.gif"}`,
		"site/logo/foobar3.jpg":              validPNG,
		"site/logo/foobar3.jpg.skuid.json":   `{"name": "foobar3.jpg"}`,
		"site/favicon/foobar.ico":            validPNG,
		"site/favicon/foobar.ico.skuid.json": `{"name": "foobar.ico"}`,
		// Skuid Review Required - Unable to determine how to create themes. Docs indicate v1 only but unable to determine
		// how to create v1 page since the "API Version" option is not available when creating a page as the docs indicate
		// it should be (https://docs.skuid.com/nlx/v2/en/skuid/api-version/#how-do-i-use-skuid-s-api-versions).  What is
		// the structure and how to create?
		"themes/my_theme.json": `{"name": "my_theme"}`,
	}
}

func createValidMetadataTypeFilesNamesFixture() []string {
	var files []string
	for fn := range createValidMetadataTypeFilesFixture() {
		files = append(files, fn)
	}
	return files
}

func createEntitiesFromFiles(t *testing.T, files testutil.TestFiles) []metadata.MetadataEntity {
	entities := make(map[metadata.MetadataEntity]metadata.MetadataEntity)
	for f := range files {
		entityFile, err := metadata.NewMetadataEntityFile(f)
		require.NoError(t, err)
		entity := entityFile.Entity
		// if duplicated, just overwrite it
		entities[entity] = entity
	}

	return slices.Collect(maps.Keys(entities))
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
