package util_test

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGetAbs(t *testing.T) {
	wd, _ := os.Getwd()
	for _, tc := range []struct {
		description string
		given       string
		expected    string
	}{
		{
			description: "relative",
			given:       ".relative",
			expected: func() string {
				p, _ := filepath.Abs(filepath.Join(wd, ".relative"))
				return p
			}(),
		},
		{
			description: "absolute",
			given: func() string {
				return filepath.Join(wd, ".absolute")
			}(),
			expected: func() string {
				return filepath.Join(wd, ".absolute")
			}(),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actual := util.GetAbsolutePath(tc.given)
			if actual != tc.expected {
				t.Logf("actual %v not equal %v", actual, tc.expected)
				t.FailNow()
			}
		})
	}
}

func TestWindowsPath(t *testing.T) {
	for _, tc := range []struct {
		description string
		given       string
		expected    string
	}{
		{
			description: "not windows path",
			given:       "/usr/local",
			expected:    "/usr/local",
		},
		{
			description: "windows path",
			given:       `\i\hate\windows`,
			expected:    `/i/hate/windows`,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actual := util.FromWindowsPath(tc.given)
			if actual != tc.expected {
				t.Logf("actual %v not equal expected %v", actual, tc.expected)
				t.FailNow()
			}
		})
	}
}

const existingProfileBody = `{
	"name": "Admin",
	"enableSignupUi": false,
	"requireEmailVerificationOnSignup": true
}`

const messySitePermissionSetBody = `{
	"signupUi": null,
	"name": "Admin", 
	"permissionSet": {
		"dataSourcePermissions": {
			"Racer": {
				"dataSourceObjectPermissions": null
			}
		},
		"appPermissions": {
			"Admin": {
				"isDefault": false
			},
			"Racer": {
				"isDefault": false
			}
		}
	},
	"enableSignupApi": false
}`

const mergedSitePermissionSetBody = `{
	"name": "Admin",
	"enableSignupApi": false,
	"enableSignupUi": false,
	"permissionSet": {
		"appPermissions": {
			"Admin": {
				"isDefault": false
			},
			"Racer": {
				"isDefault": false
			}
		},
		"dataSourcePermissions": {
			"Racer": {}
		}
	},
	"requireEmailVerificationOnSignup": true
}`

type RetrieveFile struct {
	Name string
	Body string
}

func TestRetrieve(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveTargetDir   string
		givenNoZip      bool
		giveFiles       []RetrieveFile
		wantFiles       []RetrieveFile
		wantDirectories []string
		wantError       error
	}{
		{
			testDescription: "retrieve nothing",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       nil,
		},
		{
			testDescription: "retrieve nonvalid skuid metadata files",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"readme.txt", "This archive contains some text files."}, {"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"}, {"todo.txt", "Get animal handling licence.\nWrite more examples."}},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       nil,
		},
		{
			testDescription: "retrieve a data source",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", "this is not even close to good JSON"}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("datasources/mydatasource.json"), "this is not even close to good JSON"}},
			wantDirectories: []string{"datasources"},
			wantError:       nil,
		},
		{
			testDescription: "retrieve two data sources",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", "this is not even close to good JSON"}, {"datasources/mydatasource2.json", "this is not even close to good JSON2"}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("datasources/mydatasource.json"), "this is not even close to good JSON"}, {filepath.FromSlash("datasources/mydatasource2.json"), "this is not even close to good JSON2"}},
			wantDirectories: []string{"datasources"},
			wantError:       nil,
		},
		{
			testDescription: "retrieve a data source with targetdir",
			giveTargetDir:   "myTargetDir",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", "this is not even close to good JSON"}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("myTargetDir/datasources/mydatasource.json"), "this is not even close to good JSON"}},
			wantDirectories: []string{"myTargetDir", filepath.FromSlash("myTargetDir/datasources")},
			wantError:       nil,
		},
		{
			testDescription: "retrieve merged profile",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"sitepermissionsets/myprofile.json", existingProfileBody}, {"sitepermissionsets/myprofile.json", messySitePermissionSetBody}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("sitepermissionsets/myprofile.json"), mergedSitePermissionSetBody}},
			wantDirectories: []string{"sitepermissionsets"},
			wantError:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {

			// Create a buffer to write our archive to.
			buf := new(bytes.Buffer)

			// Create a new zip archive.
			w := zip.NewWriter(buf)

			for _, file := range tc.giveFiles {
				f, err := w.Create(file.Name)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.Write([]byte(file.Body))
				if err != nil {
					t.Fatal(err)
				}
			}

			// Make sure to check the error on Close.
			err := w.Close()
			if err != nil {
				t.Fatal(err)
			}

			var filesMap = map[string]RetrieveFile{}
			var directoriesCreated = []string{}

			var mockFileMaker = func(fileReader io.ReadCloser, path string) error {
				body, err := ioutil.ReadAll(fileReader)
				if err != nil {
					t.Fatal(err)
				}
				filesMap[path] = RetrieveFile{
					Name: path,
					Body: string(body),
				}
				return nil
			}

			var mockDirectoryMaker = func(path string, fileMode os.FileMode) error {
				if !util.StringSliceContainsKey(directoriesCreated, path) {
					directoriesCreated = append(directoriesCreated, path)
				}
				return nil
			}

			var mockExistingFileReader = func(path string) ([]byte, error) {
				return []byte(existingProfileBody), nil
			}

			err = util.WriteResultsToDiskInjection(tc.giveTargetDir, [][]byte{buf.Bytes()}, tc.givenNoZip, mockFileMaker, mockDirectoryMaker, mockExistingFileReader)

			filesCreated := []RetrieveFile{}
			for _, file := range filesMap {
				filesCreated = append(filesCreated, file)
			}
			assert.ElementsMatch(t, tc.wantFiles, filesCreated)
			assert.ElementsMatch(t, tc.wantDirectories, directoriesCreated)
			if tc.wantError != nil {
				assert.Equal(t, tc.wantError, err)
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
