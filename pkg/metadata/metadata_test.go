package metadata_test

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/orsinium-labs/enum"
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
	VALID_ENTITY_NAME      = "my _-0123456789 FILE"
	VALID_FILE_ENTITY_NAME = "my FO0123456789_-.() FILE"
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
			wantPanicError: fmt.Errorf("unable to locate metadata field for metadata type %q", "bad"),
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
	createFixture := func(mdt metadata.MetadataType, name string) *metadata.MetadataEntity {
		return &metadata.MetadataEntity{
			Type:         mdt,
			Name:         name,
			Path:         mdt.DirName() + "/" + name,
			PathRelative: name,
		}
	}

	createStandardValidTestCases := func(mdt metadata.MetadataType, entityName string) []NewMetadataEntityTestCase {
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

	createStandardTestCases := func(mdt metadata.MetadataType, entityName string) []NewMetadataEntityTestCase {
		return append(createStandardValidTestCases(mdt, entityName), createStandardInvalidTestCases(mdt, entityName)...)
	}

	filesTestCases := []NewMetadataEntityTestCase{
		{
			testDescription: "Files valid no extension",
			givePath:        "files/" + VALID_FILE_ENTITY_NAME,
			wantEntity:      createFixture(metadata.MetadataTypeFiles, VALID_FILE_ENTITY_NAME),
		},
		{
			testDescription: "Files valid contains json extension",
			givePath:        "files/md_file.json",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, "md_file.json"),
		},
		{
			testDescription: "Files valid contains skuid.json extension",
			givePath:        "files/md_file.skuid.json",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, "md_file.skuid.json"),
		},
		{
			testDescription: "Files valid contains txt filename",
			givePath:        "files/md_file.txt",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, "md_file.txt"),
		},
		{
			testDescription: "Files valid contains xml filename",
			givePath:        "files/md_file.xml",
			wantEntity:      createFixture(metadata.MetadataTypeFiles, "md_file.xml"),
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
			testCases = append(testCases, createStandardTestCases(mdt, "site")...)
			testCases = append(testCases, siteTestCases...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, VALID_ENTITY_NAME)...)
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
	createFixture := func(mdt metadata.MetadataType, name string, fileName string, isEntityDefinitionFile bool) *metadata.MetadataEntityFile {
		return &metadata.MetadataEntityFile{
			Entity: metadata.MetadataEntity{
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

	createStandardValidTestCases := func(mdt metadata.MetadataType, entityName string) []NewMetadataEntityFileTestCase {
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

	createStandardTestCases := func(mdt metadata.MetadataType, entityName string) []NewMetadataEntityFileTestCase {
		return append(createStandardValidTestCases(mdt, entityName), createStandardInvalidTestCases(mdt, entityName, true, true)...)
	}

	themesTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: "Themes valid .inline.css extension",
			givePath:        fmt.Sprintf("%v/%v.inline.css", metadata.MetadataTypeThemes.DirName(), VALID_ENTITY_NAME),
			wantEntityFile:  createFixture(metadata.MetadataTypeThemes, VALID_ENTITY_NAME, VALID_ENTITY_NAME+".inline.css", false),
		},
	}

	componentPacksTestCases := []NewMetadataEntityFileTestCase{
		{
			testDescription: fmt.Sprintf("%v valid nested dir json extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".json",
				Path:                   fmt.Sprintf("%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.json", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir js extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.js", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".js",
				Path:                   fmt.Sprintf("%v/subdir/%v.js", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.js", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir css extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v.css", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".css",
				Path:                   fmt.Sprintf("%v/subdir/%v.css", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.css", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir no extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/subdir/%v", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME,
				Path:                   fmt.Sprintf("%v/subdir/%v", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v valid nested dir . relative json extension", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("./%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantEntityFile: &metadata.MetadataEntityFile{
				Entity: metadata.MetadataEntity{
					Type:         metadata.MetadataTypeComponentPacks,
					Name:         "subdir",
					Path:         "componentpacks/subdir",
					PathRelative: "subdir",
				},
				Name:                   VALID_ENTITY_NAME + ".json",
				Path:                   fmt.Sprintf("%v/subdir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
				PathRelative:           fmt.Sprintf("subdir/%v.json", VALID_ENTITY_NAME),
				IsEntityDefinitionFile: false,
			},
		},
		{
			testDescription: fmt.Sprintf("%v invalid json file in metadata dir root", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid . relative to componentpacks directory", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("./%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("./%v/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains invalid $ character", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/sub$dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/sub$dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
		},
		{
			testDescription: fmt.Sprintf("%v invalid nested dir contains invalid ^ character", metadata.MetadataTypeComponentPacks.Name()),
			givePath:        fmt.Sprintf("%v/sub^dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME),
			wantError:       createPathError(metadata.MetadataTypeComponentPacks, fmt.Sprintf("%v/sub^dir/%v.json", metadata.MetadataTypeComponentPacks.DirName(), VALID_ENTITY_NAME)),
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
			givePath:        fmt.Sprintf("%v/%v.xml", metadata.MetadataTypePages.DirName(), VALID_ENTITY_NAME),
			wantEntityFile:  createFixture(metadata.MetadataTypePages, VALID_ENTITY_NAME, VALID_ENTITY_NAME+".xml", false),
		},
	}

	createSiteImageTestCases := func(subdir string, extensions []string) []NewMetadataEntityFileTestCase {
		mdt := metadata.MetadataTypeSite

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
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
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
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
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
		mdt := metadata.MetadataTypeFiles
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
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
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
					wantEntityFile: &metadata.MetadataEntityFile{
						Entity: metadata.MetadataEntity{
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

	for _, mdt := range metadata.MetadataTypes.Members() {
		switch mdt {
		case metadata.MetadataTypeFiles:
			testCases = append(testCases, filesTestCases...)
		case metadata.MetadataTypeSite:
			testCases = append(testCases, createStandardTestCases(mdt, "site")...)
			testCases = append(testCases, createStandardInvalidTestCases(mdt, VALID_ENTITY_NAME, true, true)...)
			testCases = append(testCases, createSiteImageTestCases("favicon", []string{".ico"})...)
			testCases = append(testCases, createSiteImageTestCases("logo", []string{".jpg", ".png", ".gif"})...)
		case metadata.MetadataTypePages:
			testCases = append(testCases, createStandardValidTestCases(mdt, VALID_ENTITY_NAME)...)
			testCases = append(testCases, createStandardInvalidTestCases(mdt, VALID_ENTITY_NAME, true, false)...)
			testCases = append(testCases, pagesTestCases...)
		case metadata.MetadataTypeComponentPacks:
			testCases = append(testCases, createStandardInvalidTestCases(mdt, VALID_ENTITY_NAME, false, true)...)
			testCases = append(testCases, componentPacksTestCases...)
		case metadata.MetadataTypeThemes:
			testCases = append(testCases, createStandardTestCases(mdt, VALID_ENTITY_NAME)...)
			testCases = append(testCases, themesTestCases...)
		default:
			testCases = append(testCases, createStandardTestCases(mdt, VALID_ENTITY_NAME)...)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			// Note - Forcing givePath to platform file separator to simulate reading from disk
			entityFile, err := metadata.NewMetadataEntityFile(filepath.FromSlash(tc.givePath))
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error(), "expected not nil err, got nil for path %v", tc.givePath)
			} else {
				assert.NoError(t, err, "expected nil err, got not nil for path %q", tc.givePath)
			}
			assert.Equal(t, tc.wantEntityFile, entityFile)
		})
	}
}

func createPathError(mdt metadata.MetadataType, path string) error {
	return fmt.Errorf("metadata type %q does not support the entity path: %q", mdt.Name(), filepath.FromSlash(path))
}

func createContainMetadataNameError(path string) error {
	return fmt.Errorf("must contain a metadata type name: %q", filepath.FromSlash(path))
}

func createInvalidMetadataNameError(typename string, path string) error {
	return fmt.Errorf("invalid metadata name %q for entity path: %q", typename, filepath.FromSlash(path))
}
