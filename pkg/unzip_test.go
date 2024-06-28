package pkg_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/util"
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

const targetDir = "myTargetDir"

type RetrieveFile struct {
	Name string
	Body string
}

func TestWriteResultsToDisk(t *testing.T) {
	curDir, err := filepath.Abs(targetDir)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		testDescription string
		giveTargetDir   string
		giveFiles       []RetrieveFile
		wantFiles       []RetrieveFile
		wantDirectories []string
		wantError       error
	}{
		{
			testDescription: "no targetDirectory specified",
			giveTargetDir:   "",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       fmt.Errorf("targetDirectory must be an absolute path"),
		},
		{
			testDescription: "whitespace targetDirectory specified",
			giveTargetDir:   "    ",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       fmt.Errorf("targetDirectory must be an absolute path"),
		},
		{
			testDescription: "current directory targetDirectory specified",
			giveTargetDir:   ".",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       fmt.Errorf("targetDirectory must be an absolute path"),
		},
		{
			testDescription: "non-absolute targetDirectory specified",
			giveTargetDir:   "someDir",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       fmt.Errorf("targetDirectory must be an absolute path"),
		},
		{
			testDescription: "non-absolute relative to current directory targetDirectory specified",
			giveTargetDir:   "./anotherDir",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
			wantError:       fmt.Errorf("targetDirectory must be an absolute path"),
		},
		{
			testDescription: "retrieve nothing",
			giveTargetDir:   curDir,
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{curDir},
			wantError:       nil,
		},
		{
			testDescription: "retrieve nonvalid skuid metadata files",
			giveTargetDir:   curDir,
			giveFiles:       []RetrieveFile{{"invalidfolder/myfile.txt", "This file is in non-metadata folder"}, {"readme.txt", "This archive contains some text files"}, {"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"}, {"todo.txt", "Get animal handling licence.\nWrite more examples"}},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{curDir},
			wantError:       nil,
		},
		{
			testDescription: "retrieve valid and nonvalid skuid metadata files",
			giveTargetDir:   curDir,
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}, {"invalidfolder/myfile.txt", "This file is in non-metadata folder"}, {"readme.txt", "This archive contains some text files"}, {"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"}, {"todo.txt", "Get animal handling licence.\nWrite more examples"}, {"datasources/mydatasource1.json", goodJson}},
			wantFiles:       []RetrieveFile{{filepath.Join(curDir, filepath.FromSlash("datasources/mydatasource.json")), goodJson}, {filepath.Join(curDir, filepath.FromSlash("datasources/mydatasource1.json")), goodJson}},
			wantDirectories: []string{curDir, filepath.Join(curDir, "datasources")},
			wantError:       nil,
		},

		{
			testDescription: "bad json",
			giveTargetDir:   curDir,
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", badJson}},
			wantDirectories: []string{curDir, filepath.Join(curDir, "datasources")},
			wantError:       fmt.Errorf("invalid character 'h' in literal true (expecting 'r')"),
		},
		{
			testDescription: "retrieve a data source",
			giveTargetDir:   curDir,
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}},
			wantFiles:       []RetrieveFile{{filepath.Join(curDir, filepath.FromSlash("datasources/mydatasource.json")), goodJson}},
			wantDirectories: []string{curDir, filepath.Join(curDir, "datasources")},
			wantError:       nil,
		},
		{
			testDescription: "retrieve two data sources",
			giveTargetDir:   curDir,
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}, {"datasources/mydatasource2.json", strings.ReplaceAll(goodJson, "json", "json2")}},
			wantFiles:       []RetrieveFile{{filepath.Join(curDir, filepath.FromSlash("datasources/mydatasource.json")), goodJson}, {filepath.Join(curDir, filepath.FromSlash("datasources/mydatasource2.json")), strings.ReplaceAll(goodJson, "json", "json2")}},
			wantDirectories: []string{curDir, filepath.Join(curDir, "datasources")},
			wantError:       nil,
		},
		{
			testDescription: "retrieve a data source with nested targetdir",
			giveTargetDir:   filepath.Join(curDir, "myTargetDirSubfolder"),
			giveFiles:       []RetrieveFile{{"datasources/mydatasource.json", goodJson}},
			wantFiles:       []RetrieveFile{{filepath.Join(curDir, filepath.FromSlash("myTargetDirSubfolder/datasources/mydatasource.json")), goodJson}},
			wantDirectories: []string{filepath.Join(curDir, "myTargetDirSubfolder"), filepath.Join(curDir, filepath.FromSlash("myTargetDirSubfolder/datasources"))},
			wantError:       nil,
		},
		{
			testDescription: "retrieve merged profile",
			giveTargetDir:   curDir,
			giveFiles: []RetrieveFile{
				{"sitepermissionsets/myprofile.json", existingProfileBody},
				{"sitepermissionsets/myprofile.json", messySitePermissionSetBody},
			},
			wantFiles: []RetrieveFile{
				{filepath.Join(curDir, filepath.FromSlash("sitepermissionsets/myprofile.json")), mergedSitePermissionSetBody},
			},
			wantDirectories: []string{
				curDir,
				filepath.Join(curDir, "sitepermissionsets"),
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			pkg.ResetPathMap()

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
				body, err := io.ReadAll(fileReader)
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

			err = pkg.WriteResults(tc.giveTargetDir, pkg.WritePayload{PlanData: buf.Bytes(), PlanName: "test"}, mockFileMaker, mockDirectoryMaker, mockExistingFileReader)
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
