package pkg_test

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/orsinium-labs/enum"
	"github.com/skuid/skuid-cli/pkg"
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
// 			plan := pkg.NlxPlan{}
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
	VALID_ENTITY_NAME      = "my _-0123456789 FILE"
	VALID_FILE_ENTITY_NAME = "my FO0123456789_-.() FILE"
)

type NewMetadataEntityTestCase struct {
	testDescription string
	givePath        string
	wantEntity      *pkg.MetadataEntity
	wantError       error
}

type NewMetadataEntityFileTestCase struct {
	testDescription string
	givePath        string
	wantEntityFile  *pkg.MetadataEntityFile
	wantError       error
}

type NlxMetadataTestSuite struct {
	suite.Suite
}

func (suite *NlxMetadataTestSuite) TestGetFieldValue() {
	metadata := pkg.NlxMetadata{
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
		given          pkg.MetadataType
		expected       []string
		wantPanicError error
	}{
		{
			description: "apps",
			given:       pkg.MetadataTypeApps,
			expected:    metadata.Apps,
		},
		{
			description: "authproviders",
			given:       pkg.MetadataTypeAuthProviders,
			expected:    metadata.AuthProviders,
		},
		{
			description: "componentpacks",
			given:       pkg.MetadataTypeComponentPacks,
			expected:    metadata.ComponentPacks,
		},
		{
			description: "dataservices",
			given:       pkg.MetadataTypeDataServices,
			expected:    metadata.DataServices,
		},
		{
			description: "datasources",
			given:       pkg.MetadataTypeDataSources,
			expected:    metadata.DataSources,
		},
		{
			description: "designsystems",
			given:       pkg.MetadataTypeDesignSystems,
			expected:    metadata.DesignSystems,
		},
		{
			description: "variables",
			given:       pkg.MetadataTypeVariables,
			expected:    metadata.Variables,
		},
		{
			description: "files",
			given:       pkg.MetadataTypeFiles,
			expected:    metadata.Files,
		},
		{
			description: "pages",
			given:       pkg.MetadataTypePages,
			expected:    metadata.Pages,
		},
		{
			description: "permissionsets",
			given:       pkg.MetadataTypePermissionSets,
			expected:    metadata.PermissionSets,
		},
		{
			description: "sitepermissionsets",
			given:       pkg.MetadataTypeSitePermissionSets,
			expected:    metadata.SitePermissionSets,
		},
		{
			description: "sessionvariables",
			given:       pkg.MetadataTypeSessionVariables,
			expected:    metadata.SessionVariables,
		},
		{
			description: "site",
			given:       pkg.MetadataTypeSite,
			expected:    metadata.Site,
		},
		{
			description: "themes",
			given:       pkg.MetadataTypeThemes,
			expected:    metadata.Themes,
		},
		{
			description:    "bad",
			given:          pkg.MetadataType{"bad"},
			wantPanicError: fmt.Errorf("unable to locate metadata field for metadata type %q", "bad"),
		},
	} {
		suite.Run(tc.description, func() {
			t := suite.T()
			if tc.wantPanicError != nil {
				assert.PanicsWithError(t, tc.wantPanicError.Error(), func() {
					metadata.GetFieldValue(tc.given)
				})
			} else {
				assert.Equal(t, tc.expected, metadata.GetFieldValue(tc.given))
			}
		})
	}
}

func (suite *NlxMetadataTestSuite) TestExistsInMetadataTypes() {
	t := suite.T()
	metadataType := reflect.TypeOf(pkg.NlxMetadata{})

	for i := 0; i < metadataType.NumField(); i++ {
		field := metadataType.Field(i)
		dirName := field.Tag.Get("json")
		assert.GreaterOrEqual(t, len(dirName), 1)
		mdt := enum.Parse(pkg.MetadataTypes, pkg.MetadataTypeValue(field.Name))
		require.NotNil(t, mdt)
		assert.Equal(t, dirName, mdt.DirName())
	}
}

func (suite *NlxMetadataTestSuite) TestFilterItem() {
	testCases := []struct {
		testDescription string
		giveMetadata    pkg.NlxMetadata
		giveFile        pkg.MetadataEntityFile
		wantResult      bool
		wantPanicError  error
	}{
		{
			testDescription: "exists in metadata",
			giveMetadata: pkg.NlxMetadata{
				Apps: []string{"my_app"},
			},
			giveFile: pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeApps,
					PathRelative: "my_app",
				},
			},
			wantResult: true,
		},
		{
			testDescription: "does not exist in metadata",
			giveMetadata:    pkg.NlxMetadata{},
			giveFile: pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeApps,
					PathRelative: "not_there",
				},
			},
			wantResult: false,
		},
		{
			testDescription: "metadatatype not found in metadata",
			giveFile: pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type: pkg.MetadataType{"bad"},
				},
			},
			wantPanicError: fmt.Errorf("unable to locate metadata field for metadata type %q", "bad"),
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
		giveSource      pkg.MetadataTypeValue
		giveTarget      pkg.MetadataTypeValue
		wantResult      bool
	}{
		{
			testDescription: "equals both lower",
			giveSource:      pkg.MetadataTypeValue("apps"),
			giveTarget:      pkg.MetadataTypeValue("apps"),
			wantResult:      true,
		},
		{
			testDescription: "equals both upper",
			giveSource:      pkg.MetadataTypeValue("APPS"),
			giveTarget:      pkg.MetadataTypeValue("APPS"),
			wantResult:      true,
		},
		{
			testDescription: "equals both mixed",
			giveSource:      pkg.MetadataTypeValue("AppS"),
			giveTarget:      pkg.MetadataTypeValue("AppS"),
			wantResult:      true,
		},
		{
			testDescription: "equals lower and upper",
			giveSource:      pkg.MetadataTypeValue("apps"),
			giveTarget:      pkg.MetadataTypeValue("APPS"),
			wantResult:      true,
		},
		{
			testDescription: "equals upper and lower",
			giveSource:      pkg.MetadataTypeValue("APPS"),
			giveTarget:      pkg.MetadataTypeValue("apps"),
			wantResult:      true,
		},
		{
			testDescription: "not equals",
			giveSource:      pkg.MetadataTypeValue("foo"),
			giveTarget:      pkg.MetadataTypeValue("bar"),
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
			mdt := pkg.MetadataType{pkg.MetadataTypeValue(tc.giveName)}
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
			mdt := pkg.MetadataType{pkg.MetadataTypeValue(tc.giveName)}
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
	metadataType := reflect.TypeOf(pkg.NlxMetadata{})
	for _, mdt := range pkg.MetadataTypes.Members() {
		f, ok := metadataType.FieldByName(mdt.Name())
		require.True(t, ok)
		assert.Equal(t, mdt.DirName(), f.Tag.Get("json"))
	}
}

func (suite *MetadataTypesTestSuite) TestNameValid() {
	t := suite.T()
	for _, mdt := range pkg.MetadataTypes.Members() {
		assert.GreaterOrEqual(t, len(strings.TrimSpace(mdt.Name())), 1)
	}
}

func (suite *MetadataTypesTestSuite) TestDirNameValid() {
	t := suite.T()
	for _, mdt := range pkg.MetadataTypes.Members() {
		assert.GreaterOrEqual(t, len(strings.TrimSpace(mdt.DirName())), 1)
	}
}

func TestMetadataTypesTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataTypesTestSuite))
}

func TestNewMetadataEntity(t *testing.T) {
	createFixture := func(mdt pkg.MetadataType, name string) *pkg.MetadataEntity {
		return &pkg.MetadataEntity{
			Type:         mdt,
			Name:         name,
			Path:         mdt.DirName() + "/" + name,
			PathRelative: name,
		}
	}

	createStandardValidTestCases := func(mdt pkg.MetadataType, entityName string) []NewMetadataEntityTestCase {
		return []NewMetadataEntityTestCase{
			{
				testDescription: fmt.Sprintf("%v valid", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v", mdt.DirName(), entityName),
				wantEntity:      createFixture(mdt, entityName),
			},
			{
				testDescription: fmt.Sprintf("%v valid . relative", mdt.Name()),
				givePath:        fmt.Sprintf("./%v/%v", mdt.DirName(), entityName),
				wantEntity:      createFixture(mdt, entityName),
			},
		}
	}

	createStandardInvalidTestCases := func(mdt pkg.MetadataType, entityName string) []NewMetadataEntityTestCase {
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

	createStandardTestCases := func(mdt pkg.MetadataType, entityName string) []NewMetadataEntityTestCase {
		return append(createStandardValidTestCases(mdt, entityName), createStandardInvalidTestCases(mdt, entityName)...)
	}

	filesTestCases := []NewMetadataEntityTestCase{
		{
			testDescription: "Files valid no extension",
			givePath:        "files/" + VALID_FILE_ENTITY_NAME,
			wantEntity:      createFixture(pkg.MetadataTypeFiles, VALID_FILE_ENTITY_NAME),
		},
		{
			testDescription: "Files valid contains json extension",
			givePath:        "files/md_file.json",
			wantEntity:      createFixture(pkg.MetadataTypeFiles, "md_file.json"),
		},
		{
			testDescription: "Files valid contains skuid.json extension",
			givePath:        "files/md_file.skuid.json",
			wantEntity:      createFixture(pkg.MetadataTypeFiles, "md_file.skuid.json"),
		},
		{
			testDescription: "Files valid contains txt filename",
			givePath:        "files/md_file.txt",
			wantEntity:      createFixture(pkg.MetadataTypeFiles, "md_file.txt"),
		},
		{
			testDescription: "Files valid contains xml filename",
			givePath:        "files/md_file.xml",
			wantEntity:      createFixture(pkg.MetadataTypeFiles, "md_file.xml"),
		},
		{
			testDescription: "Files invalid contains nested no extension",
			givePath:        "files/subdir/md_file",
			wantError:       createPathError(pkg.MetadataTypeFiles, "files/subdir/md_file"),
		},
		{
			testDescription: "Files invalid contains nested with extension",
			givePath:        "files/subdir/md_file.txt",
			wantError:       createPathError(pkg.MetadataTypeFiles, "files/subdir/md_file.txt"),
		},
		{
			testDescription: "Files invalid contains invalid $ character",
			givePath:        "files/pre$%v",
			wantError:       createPathError(pkg.MetadataTypeFiles, "files/pre$%v"),
		},
		{
			testDescription: "Files invalid contains invalid $ character",
			givePath:        "files/pre^%v",
			wantError:       createPathError(pkg.MetadataTypeFiles, "files/pre^%v"),
		},
	}

	siteTestCases := []NewMetadataEntityTestCase{
		{
			testDescription: "Site favicon valid",
			givePath:        "site/favicon/my_favicon.ico",
			wantEntity: &pkg.MetadataEntity{
				Type:         pkg.MetadataTypeSite,
				Name:         "my_favicon.ico",
				Path:         "site/favicon/my_favicon.ico",
				PathRelative: "favicon/my_favicon.ico",
			},
		},
		{
			testDescription: "Site favicon invalid no extension",
			givePath:        "site/favicon/my_favicon",
			wantError:       createPathError(pkg.MetadataTypeSite, "site/favicon/my_favicon"),
		},
		{
			testDescription: "Site favicon invalid contains extension",
			givePath:        "site/favicon/my_favicon.ico.skuid.json",
			wantError:       createPathError(pkg.MetadataTypeSite, "site/favicon/my_favicon.ico.skuid.json"),
		},
		{
			testDescription: "Site logo invalid no extension",
			givePath:        "site/logo/my_logo",
			wantError:       createPathError(pkg.MetadataTypeSite, "site/logo/my_logo"),
		},
		{
			testDescription: "Site logo valid png",
			givePath:        "site/logo/my_logo.png",
			wantEntity: &pkg.MetadataEntity{
				Type:         pkg.MetadataTypeSite,
				Name:         "my_logo.png",
				Path:         "site/logo/my_logo.png",
				PathRelative: "logo/my_logo.png",
			},
		},
		{
			testDescription: "Site logo invalid png contains extension",
			givePath:        "site/logo/my_logo.png.skuid.json",
			wantError:       createPathError(pkg.MetadataTypeSite, "site/logo/my_logo.png.skuid.json"),
		},
		{
			testDescription: "Site logo valid jpg",
			givePath:        "site/logo/my_logo.jpg",
			wantEntity: &pkg.MetadataEntity{
				Type:         pkg.MetadataTypeSite,
				Name:         "my_logo.jpg",
				Path:         "site/logo/my_logo.jpg",
				PathRelative: "logo/my_logo.jpg",
			},
		},
		{
			testDescription: "Site logo invalid jpg contains extension",
			givePath:        "site/logo/my_logo.jpg.skuid.json",
			wantError:       createPathError(pkg.MetadataTypeSite, "site/logo/my_logo.jpg.skuid.json"),
		},
		{
			testDescription: "Site logo valid gif",
			givePath:        "site/logo/my_logo.gif",
			wantEntity: &pkg.MetadataEntity{
				Type:         pkg.MetadataTypeSite,
				Name:         "my_logo.gif",
				Path:         "site/logo/my_logo.gif",
				PathRelative: "logo/my_logo.gif",
			},
		},
		{
			testDescription: "Site logo invalid gif contains extension",
			givePath:        "site/logo/my_logo.gif.skuid.json",
			wantError:       createPathError(pkg.MetadataTypeSite, "site/logo/my_logo.gif.skuid.json"),
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

	for _, mdt := range pkg.MetadataTypes.Members() {
		switch mdt {
		case pkg.MetadataTypeFiles:
			testCases = append(testCases, filesTestCases...)
		case pkg.MetadataTypeSite:
			testCases = append(testCases, createStandardTestCases(mdt, "site")...)
			testCases = append(testCases, siteTestCases...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, VALID_ENTITY_NAME)...)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			// Note - Forcing givePath to platform file separator to simulate reading from disk
			entity, err := pkg.NewMetadataEntity(filepath.FromSlash(tc.givePath))
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
	createFixture := func(mdt pkg.MetadataType, name string, fileName string, isEntityDefinitionFile bool) *pkg.MetadataEntityFile {
		return &pkg.MetadataEntityFile{
			Entity: pkg.MetadataEntity{
				Type:         mdt,
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

	createStandardValidTestCases := func(mdt pkg.MetadataType, entityName string) []NewMetadataEntityFileTestCase {
		return []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v valid json extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/%v.json", mdt.DirName(), entityName),
				wantEntityFile:  createFixture(mdt, entityName, entityName+".json", true),
			},
			{
				testDescription: fmt.Sprintf("%v valid . relative", mdt.Name()),
				givePath:        fmt.Sprintf("./%v/%v.json", mdt.DirName(), entityName),
				wantEntityFile:  createFixture(mdt, entityName, entityName+".json", true),
			},
		}
	}

	createStandardInvalidTestCases := func(mdt pkg.MetadataType, entityName string, includeNestedDirTests bool, includeXmlTests bool) []NewMetadataEntityFileTestCase {
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

	createStandardTestCases := func(mdt pkg.MetadataType, entityName string) []NewMetadataEntityFileTestCase {
		return append(createStandardValidTestCases(mdt, entityName), createStandardInvalidTestCases(mdt, entityName, true, true)...)
	}

	themesTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: "Themes valid .inline.css extension",
			givePath:        fmt.Sprintf("%v/%v.inline.css", pkg.MetadataTypeThemes.DirName(), VALID_ENTITY_NAME),
			wantEntityFile:  createFixture(pkg.MetadataTypeThemes, VALID_ENTITY_NAME, VALID_ENTITY_NAME+".inline.css", false),
		},
	}

	componentPacksTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: fmt.Sprintf("%v valid nested dir json extension", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".json",
				Path:                   fmt.Sprintf("%v/subdir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.json", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir js extension", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.js", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".js",
				Path:                   fmt.Sprintf("%v/subdir/%v.js", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.js", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir css extension", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.css", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".css",
				Path:                   fmt.Sprintf("%v/subdir/%v.css", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.css", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir no extension", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME,
				Path:                   fmt.Sprintf("%v/subdir/%v", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir . relative json extension", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("./%v/subdir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &pkg.MetadataEntityFile{
				Entity: pkg.MetadataEntity{
					Type:         pkg.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".json",
				Path:                   fmt.Sprintf("%v/subdir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.json", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v invalid json file in metadata dir root", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(pkg.MetadataTypeComponentPacks, fmt.Sprintf("%v/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid . relative to componentpacks directory", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("./%v/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(pkg.MetadataTypeComponentPacks, fmt.Sprintf("./%v/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains invalid $ character", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/sub$dir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(pkg.MetadataTypeComponentPacks, fmt.Sprintf("%v/sub$dir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains invalid ^ character", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/sub^dir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(pkg.MetadataTypeComponentPacks, fmt.Sprintf("%v/sub^dir/%v.json", pkg.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains no filename", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir", pkg.MetadataTypeComponentPacks.DirName()),
			wantError:       createPathError(pkg.MetadataTypeComponentPacks, fmt.Sprintf("%v/subdir", pkg.MetadataTypeComponentPacks.DirName())),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains no filename ends with slash", pkg.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/", pkg.MetadataTypeComponentPacks.DirName()),
			wantError:       createPathError(pkg.MetadataTypeComponentPacks, fmt.Sprintf("%v/subdir/", pkg.MetadataTypeComponentPacks.DirName())),
		},
	}

	pagesTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: fmt.Sprintf("%v valid xml extension", pkg.MetadataTypePages.Name()),
			givePath:        fmt.Sprintf("%v/%v.xml", pkg.MetadataTypePages.DirName(), VALID_ENTITY_NAME),
			wantEntityFile:  createFixture(pkg.MetadataTypePages, VALID_ENTITY_NAME, VALID_ENTITY_NAME+".xml", false),
		},
	}

	createSiteImageTestCases := func(subdir string, extensions []string) []NewMetadataEntityFileTestCase {
		mdt := pkg.MetadataTypeSite

		testCases := []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v %v invalid no extension", mdt.Name(), subdir),
				givePath:        fmt.Sprintf("%v/%v/%v.xml", mdt.DirName(), subdir, VALID_FILE_ENTITY_NAME),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/%v/%v.xml", mdt.DirName(), subdir, VALID_FILE_ENTITY_NAME)),
			},
		}

		for _, ext := range extensions {
			entityName := VALID_FILE_ENTITY_NAME + ext
			testCases = append(testCases, []NewMetadataEntityFileTestCase{
				{
					testDescription: fmt.Sprintf("%v %v valid %v extension", mdt.Name(), subdir, ext),
					givePath:        fmt.Sprintf("%v/%v/%v", mdt.DirName(), subdir, entityName),
					wantEntityFile: &pkg.MetadataEntityFile{
						Entity: pkg.MetadataEntity{
							Type:         mdt,
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
					wantEntityFile: &pkg.MetadataEntityFile{
						Entity: pkg.MetadataEntity{
							Type:         mdt,
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
		mdt := pkg.MetadataTypeFiles
		testCases := []NewMetadataEntityFileTestCase{
			{
				testDescription: fmt.Sprintf("%v invalid contains nested no extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/subdir/%v", mdt.DirName(), VALID_FILE_ENTITY_NAME),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v", mdt.DirName(), VALID_FILE_ENTITY_NAME)),
			},
			{
				testDescription: fmt.Sprintf("%v invalid contains nested json extension", mdt.Name()),
				givePath:        fmt.Sprintf("%v/subdir/%v.json", mdt.DirName(), VALID_FILE_ENTITY_NAME),
				wantError:       createPathError(mdt, fmt.Sprintf("%v/subdir/%v.json", mdt.DirName(), VALID_FILE_ENTITY_NAME)),
			},
		}

		extensions := []string{"", ".xml", ".json", ".txt", ".jpg", ".gif", ".png", ".js", ".css", ".log", ".txt.log"}
		for _, ext := range extensions {
			entityName := VALID_FILE_ENTITY_NAME + ext
			testCases = append(testCases, []NewMetadataEntityFileTestCase{
				{
					testDescription: fmt.Sprintf("%v valid %v extension", mdt.Name(), ext),
					givePath:        fmt.Sprintf("%v/%v", mdt.DirName(), entityName),
					wantEntityFile: &pkg.MetadataEntityFile{
						Entity: pkg.MetadataEntity{
							Type:         mdt,
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
					wantEntityFile: &pkg.MetadataEntityFile{
						Entity: pkg.MetadataEntity{
							Type:         mdt,
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

	for _, mdt := range pkg.MetadataTypes.Members() {
		switch mdt {
		case pkg.MetadataTypeFiles:
			testCases = append(testCases, filesTestCases...)
		case pkg.MetadataTypeSite:
			testCases = append(testCases, createStandardTestCases(mdt, "site")...)
			testCases = append(testCases, createStandardInvalidTestCases(mdt, VALID_ENTITY_NAME, true, true)...)
			testCases = append(testCases, createSiteImageTestCases("favicon", []string{".ico"})...)
			testCases = append(testCases, createSiteImageTestCases("logo", []string{".jpg", ".png", ".gif"})...)
		case pkg.MetadataTypePages:
			testCases = append(testCases, createStandardValidTestCases(mdt, VALID_ENTITY_NAME)...)
			testCases = append(testCases, createStandardInvalidTestCases(mdt, VALID_ENTITY_NAME, true, false)...)
			testCases = append(testCases, pagesTestCases...)
		case pkg.MetadataTypeComponentPacks:
			testCases = append(testCases, createStandardInvalidTestCases(mdt, VALID_ENTITY_NAME, false, true)...)
			testCases = append(testCases, componentPacksTestCases...)
		case pkg.MetadataTypeThemes:
			testCases = append(testCases, createStandardTestCases(mdt, VALID_ENTITY_NAME)...)
			testCases = append(testCases, themesTestCases...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, VALID_ENTITY_NAME)...)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			// Note - Forcing givePath to platform file separator to simulate reading from disk
			entityFile, err := pkg.NewMetadataEntityFile(filepath.FromSlash(tc.givePath))
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error(), "expected not nil err, got nil for path %v", tc.givePath)
			} else {
				assert.NoError(t, err, "expected nil err, got not nil for path %q", tc.givePath)
			}
			assert.Equal(t, tc.wantEntityFile, entityFile)
		})
	}
}

func createPathError(mdt pkg.MetadataType, path string) error {
	return fmt.Errorf("metadata type %q does not support the entity path: %q", mdt.Name(), path)
}

func createContainMetadataNameError(path string) error {
	return fmt.Errorf("must contain a metadata type name: %q", path)
}

func createInvalidMetadataNameError(typename string, path string) error {
	return fmt.Errorf("invalid metadata name %q for entity path: %q", typename, path)
}
