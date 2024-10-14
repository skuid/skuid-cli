package metadata_test

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bobg/go-generics/v4/set"
	"github.com/orsinium-labs/enum"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TODO: uncomment
// func TestJsonUnmarshalMetadata(t *testing.T) {
// 	for _, tc := range []struct {
// 		description string
// 		given       string
// 		expected    []string
// 	}{
// 		{
// 			description: "profiles",
// 			given: `
// 			{
// 			 "host": "",
// 			 "port": "",
// 			 "url": "",
// 			 "type": "",
// 			 "metadata": {
// 				"apps": null,
// 				"authproviders": null,
// 				"componentpacks": null,
// 				"dataservices": null,
// 				"datasources": null,
// 				"designsystems": null,
// 				"variables": null,
// 				"files": null,
// 				"pages": null,
// 				"permissionsets": null,
// 				"sitepermissionsets": ["A", "B"],
// 				"site": null,
// 				"themes": null
// 			 },
// 			 "warnings": null
// 			}`,
// 			expected: []string{"A", "B"},
// 		},
// 		{
// 			description: "sitepermissionsets",
// 			given: `
// 			{
// 			 "host": "",
// 			 "port": "",
// 			 "url": "",
// 			 "type": "",
// 			 "metadata": {
// 				"apps": null,
// 				"authproviders": null,
// 				"componentpacks": null,
// 				"dataservices": null,
// 				"datasources": null,
// 				"designsystems": null,
// 				"variables": null,
// 				"files": null,
// 				"pages": null,
// 				"permissionsets": null,
// 				"profiles": ["A", "B"],
// 				"site": null,
// 				"themes": null
// 			 },
// 			 "warnings": null
// 			}`,
// 			expected: []string{"A", "B"},
// 		},
// 	} {
// 		t.Run(tc.description, func(t *testing.T) {
// 			plan := metadata.NlxPlan{}
// 			err := json.Unmarshal([]byte(tc.given), &plan)
// 			if err != nil {
// 				t.Log(err)
// 				t.FailNow()
// 			}
// 			assert.Equal(t, tc.expected, plan.Metadata.SitePermissionSets)
// 		})
// 	}
// }

// contains one or more of every valid character for an entity name (except for every lower & upper letter)
const (
	ValidEntityName     = "my _-0123456789 FILE"
	ValidFileEntityName = "my FO0123456789_-.() FILE"
)

type NewMetadataEntityTestCase struct {
	testDescription string
	givePath        string
	wantEntity      *metadata.MetadataEntity
	wantError       error
}

type NewMetadataEntityFileTestCase struct {
	testDescription string
	givePath        string
	wantEntityFile  *metadata.MetadataEntityFile
	wantError       error
}

type ValidateTestCase struct {
	testDescription string
	giveEntity      metadata.MetadataEntity
	giveFiles       []metadata.MetadataEntityFileContents
	wantError       bool
	wantMessages    bool
}

type ValidateFile struct {
	FileName         string
	FileContent      string
	IsDefinitionFile bool
}

func TestMetadataPathError(t *testing.T) {
	testCases := []struct {
		testDescription string
		givePath        string
		givePathType    metadata.MetadataPathType
		giveError       error
		wantName        string
		wantPanic       bool
		wantError       error
	}{
		{
			testDescription: "panics when path empty and err nil",
			givePath:        "",
			givePathType:    metadata.MetadataPathTypeEntity,
			giveError:       nil,
			wantName:        "",
			wantPanic:       true,
			wantError:       nil,
		},
		{
			testDescription: "panics when path not empty and err nil",
			givePath:        "pages/foobar",
			givePathType:    metadata.MetadataPathTypeEntity,
			giveError:       nil,
			wantName:        "mycmd",
			wantPanic:       true,
			wantError:       nil,
		},
		{
			testDescription: "return error when path empty and err not nil",
			givePath:        "",
			givePathType:    metadata.MetadataPathTypeEntity,
			giveError:       assert.AnError,
			wantName:        "",
			wantError:       assert.AnError,
		},
		{
			testDescription: "return error when path not empty and err not nil",
			givePath:        "pages/foobar",
			givePathType:    metadata.MetadataPathTypeEntity,
			giveError:       assert.AnError,
			wantName:        "mycmd",
			wantError:       assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			e := metadata.NewMetadataPathError(tc.givePath, tc.givePathType, tc.giveError)
			require.NotNil(t, e)
			var asMpError *metadata.MetadataPathError
			require.ErrorAs(t, e, &asMpError)
			assert.Equal(t, tc.givePath, asMpError.Path)
			assert.Equal(t, tc.givePathType, asMpError.PathType)
			assert.Equal(t, tc.wantError, asMpError.Unwrap())
			if tc.wantPanic {
				assert.Panics(t, func() {
					assert.Equal(t, tc.wantError.Error(), asMpError.Error())
				})
			} else {
				assert.Equal(t, tc.wantError.Error(), asMpError.Error())
			}
			var isMpError *metadata.MetadataPathError
			assert.False(t, errors.Is(e, isMpError))
		})
	}
}

type NlxMetadataTestSuite struct {
	suite.Suite
}

func (suite *NlxMetadataTestSuite) TestGetFieldValue() {
	nlxMetadata := metadata.NlxMetadata{
		Apps:               []string{"apps"},
		AuthProviders:      []string{"authproviders"},
		ComponentPacks:     []string{"componentpacks"},
		DataServices:       []string{"dataservices"},
		DataSources:        []string{"datasources"},
		DesignSystems:      []string{"designsystems"},
		Variables:          []string{"variables"},
		Files:              []string{"files"},
		Pages:              []string{"pages"},
		PermissionSets:     []string{"permissionsets"},
		SitePermissionSets: []string{"sitepermissionsets"},
		SessionVariables:   []string{"sessionvariables"},
		Site:               []string{"site"},
		Themes:             []string{"themes"},
	}

	for _, tc := range []struct {
		description    string
		given          metadata.MetadataType
		expected       []string
		wantPanicError error
	}{
		{
			description: "apps",
			given:       metadata.MetadataTypeApps,
			expected:    nlxMetadata.Apps,
		},
		{
			description: "authproviders",
			given:       metadata.MetadataTypeAuthProviders,
			expected:    nlxMetadata.AuthProviders,
		},
		{
			description: "componentpacks",
			given:       metadata.MetadataTypeComponentPacks,
			expected:    nlxMetadata.ComponentPacks,
		},
		{
			description: "dataservices",
			given:       metadata.MetadataTypeDataServices,
			expected:    nlxMetadata.DataServices,
		},
		{
			description: "datasources",
			given:       metadata.MetadataTypeDataSources,
			expected:    nlxMetadata.DataSources,
		},
		{
			description: "designsystems",
			given:       metadata.MetadataTypeDesignSystems,
			expected:    nlxMetadata.DesignSystems,
		},
		{
			description: "variables",
			given:       metadata.MetadataTypeVariables,
			expected:    nlxMetadata.Variables,
		},
		{
			description: "files",
			given:       metadata.MetadataTypeFiles,
			expected:    nlxMetadata.Files,
		},
		{
			description: "pages",
			given:       metadata.MetadataTypePages,
			expected:    nlxMetadata.Pages,
		},
		{
			description: "permissionsets",
			given:       metadata.MetadataTypePermissionSets,
			expected:    nlxMetadata.PermissionSets,
		},
		{
			description: "sitepermissionsets",
			given:       metadata.MetadataTypeSitePermissionSets,
			expected:    nlxMetadata.SitePermissionSets,
		},
		{
			description: "sessionvariables",
			given:       metadata.MetadataTypeSessionVariables,
			expected:    nlxMetadata.SessionVariables,
		},
		{
			description: "site",
			given:       metadata.MetadataTypeSite,
			expected:    nlxMetadata.Site,
		},
		{
			description: "themes",
			given:       metadata.MetadataTypeThemes,
			expected:    nlxMetadata.Themes,
		},
		{
			description:    "bad",
			given:          metadata.MetadataType{"bad"},
			wantPanicError: fmt.Errorf("unable to locate metadata field for metadata type %v", "bad"),
		},
	} {
		suite.Run(tc.description, func() {
			t := suite.T()
			if tc.wantPanicError != nil {
				assert.PanicsWithError(t, tc.wantPanicError.Error(), func() {
					nlxMetadata.GetFieldValue(tc.given)
				})
			} else {
				assert.Equal(t, tc.expected, nlxMetadata.GetFieldValue(tc.given))
			}
		})
	}
}

func (suite *NlxMetadataTestSuite) TestExistsInMetadataTypes() {
	t := suite.T()
	metadataType := reflect.TypeOf(metadata.NlxMetadata{})

	for i := 0; i < metadataType.NumField(); i++ {
		field := metadataType.Field(i)
		dirName := field.Tag.Get("json")
		assert.GreaterOrEqual(t, len(dirName), 1)
		mdt := enum.Parse(metadata.MetadataTypes, metadata.MetadataTypeValue(field.Name))
		require.NotNil(t, mdt)
		assert.Equal(t, dirName, mdt.DirName())
	}
}

func (suite *NlxMetadataTestSuite) TestFilterItem() {
	testCases := []struct {
		testDescription string
		giveMetadata    metadata.NlxMetadata
		giveFile        metadata.MetadataEntityFile
		wantResult      bool
		wantPanicError  error
	}{
		{
			testDescription: "exists in metadata",
			giveMetadata: metadata.NlxMetadata{
				Apps: []string{"my_app"},
			},
			giveFile: metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeApps,
					PathRelative: "my_app",
				},
			},
			wantResult: true,
		},
		{
			testDescription: "does not exist in metadata",
			giveMetadata:    metadata.NlxMetadata{},
			giveFile: metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeApps,
					PathRelative: "not_there",
				},
			},
			wantResult: false,
		},
		{
			testDescription: "metadatatype not found in metadata",
			giveFile: metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type: metadata.MetadataType{"bad"},
				},
			},
			wantPanicError: fmt.Errorf("unable to locate metadata field for metadata type %v", "bad"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			if tc.wantPanicError != nil {
				assert.PanicsWithError(t, tc.wantPanicError.Error(), func() {
					tc.giveMetadata.FilterItem(tc.giveFile)
				})
			} else {
				actualResult := tc.giveMetadata.FilterItem(tc.giveFile)
				assert.Equal(t, tc.wantResult, actualResult)
			}
		})
	}
}

func TestNlxMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(NlxMetadataTestSuite))
}

type MetadataTypeValueTestSuite struct {
	suite.Suite
}

func (suite *MetadataTypeValueTestSuite) TestEqual() {
	testCases := []struct {
		testDescription string
		giveSource      metadata.MetadataTypeValue
		giveTarget      metadata.MetadataTypeValue
		wantResult      bool
	}{
		{
			testDescription: "equals both lower",
			giveSource:      metadata.MetadataTypeValue("apps"),
			giveTarget:      metadata.MetadataTypeValue("apps"),
			wantResult:      true,
		},
		{
			testDescription: "equals both upper",
			giveSource:      metadata.MetadataTypeValue("APPS"),
			giveTarget:      metadata.MetadataTypeValue("APPS"),
			wantResult:      true,
		},
		{
			testDescription: "equals both mixed",
			giveSource:      metadata.MetadataTypeValue("AppS"),
			giveTarget:      metadata.MetadataTypeValue("AppS"),
			wantResult:      true,
		},
		{
			testDescription: "equals lower and upper",
			giveSource:      metadata.MetadataTypeValue("apps"),
			giveTarget:      metadata.MetadataTypeValue("APPS"),
			wantResult:      true,
		},
		{
			testDescription: "equals upper and lower",
			giveSource:      metadata.MetadataTypeValue("APPS"),
			giveTarget:      metadata.MetadataTypeValue("apps"),
			wantResult:      true,
		},
		{
			testDescription: "not equals",
			giveSource:      metadata.MetadataTypeValue("foo"),
			giveTarget:      metadata.MetadataTypeValue("bar"),
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			assert.Equal(t, tc.wantResult, tc.giveSource.Equal(tc.giveTarget))
		})
	}
}

func TestMetadataTypeValueTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataTypeValueTestSuite))
}

type MetadataTypeTestSuite struct {
	suite.Suite
}

func (suite *MetadataTypeTestSuite) TestName() {
	testCases := []struct {
		testDescription string
		giveName        string
		wantName        string
	}{
		{
			testDescription: "lowercase",
			giveName:        "hello",
			wantName:        "hello",
		},
		{
			testDescription: "uppercase",
			giveName:        "HELLO",
			wantName:        "HELLO",
		},
		{
			testDescription: "mixed case",
			giveName:        "HeLlO",
			wantName:        "HeLlO",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			mdt := metadata.MetadataType{metadata.MetadataTypeValue(tc.giveName)}
			assert.Equal(t, tc.wantName, mdt.Name())
		})
	}
}

func (suite *MetadataTypeTestSuite) TestDirName() {
	testCases := []struct {
		testDescription string
		giveName        string
		wantDirName     string
	}{
		{
			testDescription: "lowercase",
			giveName:        "hello",
			wantDirName:     "hello",
		},
		{
			testDescription: "uppercase",
			giveName:        "HELLO",
			wantDirName:     "hello",
		},
		{
			testDescription: "mixed case",
			giveName:        "HeLlO",
			wantDirName:     "hello",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			mdt := metadata.MetadataType{metadata.MetadataTypeValue(tc.giveName)}
			assert.Equal(t, tc.wantDirName, mdt.DirName())
		})
	}
}

func TestMetadataTypeTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataTypeTestSuite))
}

type MetadataTypesTestSuite struct {
	suite.Suite
}

func (suite *MetadataTypesTestSuite) TestExistsInNlxMetadata() {
	t := suite.T()
	metadataType := reflect.TypeOf(metadata.NlxMetadata{})
	for _, mdt := range metadata.MetadataTypes.Members() {
		f, ok := metadataType.FieldByName(mdt.Name())
		require.True(t, ok)
		assert.Equal(t, mdt.DirName(), f.Tag.Get("json"))
	}
}

func (suite *MetadataTypesTestSuite) TestNameValid() {
	t := suite.T()
	for _, mdt := range metadata.MetadataTypes.Members() {
		assert.GreaterOrEqual(t, len(strings.TrimSpace(mdt.Name())), 1)
	}
}

func (suite *MetadataTypesTestSuite) TestDirNameValid() {
	t := suite.T()
	for _, mdt := range metadata.MetadataTypes.Members() {
		assert.GreaterOrEqual(t, len(strings.TrimSpace(mdt.DirName())), 1)
	}
}

func TestMetadataTypesTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataTypesTestSuite))
}

func TestNewMetadataEntity(t *testing.T) {
	createFixture := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, name string) *metadata.MetadataEntity {
		return &metadata.MetadataEntity{
			Type:         mdt,
			SubType:      mdtst,
			Name:         name,
			Path:         mdt.DirName() + "/" + name,
			PathRelative: name,
		}
	}

	createStandardValidTestCases := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, entityName string) []NewMetadataEntityTestCase {
		return []NewMetadataEntityTestCase{
			{
				testDescription: fmt.Sprintf("%v valid", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v", mdt.DirName(), entityName),
				wantEntity:      createFixture(mdt, mdtst, entityName),
			},
			{
				testDescription: fmt.Sprintf("%v valid . relative", mdt.Name()),
				givePath:        fmt.Sprintf("./%v/%v", mdt.DirName(), entityName),
				wantEntity:      createFixture(mdt, mdtst, entityName),
			},
		}
	}

	createStandardInvalidTestCases := func(mdt metadata.MetadataType, entityName string) []NewMetadataEntityTestCase {
		return []NewMetadataEntityTestCase{
			{
				testDescription: fmt.Sprintf("%v invalid contains json filename", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.json", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.json", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains skuid.json filename", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.skuid.json", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.skuid.json", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains txt filename", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.txt", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.txt", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains xml filename", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.xml", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.xml", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains nested", mdt.Name()),
				givePath:        fmt.Sprintf("%v/subdir/%v", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains invalid . character", mdt.Name()),
				givePath:        fmt.Sprintf("%v/pre.%v", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/pre.%v", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains invalid $ character", mdt.Name()),
				givePath:        fmt.Sprintf("%v/pre$%v", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/pre$%v", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains invalid ^ character", mdt.Name()),
				givePath:        fmt.Sprintf("%v/pre^%v", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/pre^%v", mdt.DirName(), entityName)),
			},
		}
	}

	createStandardTestCases := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, entityName string) []NewMetadataEntityTestCase {
		return append(createStandardValidTestCases(mdt, mdtst, entityName), createStandardInvalidTestCases(mdt, entityName)...)
	}

	filesTestCases := []NewMetadataEntityTestCase{
		{
			testDescription: "Files valid no extension",
			givePath:        "files/" + ValidFileEntityName,
			wantEntity:      createFixture(metadata.MetadataTypeFiles, metadata.MetadataSubTypeNone, ValidFileEntityName),
		},
		{
			testDescription: "Files valid contains json extension",
			givePath:        "files/md_file.json",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, metadata.MetadataSubTypeNone, "md_file.json"),
		},
		{
			testDescription: "Files valid contains skuid.json extension",
			givePath:        "files/md_file.skuid.json",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, metadata.MetadataSubTypeNone, "md_file.skuid.json"),
		},
		{
			testDescription: "Files valid contains txt filename",
			givePath:        "files/md_file.txt",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, metadata.MetadataSubTypeNone, "md_file.txt"),
		},
		{
			testDescription: "Files valid contains xml filename",
			givePath:        "files/md_file.xml",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, metadata.MetadataSubTypeNone, "md_file.xml"),
		},
		{
			testDescription: "Files invalid contains nested no extension",
			givePath:        "files/subdir/md_file",
			wantError:       createPathError(metadata.MetadataTypeFiles, "files/subdir/md_file"),
		},
		{
			testDescription: "Files invalid contains nested with extension",
			givePath:        "files/subdir/md_file.txt",
			wantError:       createPathError(metadata.MetadataTypeFiles, "files/subdir/md_file.txt"),
		},
		{
			testDescription: "Files invalid contains invalid $ character",
			givePath:        "files/pre$%v",
			wantError:       createPathError(metadata.MetadataTypeFiles, "files/pre$%v"),
		},
		{
			testDescription: "Files invalid contains invalid $ character",
			givePath:        "files/pre^%v",
			wantError:       createPathError(metadata.MetadataTypeFiles, "files/pre^%v"),
		},
	}

	siteTestCases := []NewMetadataEntityTestCase{
		{
			testDescription: "Site favicon valid",
			givePath:        "site/favicon/my_favicon.ico",
			wantEntity: &metadata.MetadataEntity{
				Type:         metadata.MetadataTypeSite,
				SubType:      metadata.MetadataSubTypeSiteFavicon,
				Name:         "my_favicon.ico",
				Path:         "site/favicon/my_favicon.ico",
				PathRelative: "favicon/my_favicon.ico",
			},
		},
		{
			testDescription: "Site favicon invalid no extension",
			givePath:        "site/favicon/my_favicon",
			wantError:       createPathError(metadata.MetadataTypeSite, "site/favicon/my_favicon"),
		},
		{
			testDescription: "Site favicon invalid contains extension",
			givePath:        "site/favicon/my_favicon.ico.skuid.json",
			wantError:       createPathError(metadata.MetadataTypeSite, "site/favicon/my_favicon.ico.skuid.json"),
		},
		{
			testDescription: "Site logo invalid no extension",
			givePath:        "site/logo/my_logo",
			wantError:       createPathError(metadata.MetadataTypeSite, "site/logo/my_logo"),
		},
		{
			testDescription: "Site logo valid png",
			givePath:        "site/logo/my_logo.png",
			wantEntity: &metadata.MetadataEntity{
				Type:         metadata.MetadataTypeSite,
				SubType:      metadata.MetadataSubTypeSiteLogo,
				Name:         "my_logo.png",
				Path:         "site/logo/my_logo.png",
				PathRelative: "logo/my_logo.png",
			},
		},
		{
			testDescription: "Site logo invalid png contains extension",
			givePath:        "site/logo/my_logo.png.skuid.json",
			wantError:       createPathError(metadata.MetadataTypeSite, "site/logo/my_logo.png.skuid.json"),
		},
		{
			testDescription: "Site logo valid jpg",
			givePath:        "site/logo/my_logo.jpg",
			wantEntity: &metadata.MetadataEntity{
				Type:         metadata.MetadataTypeSite,
				SubType:      metadata.MetadataSubTypeSiteLogo,
				Name:         "my_logo.jpg",
				Path:         "site/logo/my_logo.jpg",
				PathRelative: "logo/my_logo.jpg",
			},
		},
		{
			testDescription: "Site logo invalid jpg contains extension",
			givePath:        "site/logo/my_logo.jpg.skuid.json",
			wantError:       createPathError(metadata.MetadataTypeSite, "site/logo/my_logo.jpg.skuid.json"),
		},
		{
			testDescription: "Site logo valid gif",
			givePath:        "site/logo/my_logo.gif",
			wantEntity: &metadata.MetadataEntity{
				Type:         metadata.MetadataTypeSite,
				SubType:      metadata.MetadataSubTypeSiteLogo,
				Name:         "my_logo.gif",
				Path:         "site/logo/my_logo.gif",
				PathRelative: "logo/my_logo.gif",
			},
		},
		{
			testDescription: "Site logo invalid gif contains extension",
			givePath:        "site/logo/my_logo.gif.skuid.json",
			wantError:       createPathError(metadata.MetadataTypeSite, "site/logo/my_logo.gif.skuid.json"),
		},
	}

	testCases := []NewMetadataEntityTestCase{
		{
			testDescription: "invalid empty path",
			givePath:        "",
			wantError:       createContainMetadataNameError(""),
		},
		{
			testDescription: "invalid whitespace path",
			givePath:        "         ",
			wantError:       createContainMetadataNameError("         "),
		},
		{
			testDescription: "invalid no metadata type",
			givePath:        "md_file",
			wantError:       createContainMetadataNameError("md_file"),
		},
		{
			testDescription: "invalid no metadata type with extension",
			givePath:        "md_file.xml",
			wantError:       createContainMetadataNameError("md_file.xml"),
		},
		{
			testDescription: "invalid absolute path",
			givePath:        "/foo/md_file",
			wantError:       createContainMetadataNameError("/foo/md_file"),
		},
		{
			testDescription: "invalid absolute root with name path",
			givePath:        "/md_file",
			wantError:       createContainMetadataNameError("/md_file"),
		},
		{
			testDescription: "invalid absolute root without name path",
			givePath:        "/",
			wantError:       createContainMetadataNameError("/"),
		},
		{
			testDescription: "invalid relative .. path",
			givePath:        "../pages/md_file",
			wantError:       createInvalidMetadataNameError("..", "../pages/md_file"),
		},
		{
			testDescription: "invalid metadata type",
			givePath:        "unknowntype/md_file",
			wantError:       createInvalidMetadataNameError("unknowntype", "unknowntype/md_file"),
		},
	}

	for _, mdt := range metadata.MetadataTypes.Members() {
		switch mdt {
		case metadata.MetadataTypeFiles:
			testCases = append(testCases, filesTestCases...)
		case metadata.MetadataTypeSite:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, "site")...)
			testCases = append(testCases, siteTestCases...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, ValidEntityName)...)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			// Note - Forcing givePath to platform file separator to simulate reading from disk
			entity, err := metadata.NewMetadataEntity(filepath.FromSlash(tc.givePath))
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantEntity, entity)
		})
	}
}

func TestNewMetadataEntityFile(t *testing.T) {
	createFixture := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, name string, fileName string, isEntityDefinitionFile bool) *metadata.MetadataEntityFile {
		return &metadata.MetadataEntityFile{
			Entity: metadata.MetadataEntity{
				Type:         mdt,
				SubType:      mdtst,
				Name:         name,
				Path:         path.Join(mdt.DirName(), name),
				PathRelative: name,
			},
			Name:                   fileName,
			Path:                   path.Join(mdt.DirName(), fileName),
			PathRelative:           fileName,
			IsEntityDefinitionFile: isEntityDefinitionFile,
		}
	}

	createStandardValidTestCases := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, entityName string) []NewMetadataEntityFileTestCase {
		return []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v valid json extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.json", mdt.DirName(), entityName),
				wantEntityFile:  createFixture(mdt, mdtst, entityName, entityName+".json", true),
			},
			{
				testDescription: fmt.Sprintf("%v valid . relative", mdt.Name()),
				givePath:        fmt.Sprintf("./%v/%v.json", mdt.DirName(), entityName),
				wantEntityFile:  createFixture(mdt, mdtst, entityName, entityName+".json", true),
			},
		}
	}

	createStandardInvalidTestCases := func(mdt metadata.MetadataType, entityName string, includeNestedDirTests bool, includeXmlTests bool) []NewMetadataEntityFileTestCase {
		tests := []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v invalid no extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains multiple dots filename", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.skuid.json", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.skuid.json", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains txt filename", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.txt", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.txt", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains invalid . character", mdt.Name()),
				givePath:        fmt.Sprintf("%v/pre.%v.json", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/pre.%v.json", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains invalid $ character", mdt.Name()),
				givePath:        fmt.Sprintf("%v/pre$%v.json", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/pre$%v.json", mdt.DirName(), entityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains invalid ^ character", mdt.Name()),
				givePath:        fmt.Sprintf("%v/pre^%v.json", mdt.DirName(), entityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/pre^%v.json", mdt.DirName(), entityName)),
			},
		}

		if includeNestedDirTests {
			tests = append(tests, []NewMetadataEntityFileTestCase{
				{
					testDescription: fmt.Sprintf("%v invalid contains nested no extension", mdt.Name()),
					givePath:        fmt.Sprintf("%v/subdir/%v", mdt.DirName(), entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v", mdt.DirName(), entityName)),
				},
				{
					testDescription: fmt.Sprintf("%v invalid contains nested json extension", mdt.Name()),
					givePath:        fmt.Sprintf("%v/subdir/%v.json", mdt.DirName(), entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v.json", mdt.DirName(), entityName)),
				}}...)
		}

		if includeXmlTests {
			tests = append(tests, []NewMetadataEntityFileTestCase{
				{
					testDescription: fmt.Sprintf("%v invalid contains xml filename", mdt.Name()),
					givePath:        fmt.Sprintf("%v/%v.xml", mdt.DirName(), entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/%v.xml", mdt.DirName(), entityName)),
				},
			}...)
		}

		return tests
	}

	createStandardTestCases := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, entityName string) []NewMetadataEntityFileTestCase {
		return append(createStandardValidTestCases(mdt, mdtst, entityName), createStandardInvalidTestCases(mdt, entityName, true, true)...)
	}

	themesTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: "Themes valid .inline.css extension",
			givePath:        fmt.Sprintf("%v/%v.inline.css", metadata.MetadataTypeThemes.DirName(), ValidEntityName),
			wantEntityFile:  createFixture(metadata.MetadataTypeThemes, metadata.MetadataSubTypeNone, ValidEntityName, ValidEntityName+".inline.css", false),
		},
	}

	componentPacksTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: fmt.Sprintf("%v valid nested dir json extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					SubType:      metadata.MetadataSubTypeNone,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   ValidEntityName + ".json",
				Path:                   fmt.Sprintf("%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
				PathRelative:           fmt.Sprintf("subdir/%v.json", ValidEntityName),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir js extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.js", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					SubType:      metadata.MetadataSubTypeNone,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   ValidEntityName + ".js",
				Path:                   fmt.Sprintf("%v/subdir/%v.js", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
				PathRelative:           fmt.Sprintf("subdir/%v.js", ValidEntityName),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir css extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.css", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					SubType:      metadata.MetadataSubTypeNone,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   ValidEntityName + ".css",
				Path:                   fmt.Sprintf("%v/subdir/%v.css", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
				PathRelative:           fmt.Sprintf("subdir/%v.css", ValidEntityName),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir no extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					SubType:      metadata.MetadataSubTypeNone,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   ValidEntityName,
				Path:                   fmt.Sprintf("%v/subdir/%v", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
				PathRelative:           fmt.Sprintf("subdir/%v", ValidEntityName),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir . relative json extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("./%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					SubType:      metadata.MetadataSubTypeNone,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   ValidEntityName + ".json",
				Path:                   fmt.Sprintf("%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
				PathRelative:           fmt.Sprintf("subdir/%v.json", ValidEntityName),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v invalid json file in metadata dir root", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid . relative to componentpacks directory", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("./%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("./%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains invalid $ character", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/sub$dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/sub$dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains invalid ^ character", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/sub^dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/sub^dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), ValidEntityName)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains no filename", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir", metadata.MetadataTypeComponentPacks.DirName()),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/subdir", metadata.MetadataTypeComponentPacks.DirName())),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains no filename ends with slash", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/", metadata.MetadataTypeComponentPacks.DirName()),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/subdir/", metadata.MetadataTypeComponentPacks.DirName())),
		},
	}

	pagesTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: fmt.Sprintf("%v valid xml extension", metadata.MetadataTypePages.Name()),
			givePath:        fmt.Sprintf("%v/%v.xml", metadata.MetadataTypePages.DirName(), ValidEntityName),
			wantEntityFile:  createFixture(metadata.MetadataTypePages, metadata.MetadataSubTypeNone, ValidEntityName, ValidEntityName+".xml", false),
		},
	}

	createSiteImageTestCases := func(subdir string, subType metadata.MetadataSubType, extensions []string) []NewMetadataEntityFileTestCase {
		mdt := metadata.MetadataTypeSite

		testCases := []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v %v invalid no extension", mdt.Name(), subdir),
				givePath:        fmt.Sprintf("%v/%v/%v.xml", mdt.DirName(), subdir, ValidFileEntityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v/%v.xml", mdt.DirName(), subdir, ValidFileEntityName)),
			},
		}

		for _, ext := range extensions {
			entityName := ValidFileEntityName + ext
			testCases = append(testCases, []NewMetadataEntityFileTestCase{
				{
					testDescription: fmt.Sprintf("%v %v valid %v extension", mdt.Name(), subdir, ext),
					givePath:        fmt.Sprintf("%v/%v/%v", mdt.DirName(), subdir, entityName),
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
							Type:         mdt,
							SubType:      subType,
							Name:         entityName,
							Path:         path.Join(mdt.DirName(), subdir, entityName),
							PathRelative: path.Join(subdir, entityName),
						},
						Name:                   entityName,
						Path:                   path.Join(mdt.DirName(), subdir, entityName),
						PathRelative:           path.Join(subdir, entityName),
						IsEntityDefinitionFile: false,
					},
				},
				{
					testDescription: fmt.Sprintf("%v %v valid %v.skuid.json extension", mdt.Name(), subdir, ext),
					givePath:        fmt.Sprintf("%v/%v/%v.skuid.json", mdt.DirName(), subdir, entityName),
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
							Type:         mdt,
							SubType:      subType,
							Name:         entityName,
							Path:         path.Join(mdt.DirName(), subdir, entityName),
							PathRelative: path.Join(subdir, entityName),
						},
						Name:                   entityName + ".skuid.json",
						Path:                   path.Join(mdt.DirName(), subdir, entityName+".skuid.json"),
						PathRelative:           path.Join(subdir, entityName+".skuid.json"),
						IsEntityDefinitionFile: true,
					},
				},
				{
					testDescription: fmt.Sprintf("%v %v invalid %v contains json extension", mdt.Name(), subdir, ext),
					givePath:        fmt.Sprintf("%v/%v/%v.json", mdt.DirName(), subdir, entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/%v/%v.json", mdt.DirName(), subdir, entityName)),
				},
				{
					testDescription: fmt.Sprintf("%v %v invalid %v contains $ character", mdt.Name(), subdir, ext),
					givePath:        fmt.Sprintf("%v/%v/pre$%v.json", mdt.DirName(), subdir, entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/%v/pre$%v.json", mdt.DirName(), subdir, entityName)),
				},
				{
					testDescription: fmt.Sprintf("%v %v invalid %v contains ^ character", mdt.Name(), subdir, ext),
					givePath:        fmt.Sprintf("%v/%v/pre^%v.json", mdt.DirName(), subdir, entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/%v/pre^%v.json", mdt.DirName(), subdir, entityName)),
				},
			}...)
		}

		return testCases
	}

	filesTestCases := func() []NewMetadataEntityFileTestCase {
		mdt := metadata.MetadataTypeFiles
		testCases := []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v invalid contains nested no extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/subdir/%v", mdt.DirName(), ValidFileEntityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v", mdt.DirName(), ValidFileEntityName)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains nested json extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/subdir/%v.json", mdt.DirName(), ValidFileEntityName),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v.json", mdt.DirName(), ValidFileEntityName)),
			},
		}

		extensions := []string{"", ".xml", ".json", ".txt", ".jpg", ".gif", ".png", ".js", ".css", ".log", ".txt.log"}
		for _, ext := range extensions {
			entityName := ValidFileEntityName + ext
			testCases = append(testCases, []NewMetadataEntityFileTestCase{
				{
					testDescription: fmt.Sprintf("%v valid %v extension", mdt.Name(), ext),
					givePath:        fmt.Sprintf("%v/%v", mdt.DirName(), entityName),
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
							Type:         mdt,
							SubType:      metadata.MetadataSubTypeNone,
							Name:         entityName,
							Path:         path.Join(mdt.DirName(), entityName),
							PathRelative: entityName,
						},
						Name:                   entityName,
						Path:                   path.Join(mdt.DirName(), entityName),
						PathRelative:           entityName,
						IsEntityDefinitionFile: false,
					},
				},
				{
					testDescription: fmt.Sprintf("%v valid %v.skuid.json extension", mdt.Name(), ext),
					givePath:        fmt.Sprintf("%v/%v.skuid.json", mdt.DirName(), entityName),
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
							Type:         mdt,
							SubType:      metadata.MetadataSubTypeNone,
							Name:         entityName,
							Path:         path.Join(mdt.DirName(), entityName),
							PathRelative: entityName,
						},
						Name:                   entityName + ".skuid.json",
						Path:                   path.Join(mdt.DirName(), entityName+".skuid.json"),
						PathRelative:           entityName + ".skuid.json",
						IsEntityDefinitionFile: true,
					},
				},
				{
					testDescription: fmt.Sprintf("%v invalid %v contains $ character", mdt.Name(), ext),
					givePath:        fmt.Sprintf("%v/pre$%v.json", mdt.DirName(), entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/pre$%v.json", mdt.DirName(), entityName)),
				},
				{
					testDescription: fmt.Sprintf("%v invalid %v contains ^ character", mdt.Name(), ext),
					givePath:        fmt.Sprintf("%v/pre^%v.json", mdt.DirName(), entityName),
					wantError:       createPathError(mdt, fmt.Sprintf("%v/pre^%v.json", mdt.DirName(), entityName)),
				},
			}...)
		}

		return testCases
	}()

	testCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: "invalid empty path",
			givePath:        "",
			wantError:       createContainMetadataNameError(""),
		},
		{
			testDescription: "invalid whitespace path",
			givePath:        "         ",
			wantError:       createContainMetadataNameError("         "),
		},
		{
			testDescription: "invalid no metadata type",
			givePath:        "md_file",
			wantError:       createContainMetadataNameError("md_file"),
		},
		{
			testDescription: "invalid no metadata type with extension",
			givePath:        "md_file.json",
			wantError:       createContainMetadataNameError("md_file.json"),
		},
		{
			testDescription: "invalid absolute subdir path",
			givePath:        "/foo/md_file.json",
			wantError:       createContainMetadataNameError("/foo/md_file.json"),
		},
		{
			testDescription: "invalid absolute root with file path",
			givePath:        "/md_file.json",
			wantError:       createContainMetadataNameError("/md_file.json"),
		},
		{
			testDescription: "invalid absolute root without file path",
			givePath:        "/",
			wantError:       createContainMetadataNameError("/"),
		},
		{
			testDescription: "invalid relative .. path",
			givePath:        "../pages/md_file.json",
			wantError:       createInvalidMetadataNameError("..", "../pages/md_file.json"),
		},
		{
			testDescription: "invalid metadata type",
			givePath:        "unknowntype/md_file",
			wantError:       createInvalidMetadataNameError("unknowntype", "unknowntype/md_file"),
		},
	}

	for _, mdt := range metadata.MetadataTypes.Members() {
		switch mdt {
		case metadata.MetadataTypeFiles:
			testCases = append(testCases, filesTestCases...)
		case metadata.MetadataTypeSite:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, "site")...)
			testCases = append(testCases, createStandardInvalidTestCases(mdt, ValidEntityName, true, true)...)
			testCases = append(testCases, createSiteImageTestCases("favicon", metadata.MetadataSubTypeSiteFavicon, []string{".ico"})...)
			testCases = append(testCases, createSiteImageTestCases("logo", metadata.MetadataSubTypeSiteLogo, []string{".jpg", ".png", ".gif"})...)
		case metadata.MetadataTypePages:
			testCases = append(testCases, createStandardValidTestCases(mdt, metadata.MetadataSubTypeNone, ValidEntityName)...)
			testCases = append(testCases, createStandardInvalidTestCases(mdt, ValidEntityName, true, false)...)
			testCases = append(testCases, pagesTestCases...)
		case metadata.MetadataTypeComponentPacks:
			testCases = append(testCases, createStandardInvalidTestCases(mdt, ValidEntityName, false, true)...)
			testCases = append(testCases, componentPacksTestCases...)
		case metadata.MetadataTypeThemes:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, ValidEntityName)...)
			testCases = append(testCases, themesTestCases...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, ValidEntityName)...)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			// Note - Forcing givePath to platform file separator to simulate reading from disk
			entityFile, err := metadata.NewMetadataEntityFile(filepath.FromSlash(tc.givePath))
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error(), "expected not nil err, got nil for path %v", tc.givePath)
			} else {
				assert.NoError(t, err, "expected nil err, got not nil for path %v", logging.QuoteText(tc.givePath))
			}
			assert.Equal(t, tc.wantEntityFile, entityFile)
		})
	}
}

func TestIsMetadataTypePath(t *testing.T) {
	testCases := []struct {
		testDescription string
		givePath        string
		wantResult      bool
	}{
		{
			testDescription: "valid type with valid entity name",
			givePath:        "pages/my_page",
			wantResult:      true,
		},
		{
			testDescription: "valid type with valid file",
			givePath:        "pages/my_page.xml",
			wantResult:      true,
		},
		{
			testDescription: "valid type with valid definition file",
			givePath:        "pages/my_page.json",
			wantResult:      true,
		},
		{
			testDescription: "valid type with valid nested dir and nested file",
			givePath:        "site/favicon/my_icon.ico",
			wantResult:      true,
		},
		{
			testDescription: "valid type with valid nested dir and nested definition file",
			givePath:        "site/favicon/my_icon.ico.skuid.json",
			wantResult:      true,
		},
		{
			testDescription: "valid type with invalid file",
			givePath:        "pages/my_page.txt",
			wantResult:      true,
		},
		{
			testDescription: "valid type with invalid nested dir",
			givePath:        "site/foobar/my_file",
			wantResult:      true,
		},
		{
			testDescription: "valid type without file",
			givePath:        "pages",
			wantResult:      false,
		},
		{
			testDescription: "invalid type without file",
			givePath:        "foobar",
			wantResult:      false,
		},
		{
			testDescription: "invalid type with file",
			givePath:        "foobar/my_file",
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualResult := metadata.IsMetadataTypePath(tc.givePath)
			assert.Equal(t, tc.wantResult, actualResult)
		})
	}
}

func TestEntitiesMatch(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveA           []metadata.MetadataEntity
		giveB           []metadata.MetadataEntity
		wantResult      bool
	}{
		{
			testDescription: "both nil",
			wantResult:      true,
		},
		{
			testDescription: "both empty",
			giveA:           []metadata.MetadataEntity{},
			giveB:           []metadata.MetadataEntity{},
			wantResult:      true,
		},
		{
			testDescription: "equal one entity",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity")},
			wantResult:      true,
		},
		{
			testDescription: "equal multiple entities same order",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
			wantResult:      true,
		},
		{
			testDescription: "equal multiple entities different order",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "pages/my_entity"), createEntity(t, "apps/my_entity")},
			wantResult:      true,
		},
		{
			testDescription: "equal multiple entities duplicated",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity")},
			wantResult:      true,
		},
		{
			testDescription: "not equal - one nil",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity")},
			wantResult:      false,
		},
		{
			testDescription: "not equal - one empty",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity")},
			giveB:           []metadata.MetadataEntity{},
			wantResult:      false,
		},
		{
			testDescription: "not equal - different entity",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity2")},
			wantResult:      false,
		},
		{
			testDescription: "not equal - different lengths",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity2")},
			wantResult:      false,
		},
		{
			testDescription: "not equal - same length duplicated entity",
			giveA:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity")},
			giveB:           []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity2")},
			wantResult:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualResult := metadata.EntitiesMatch(tc.giveA, tc.giveB)
			assert.Equal(t, tc.wantResult, actualResult)
		})
	}
}

func TestUniqueEntities(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveValue       []metadata.MetadataEntity
		wantValue       []metadata.MetadataEntity
	}{
		{
			testDescription: "nil",
		},
		{
			testDescription: "empty",
			giveValue:       []metadata.MetadataEntity{},
			wantValue:       []metadata.MetadataEntity{},
		},
		{
			testDescription: "no duplicates - in sorted order",
			giveValue:       []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
			wantValue:       []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
		},
		{
			testDescription: "no duplicates - in random order",
			giveValue:       []metadata.MetadataEntity{createEntity(t, "pages/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "datasources/my_entity")},
			wantValue:       []metadata.MetadataEntity{createEntity(t, "datasources/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
		},
		{
			testDescription: "has duplicates - in sorted order",
			giveValue:       []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "datasources/my_entity"), createEntity(t, "datasources/my_entity")},
			wantValue:       []metadata.MetadataEntity{createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "datasources/my_entity")},
		},
		{
			testDescription: "has duplicates - in random order",
			giveValue:       []metadata.MetadataEntity{createEntity(t, "datasources/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "datasources/my_entity")},
			wantValue:       []metadata.MetadataEntity{createEntity(t, "datasources/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue := metadata.UniqueEntities(tc.giveValue)
			assert.ElementsMatch(t, tc.wantValue, actualValue)
		})
	}
}

func TestMetadataEntityPaths(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveValue       []metadata.MetadataEntity
		wantValue       set.Of[string]
	}{
		{
			testDescription: "nil entities",
			giveValue:       nil,
			wantValue:       set.New[string](),
		},
		{
			testDescription: "empty entities",
			giveValue:       []metadata.MetadataEntity{},
			wantValue:       set.New[string](),
		},
		{
			testDescription: "has entities - no duplicates",
			giveValue:       []metadata.MetadataEntity{createEntity(t, "pages/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "datasources/my_entity")},
			wantValue:       set.New[string]("pages/my_entity", "apps/my_entity", "datasources/my_entity"),
		},
		{
			testDescription: "has entities - with duplicates",
			giveValue:       []metadata.MetadataEntity{createEntity(t, "datasources/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "apps/my_entity"), createEntity(t, "pages/my_entity"), createEntity(t, "datasources/my_entity")},
			wantValue:       set.New[string]("pages/my_entity", "apps/my_entity", "datasources/my_entity"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			actualValue := metadata.MetadataEntityPaths(tc.giveValue)
			assert.Equal(t, tc.wantValue, actualValue)
		})
	}

}

type MetadataEntityTestSuite struct {
	suite.Suite
}

func (suite *MetadataEntityTestSuite) TestValidate() {
	entityFullName := func(me metadata.MetadataEntity) string {
		return fmt.Sprintf("%v_SubType_%v", me.Type.Name(), me.SubType)
	}

	createEntityFixture := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, name string) metadata.MetadataEntity {
		relFilePath := name
		if mdt == metadata.MetadataTypeSite {
			if mdtst == metadata.MetadataSubTypeSiteFavicon {
				relFilePath = path.Join("favicon", relFilePath)
			} else if mdtst == metadata.MetadataSubTypeSiteLogo {
				relFilePath = path.Join("logo", relFilePath)
			}
		}
		return metadata.MetadataEntity{
			Type:         mdt,
			SubType:      mdtst,
			Name:         name,
			Path:         path.Join(mdt.DirName(), relFilePath),
			PathRelative: name,
		}
	}

	createEntityFileContentsFixture := func(me metadata.MetadataEntity, filePath string, isEntityDefinitionFile bool, contents string) metadata.MetadataEntityFileContents {
		fileName := path.Base(filePath)
		return metadata.MetadataEntityFileContents{
			MetadataEntityFile: metadata.MetadataEntityFile{
				Entity:                 me,
				Name:                   fileName,
				Path:                   path.Join(me.Type.DirName(), filePath),
				PathRelative:           filePath,
				IsEntityDefinitionFile: isEntityDefinitionFile,
			},
			Contents: []byte(contents),
		}
	}

	createStandardTestCases := func(mdt metadata.MetadataType, mdtst metadata.MetadataSubType, entityName string, testCaseSensitive bool, testDefFile bool, testTooManyFiles bool, files []ValidateFile) []ValidateTestCase {
		me := createEntityFixture(mdt, mdtst, strings.ToLower(entityName))
		var validFiles []metadata.MetadataEntityFileContents
		var emptyContentFiles []metadata.MetadataEntityFileContents
		var emptyNameValueFiles []metadata.MetadataEntityFileContents
		var mismatchedCaseNameValueFiles []metadata.MetadataEntityFileContents
		var additionalFileFiles []metadata.MetadataEntityFileContents
		var defFileOnlyFiles []metadata.MetadataEntityFileContents
		for _, f := range files {
			validContents := f.FileContent
			emptyContents := ""
			emptyNameValueContents := f.FileContent
			mismatchedCaseNameValueContents := f.FileContent
			if f.IsDefinitionFile {
				validContents = strings.Replace(f.FileContent, "__NAME__", entityName, 1)
				emptyContents = ""
				emptyNameValueContents = strings.Replace(f.FileContent, "__NAME__", "", 1)
				mismatchedCaseNameValueContents = strings.Replace(f.FileContent, "__NAME__", strings.ToUpper(entityName), 1)
				defFileOnlyFiles = append(defFileOnlyFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, validContents))
				additionalFileFiles = append(additionalFileFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, validContents))
			}
			validFiles = append(validFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, validContents))
			emptyContentFiles = append(emptyContentFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, emptyContents))
			emptyNameValueFiles = append(emptyNameValueFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, emptyNameValueContents))
			mismatchedCaseNameValueFiles = append(mismatchedCaseNameValueFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, mismatchedCaseNameValueContents))
			additionalFileFiles = append(additionalFileFiles, createEntityFileContentsFixture(me, f.FileName, f.IsDefinitionFile, validContents))
		}

		entityFullName := entityFullName(me)
		standardCases := []ValidateTestCase{
			{
				testDescription: fmt.Sprintf("%v valid", entityFullName),
				giveEntity:      me,
				giveFiles:       validFiles,
				wantError:       false,
				wantMessages:    false,
			},
			{
				testDescription: fmt.Sprintf("%v invalid nil files", entityFullName),
				giveEntity:      me,
				giveFiles:       nil,
				wantError:       false,
				wantMessages:    true,
			},
			{
				testDescription: fmt.Sprintf("%v invalid empty files", entityFullName),
				giveEntity:      me,
				giveFiles:       []metadata.MetadataEntityFileContents{},
				wantError:       false,
				wantMessages:    true,
			},
		}

		if testDefFile {
			standardCases = append(standardCases, []ValidateTestCase{
				{
					testDescription: fmt.Sprintf("%v invalid empty file content", entityFullName),
					giveEntity:      me,
					giveFiles:       emptyContentFiles,
					wantError:       false,
					wantMessages:    true,
				},
				{
					testDescription: fmt.Sprintf("%v invalid empty name value", entityFullName),
					giveEntity:      me,
					giveFiles:       emptyNameValueFiles,
					wantError:       false,
					wantMessages:    true,
				},
				{
					testDescription: fmt.Sprintf("%v invalid too many files and multiple definition files", entityFullName),
					giveEntity:      me,
					giveFiles:       additionalFileFiles,
					wantError:       false,
					wantMessages:    true,
				},
			}...)
		}

		if testCaseSensitive && testDefFile {
			standardCases = append(standardCases, []ValidateTestCase{
				{
					testDescription: fmt.Sprintf("%v invalid mismatched case name value", entityFullName),
					giveEntity:      me,
					giveFiles:       mismatchedCaseNameValueFiles,
					wantError:       false,
					wantMessages:    true,
				},
			}...)
		}

		if testTooManyFiles && len(files) > 1 {
			standardCases = append(standardCases, []ValidateTestCase{
				{
					testDescription: fmt.Sprintf("%v invalid not enough files", entityFullName),
					giveEntity:      me,
					giveFiles:       defFileOnlyFiles,
					wantError:       false,
					wantMessages:    true,
				},
			}...)
		}

		return standardCases
	}

	createAdditionalSiteTestCases := func() []ValidateTestCase {
		me := createEntityFixture(metadata.MetadataTypeSite, metadata.MetadataSubTypeNone, metadata.EntityNameSite)
		entityFullName := entityFullName(me)
		meNotASite := createEntityFixture(metadata.MetadataTypeSite, metadata.MetadataSubTypeNone, "my_site")

		return []ValidateTestCase{
			{
				testDescription: fmt.Sprintf("%v valid custom entity name", entityFullName),
				giveEntity:      me,
				giveFiles:       []metadata.MetadataEntityFileContents{createEntityFileContentsFixture(me, "site.json", true, `{"name": "This is my name"}`)},
				wantError:       false,
				wantMessages:    false,
			},
			{
				testDescription: fmt.Sprintf("%v valid single space entity name", entityFullName),
				giveEntity:      me,
				giveFiles:       []metadata.MetadataEntityFileContents{createEntityFileContentsFixture(me, "site.json", true, `{"name": " "}`)},
				wantError:       false,
				wantMessages:    false,
			},
			{
				testDescription: fmt.Sprintf("%v invalid entity name is not site", entityFullName),
				giveEntity:      meNotASite,
				giveFiles:       []metadata.MetadataEntityFileContents{createEntityFileContentsFixture(meNotASite, "site.json", true, `{"name": "my_site"}`)},
				wantError:       true,
				wantMessages:    false,
			},
		}
	}

	var testCases []ValidateTestCase
	for _, mdt := range metadata.MetadataTypes.Members() {
		switch mdt {
		case metadata.MetadataTypeFiles:
			testCases = append(testCases, createStandardTestCases(metadata.MetadataTypeFiles, metadata.MetadataSubTypeNone, "my_file.txt", true, true, true, []ValidateFile{
				{"my_file.txt.skuid.json", `{"name": "__NAME__"}`, true},
				{"my_file.txt", "", false},
			})...)
		case metadata.MetadataTypeSite:
			testCases = append(testCases, createStandardTestCases(metadata.MetadataTypeSite, metadata.MetadataSubTypeNone, metadata.EntityNameSite, false, true, true, []ValidateFile{
				{"site.json", `{"name": "__NAME__"}`, true},
			})...)
			testCases = append(testCases, createAdditionalSiteTestCases()...)
			testCases = append(testCases, createStandardTestCases(metadata.MetadataTypeSite, metadata.MetadataSubTypeSiteFavicon, "my_favicon.ico", true, true, true, []ValidateFile{
				{"favicon/my_favicon.ico.skuid.json", `{"name": "__NAME__"}`, true},
				{"favicon/my_favicon.ico", "", false},
			})...)
			testCases = append(testCases, createStandardTestCases(metadata.MetadataTypeSite, metadata.MetadataSubTypeSiteLogo, "my_logo.png", true, true, true, []ValidateFile{
				{"logo/my_logo.png.skuid.json", `{"name": "__NAME__"}`, true},
				{"logo/my_logo.png", "", false},
			})...)
		case metadata.MetadataTypeDesignSystems:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, "my_mdt", true, true, true, []ValidateFile{
				{"my_mdt.json", `{"objectData": { "name": "__NAME__"}}`, true},
			})...)
		case metadata.MetadataTypePages:
			testCases = append(testCases, createStandardTestCases(metadata.MetadataTypePages, metadata.MetadataSubTypeNone, "my_page", true, true, true, []ValidateFile{
				{"my_page.json", `{"name": "__NAME__"}`, true},
				{"my_page.xml", "", false},
			})...)
		case metadata.MetadataTypeComponentPacks:
			testCases = append(testCases, createStandardTestCases(metadata.MetadataTypeComponentPacks, metadata.MetadataSubTypeNone, "my_pack", false, false, false, []ValidateFile{
				{"mycomponents/custom_runtime.json", "", false},
				{"mycomponents/custom_builders.json", "", false},
				{"mycomponents/js/custom.js", "", false},
				{"mycomponents/css/custom.css", "", false},
			})...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, metadata.MetadataSubTypeNone, "my_mdt", true, true, true, []ValidateFile{
				{"my_mdt.json", `{"name": "__NAME__"}`, true},
			})...)
		}
	}

	for _, tc := range testCases {
		suite.Run(tc.testDescription, func() {
			t := suite.T()
			messages, err := tc.giveEntity.Validate(tc.giveFiles)
			assert.Equal(t, tc.wantError, err != nil, "err value was not expected")
			assert.Equal(t, tc.wantMessages, len(messages) > 0, "messages length was not expected")
		})
	}
}

func TestMetadataEntityTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataEntityTestSuite))
}

func createPathError(mdt metadata.MetadataType, path string) error {
	return fmt.Errorf("metadata type %v does not support the entity path: %v", mdt.Name(), logging.QuoteText(filepath.FromSlash(path)))
}

func createContainMetadataNameError(path string) error {
	return fmt.Errorf("directory name matching a valid metadata type name must exist in entity path: %v", logging.QuoteText(filepath.FromSlash(path)))
}

func createInvalidMetadataNameError(typename string, path string) error {
	return fmt.Errorf("invalid metadata type name %v for entity path: %v", logging.QuoteText(typename), logging.QuoteText(filepath.FromSlash(path)))
}

func createEntity(t *testing.T, entityPath string) metadata.MetadataEntity {
	e, err := metadata.NewMetadataEntity(entityPath)
	require.NoError(t, err)
	return *e
}
