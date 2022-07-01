package cmd

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/skuid-cli/types"
	"github.com/stretchr/testify/assert"
)

const existingProfileBody = `{
	"enableSignupUi": false,
	"name": "Admin",
	"requireEmailVerificationOnSignup": true
}`

const messyProfileBody = `{
	"name": "Admin",
	"signupUi": null,
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

const mergedProfileBody = `{
	"enableSignupApi": false,
	"enableSignupUi": false,
	"name": "Admin",
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
		giveFiles       []RetrieveFile
		wantFiles       []RetrieveFile
		wantDirectories []string
		wantError       error
	}{
		{
			testDescription: "retrieve nothing",
			giveFiles:       []RetrieveFile{},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
		},
		{
			testDescription: "retrieve nonvalid skuid metadata files",
			giveFiles: []RetrieveFile{
				{"readme.txt", "This archive contains some text files."},
				{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
				{"todo.txt", "Get animal handling licence.\nWrite more examples."},
			},
			wantFiles:       []RetrieveFile{},
			wantDirectories: []string{},
		},
		{
			testDescription: "retrieve a data source",
			giveFiles: []RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
			},
			wantFiles: []RetrieveFile{
				{filepath.FromSlash("datasources/mydatasource.json"), "this is not even close to good JSON"},
			},
			wantDirectories: []string{
				"datasources",
			},
		},
		{
			testDescription: "retrieve two data sources",
			giveFiles: []RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
				{"datasources/mydatasource2.json", "this is not even close to good JSON2"},
			},
			wantFiles: []RetrieveFile{
				{filepath.FromSlash("datasources/mydatasource.json"), "this is not even close to good JSON"},
				{filepath.FromSlash("datasources/mydatasource2.json"), "this is not even close to good JSON2"},
			},
			wantDirectories: []string{
				"datasources",
			},
		},
		{
			testDescription: "retrieve a data source with targetdir",
			giveTargetDir:   "myTargetDir",
			giveFiles: []RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
			},
			wantFiles: []RetrieveFile{
				{filepath.FromSlash("myTargetDir/datasources/mydatasource.json"), "this is not even close to good JSON"},
			},
			wantDirectories: []string{
				"myTargetDir",
				filepath.FromSlash("myTargetDir/datasources"),
			},
		},
		{
			testDescription: "retrieve merged profile",
			giveFiles: []RetrieveFile{
				{"profiles/myprofile.json", existingProfileBody},
				{"profiles/myprofile.json", messyProfileBody},
			},
			wantFiles: []RetrieveFile{
				{filepath.FromSlash("profiles/myprofile.json"), mergedProfileBody},
			},
			wantDirectories: []string{
				"profiles",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			targetDir = tc.giveTargetDir
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
				if !types.StringSliceContainsKey(directoriesCreated, path) {
					directoriesCreated = append(directoriesCreated, path)
				}
				return nil
			}

			var mockExistingFileReader = func(path string) ([]byte, error) {
				return []byte(existingProfileBody), nil
			}

			bufData := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

			err = writeResultsToDisk([]*io.ReadCloser{&bufData}, mockFileMaker, mockDirectoryMaker, mockExistingFileReader)

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
