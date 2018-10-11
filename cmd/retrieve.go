package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"archive/zip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/skuid/skuid/platform"
	"github.com/skuid/skuid/types"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve Skuid metadata from a Site into a local directory.",
	Long:  "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
	Run: func(cmd *cobra.Command, args []string) {

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			verbose,
		)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		plan, err := getRetrievePlan(api)
		if err != nil {
			fmt.Println("Error getting retrieve plan: ", err.Error())
			os.Exit(1)
		}

		results, err := executeRetrievePlan(api, plan)
		if err != nil {
			fmt.Println("Error executing retrieve plan: ", err.Error())
			os.Exit(1)
		}

		err = writeResultsToDisk(results)
		if err != nil {
			fmt.Println("Error writing results to disk: ", err.Error())
		}
	},
}

func getFriendlyURL(targetDir string) (string, error) {
	if targetDir == "" {
		friendlyResult, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}
		return friendlyResult, nil
	}
	return targetDir, nil

}

func writeResultsToDisk(results []*io.ReadCloser) error {
	// unzip the archive into the output directory
	targetDirFriendly, err := getFriendlyURL(targetDir)
	if err != nil {
		return err
	}
	fmt.Println("Writing results to " + targetDirFriendly + " ...")

	// Remove all of our metadata directories so we get a clean slate.
	// We may want to improve this later when we do partial retrieves so that
	// we don't clear out the whole directory every time we retrieve.
	for _, dirName := range types.GetMetadataTypeDirNames() {
		dirPath := filepath.Join(targetDir, dirName)
		if verbose {
			fmt.Println("Deleting Directory: " + dirPath)
		}
		os.RemoveAll(dirPath)
	}

	// Store a map of paths that we've already encountered. We'll use this
	// to determine if we need to modify a file or overwrite it.
	pathMap := map[string]bool{}

	for _, result := range results {

		tmpFileName, err := createTemporaryZipFile(result)
		if err != nil {
			return err
		}

		// unzip the contents of our temp zip file
		err = unzip(tmpFileName, targetDir, pathMap)

		// schedule cleanup of temp file
		defer os.Remove(tmpFileName)

		if err != nil {
			return err
		}
	}

	fmt.Printf("Results written to %s\n", targetDirFriendly)
	return nil
}

func createTemporaryZipFile(data *io.ReadCloser) (name string, err error) {
	tmpfile, err := ioutil.TempFile("", "skuid")
	if err != nil {
		return "", err
	}
	defer (*data).Close()
	// write to our new file
	if _, err := io.Copy(tmpfile, *data); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func unzip(sourceFileLocation, targetLocation string, pathMap map[string]bool) error {

	reader, err := zip.OpenReader(sourceFileLocation)

	if err != nil {
		return err
	}

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err := os.MkdirAll(targetLocation, 0755); err != nil {
			return err
		}
	}

	for _, file := range reader.File {
		readFileFromZipAndWriteToFilesystem(file, targetLocation, pathMap)
	}

	return nil
}

func readFileFromZipAndWriteToFilesystem(file *zip.File, targetLocation string, pathMap map[string]bool) error {
	path := filepath.Join(targetLocation, file.Name)

	// Check to see if we've already written to this file in this retrieve
	_, fileAlreadyWritten := pathMap[path]
	if !fileAlreadyWritten {
		pathMap[path] = true
	}

	// If this file name contains a /, make sure that we create the
	if pathParts := strings.Split(file.Name, "/"); len(pathParts) > 0 {
		// Remove the actual file name from the slice,
		// i.e. pages/MyAwesomePage.xml ---> pages
		pathParts = pathParts[:len(pathParts)-1]
		// and then make dirs for all paths up to that point, i.e. pages, apps
		intermediatePath := filepath.Join(targetLocation, strings.Join(pathParts[:], ","))
		//if the desired directory isn't there, create it
		if _, err := os.Stat(intermediatePath); err != nil {
			if verbose {
				fmt.Println("Creating intermediate directory: " + intermediatePath)
			}
			os.MkdirAll(intermediatePath, 0755)
		}
	}

	if file.FileInfo().IsDir() {
		os.MkdirAll(path, file.Mode())
		return nil
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	if fileAlreadyWritten {
		if verbose {
			fmt.Println("Augmenting existing file with more data: " + file.Name)
		}
		err = combineJSONFile(fileReader, path)
		if err != nil {
			return err
		}

	} else {
		if verbose {
			fmt.Println("Creating file: " + file.Name)
		}
		err = writeNewFile(fileReader, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeNewFile(fileReader io.ReadCloser, path string) error {
	targetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer targetFile.Close()
	if _, err := io.Copy(targetFile, fileReader); err != nil {
		return err
	}

	return nil
}

func combineJSONFile(fileReader io.ReadCloser, path string) error {
	existingBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	newBytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return err
	}

	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		return err
	}

	var indented bytes.Buffer
	err = json.Indent(&indented, combined, "", "\t")
	if err != nil {
		return err
	}

	return writeNewFile(ioutil.NopCloser(bytes.NewReader(indented.Bytes())), path)
}

func getRetrievePlan(api *platform.RestApi) (map[string]types.Plan, error) {
	// Get a retrieve plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/retrieve/plan",
		nil,
		"application/json",
	)

	if err != nil {
		return nil, err
	}

	defer (*planResult).Close()

	var plans map[string]types.Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func executeRetrievePlan(api *platform.RestApi, plans map[string]types.Plan) ([]*io.ReadCloser, error) {
	planResults := []*io.ReadCloser{}
	for _, plan := range plans {
		metadataBytes, err := json.Marshal(plan.Metadata)
		if err != nil {
			return nil, err
		}
		if plan.Host == "" {
			planResult, err := api.Connection.MakeRequest(
				http.MethodPost,
				plan.URL,
				bytes.NewReader(metadataBytes),
				"application/json",
			)
			if err != nil {
				return nil, err
			}
			planResults = append(planResults, planResult)
		} else {
			url := fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.URL)
			planResult, err := api.Connection.MakeJWTRequest(
				http.MethodPost,
				url,
				bytes.NewReader(metadataBytes),
				"application/json",
			)
			if err != nil {
				return nil, err
			}
			planResults = append(planResults, planResult)
		}
	}
	return planResults, nil
}

func init() {
	RootCmd.AddCommand(retrieveCmd)
}
