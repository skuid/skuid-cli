package util_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/tides/pkg/util"
)

const badJson = "this is not even close to good JSON"

const goodJson = `{
	"good": "json"
}`

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

func TestWrite(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveTargetDir   string
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
			giveFiles:       []RetrieveFile{{"readme.txt", "This archive contains some text files"}, {"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"}, {"todo.txt", "Get animal handling licence.\nWrite more examples"}},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       nil,
		},
		{
			testDescription: "bad json",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", badJson}},
			wantDirectories: []string{"datasources"},
			wantError:       fmt.Errorf("invalid character 'h' in literal true (expecting 'r')"),
		},
		{
			testDescription: "retrieve a data source",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("datasources/mydatasource.json"), goodJson}},
			wantDirectories: []string{"datasources"},
			wantError:       nil,
		},
		{
			testDescription: "retrieve two data sources",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}, {"datasources/mydatasource2.json", strings.ReplaceAll(goodJson, "json", "json2")}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("datasources/mydatasource.json"), goodJson}, {filepath.FromSlash("datasources/mydatasource2.json"), strings.ReplaceAll(goodJson, "json", "json2")}},
			wantDirectories: []string{"datasources"},
			wantError:       nil,
		},
		{
			testDescription: "retrieve a data source with targetdir",
			giveTargetDir:   "myTargetDir",
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}},
			wantFiles:       []RetrieveFile{{filepath.FromSlash("myTargetDir/datasources/mydatasource.json"), goodJson}},
			wantDirectories: []string{"myTargetDir", filepath.FromSlash("myTargetDir/datasources")},
			wantError:       nil,
		},
		{
			testDescription: "retrieve merged profile",
			giveTargetDir:   "",
			giveFiles: []RetrieveFile{
				{"sitepermissionsets/myprofile.json", existingProfileBody},
				{"sitepermissionsets/myprofile.json", messySitePermissionSetBody},
			},
			wantFiles: []RetrieveFile{
				{filepath.FromSlash("sitepermissionsets/myprofile.json"), mergedSitePermissionSetBody},
			},
			wantDirectories: []string{
				"sitepermissionsets",
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			util.ResetPathMap()

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

			err = util.WriteResults(tc.giveTargetDir, util.WritePayload{PlanData: buf.Bytes(), PlanName: "test"}, mockFileMaker, mockDirectoryMaker, mockExistingFileReader)
			if tc.wantError != nil {
				assert.Equal(t, tc.wantError.Error(), err.Error())
			} else if err != nil {
				t.Fatal(err)
			}

			filesCreated := []RetrieveFile{}
			for _, file := range filesMap {
				filesCreated = append(filesCreated, file)
			}
			assert.ElementsMatch(t, tc.wantFiles, filesCreated)
			assert.ElementsMatch(t, tc.wantDirectories, directoriesCreated)
		})
	}
}
