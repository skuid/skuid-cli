package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsoniter "github.com/skuid/json-iterator-go" // jsoniter. Fork of github.com/json-iterator/go
	jsonpatch "github.com/skuid/json-patch"
	"github.com/skuid/skuid-cli/platform"
	"github.com/skuid/skuid-cli/text"
	"github.com/skuid/skuid-cli/types"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve Skuid metadata from a Site into a local directory.",
	Long:  "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(text.RunCommand("Retrieve Metadata"))

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			metadataServiceProxy,
			dataServiceProxy,
			verbose,
		)

		retrieveStart := time.Now()

		if err != nil {
			fmt.Println(text.PrettyError("Error logging in to Skuid site", err))
			os.Exit(1)
		}

		plan, err := getRetrievePlan(api, appName, *pages)
		if err != nil {
			fmt.Println(text.PrettyError("Error getting retrieve plan", err))
			os.Exit(1)
		}

		results, err := executeRetrievePlan(api, plan)
		if err != nil {
			fmt.Println(text.PrettyError("Error executing retrieve plan", err))
			os.Exit(1)
		}

		err = writeResultsToDisk(results, writeNewFile, createDirectory, readExistingFile)
		if err != nil {
			fmt.Println(text.PrettyError("Error writing results to disk", err))
			os.Exit(1)
		}

		successMessage := "Successfully retrieved metadata from Skuid Site"
		if verbose {
			fmt.Println(text.SuccessWithTime(successMessage, retrieveStart))
		} else {
			fmt.Println(successMessage + ".")
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

func writeResultsToDisk(results []*io.ReadCloser, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {

	// unzip the archive into the output directory
	targetDirFriendly, err := getFriendlyURL(targetDir)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println(text.VerboseSection("Writing results to " + targetDirFriendly))
	}

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

		tmpFileName, err := createTemporaryFile(result)
		if err != nil {
			return err
		}
		// schedule cleanup of temp file
		defer os.Remove(tmpFileName)

		if nozip {
			err = moveTempFile(tmpFileName, targetDir, pathMap, fileCreator, directoryCreator, existingFileReader)
			if err != nil {
				return err
			}
			continue
		}

		// unzip the contents of our temp zip file
		err = unzip(tmpFileName, targetDir, pathMap, fileCreator, directoryCreator, existingFileReader)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Results written to %s\n", targetDirFriendly)

	return nil
}

func createTemporaryFile(data *io.ReadCloser) (name string, err error) {
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

func moveTempFile(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {
	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err := directoryCreator(targetLocation, 0755); err != nil {
			return err
		}
	}
	fi, err := os.Open(sourceFileLocation)
	if err != nil {
		return err
	}
	defer fi.Close()
	fstat, err := fi.Stat()
	if err != nil {
		return err
	}
	fileReader := ioutil.NopCloser(fi)

	path := filepath.Join(targetLocation, filepath.FromSlash(fi.Name()))
	_, fileAlreadyWritten := pathMap[path]
	if !fileAlreadyWritten {
		pathMap[path] = true
	}
	if fstat.IsDir() {
		return directoryCreator(path, fstat.Mode())
	}
	if fileAlreadyWritten {
		if verbose {
			fmt.Println("Augmenting existing file with more data: " + fi.Name())
		}
		fileReader, err = combineJSONFile(fileReader, existingFileReader, path)
		if err != nil {
			return err
		}
	}
	if verbose {
		fmt.Println("Creating file: " + fi.Name())
	}
	err = fileCreator(fileReader, path)
	if err != nil {
		return err
	}

	return nil
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func unzip(sourceFileLocation, targetLocation string, pathMap map[string]bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {

	reader, err := zip.OpenReader(sourceFileLocation)
	if err != nil {
		return err
	}

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err := directoryCreator(targetLocation, 0755); err != nil {
			return err
		}
	}

	for _, file := range reader.File {
		path := filepath.Join(targetLocation, filepath.FromSlash(file.Name))
		// Check to see if we've already written to this file in this retrieve
		_, fileAlreadyWritten := pathMap[path]
		if !fileAlreadyWritten {
			pathMap[path] = true
		}
		err := readFileFromZipAndWriteToFilesystem(file, path, fileAlreadyWritten, fileCreator, directoryCreator, existingFileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func readFileFromZipAndWriteToFilesystem(file *zip.File, fullPath string, fileAlreadyWritten bool, fileCreator FileCreator, directoryCreator DirectoryCreator, existingFileReader FileReader) error {

	// If this file name contains a /, make sure that we create the directory it belongs in
	if pathParts := strings.Split(fullPath, string(filepath.Separator)); len(pathParts) > 0 {
		// Remove the actual file name from the slice,
		// i.e. pages/MyAwesomePage.xml ---> pages
		pathParts = pathParts[:len(pathParts)-1]
		// and then make dirs for all paths up to that point, i.e. pages, apps
		intermediatePath := filepath.Join(strings.Join(pathParts[:], string(filepath.Separator)))
		if intermediatePath != "" {
			err := directoryCreator(intermediatePath, 0755)
			if err != nil {
				return err
			}
		} else {
			// If we don't have an intermediate path, skip out.
			// Currently Skuid CLI does not create any files in the base directory
			return nil
		}
	}

	if file.FileInfo().IsDir() {
		return directoryCreator(fullPath, file.Mode())
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
		fileReader, err = combineJSONFile(fileReader, existingFileReader, fullPath)
		if err != nil {
			return err
		}

	}
	if verbose {
		fmt.Println("Creating file: " + file.Name)
	}
	err = fileCreator(fileReader, fullPath)
	if err != nil {
		return err
	}

	return nil
}

func createDirectory(path string, fileMode os.FileMode) error {
	if _, err := os.Stat(path); err != nil {
		if verbose {
			fmt.Println("Creating intermediate directory: " + path)
		}
		return os.MkdirAll(path, fileMode)
	}
	return nil
}

type FileCreator func(fileReader io.ReadCloser, path string) error
type DirectoryCreator func(path string, fileMode os.FileMode) error
type FileReader func(path string) ([]byte, error)

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

func readExistingFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// Define custom json object key sorter for use in combineJSONFile() below.
// The intent here is to always have deterministically sorted maps from merged JSON objects.
type nameFirstKeySorter struct{}

func (sorter *nameFirstKeySorter) Sort(keyA string, keyB string) bool {
	if keyA == "name" {
		return true
	} else if keyB == "name" {
		return false
	} else {
		return keyA < keyB
	}
}

type nameFirstKeyExtension struct {
	jsoniter.DummyExtension
	sorter jsoniter.MapKeySorter
}

func (extension *nameFirstKeyExtension) CreateMapKeySorter() jsoniter.MapKeySorter {
	return extension.sorter
}

func combineJSONFile(newFileReader io.ReadCloser, existingFileReader FileReader, path string) (io.ReadCloser, error) {
	existingBytes, err := existingFileReader(path)
	if err != nil {
		return nil, err
	}
	newBytes, err := ioutil.ReadAll(newFileReader)
	if err != nil {
		return nil, err
	}

	// Configure jsoniter to sort map keys alpha, unless key is "name", which goes first
	jsonConfig := jsoniter.Config{
		SortMapKeys:           true,
		DisallowUnknownFields: false,
	}.Froze()
	jsonConfig.RegisterExtension(&nameFirstKeyExtension{
		sorter: &nameFirstKeySorter{},
	})
	// Configure jsonpatch to use jsoniter with custom sorter for merging json
	jsonpatch.SetAPI(jsonConfig)

	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		return nil, err
	}

	var indented bytes.Buffer
	err = json.Indent(&indented, combined, "", "\t")
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewReader(indented.Bytes())), nil
}

func getRetrievePlan(api *platform.RestApi, appName string, pages []string) (map[string]types.Plan, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Getting Retrieve Plan"))
	}
	var postBody io.Reader
	var retFilter types.RetrieveFilter
	if len(pages) > 0 {
		retFilter.Pages = pages
	}
	if appName != "" {
		retFilter.AppName = appName
	}
	if retFilter.Pages != nil || retFilter.AppName != "" {
		if retFilterBytes, err := json.Marshal(retFilter); err != nil {
			return nil, err
		} else {
			postBody = bytes.NewReader(retFilterBytes)
		}

	}

	planStart := time.Now()
	// Get a retrieve plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/retrieve/plan",
		postBody,
		"application/json",
	)

	if err != nil {
		return nil, err
	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success Getting Retrieve Plan", planStart))
	}

	defer (*planResult).Close()

	var plans map[string]types.Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func executeRetrievePlan(api *platform.RestApi, plans map[string]types.Plan) ([]*io.ReadCloser, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Executing Retrieve Plan"))
	}
	planResults := []*io.ReadCloser{}
	for _, plan := range plans {
		metadataBytes, err := json.Marshal(types.RetrieveRequest{
			Metadata: plan.Metadata,
			DoZip:    !nozip,
		})
		if err != nil {
			return nil, err
		}
		retrieveStart := time.Now()
		if plan.Host == "" {
			if verbose {
				fmt.Println(fmt.Sprintf("Making Retrieve Request: URL: [%s] Type: [%s]", plan.URL, plan.Type))
			}
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
			if verbose {
				fmt.Println(fmt.Sprintf("Making Retrieve Request: URL: [%s] Type: [%s]", url, plan.Type))
			}
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

		if verbose {
			fmt.Println(text.SuccessWithTime("Success Retrieving from Source", retrieveStart))
		}
	}
	return planResults, nil
}

func init() {
	RootCmd.AddCommand(retrieveCmd)
}
