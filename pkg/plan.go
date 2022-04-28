package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/skuid/tides/pkg/logging"
)

type Plan struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	URL      string   `json:"url"`
	Type     string   `json:"type"`
	Metadata Metadata `json:"metadata"`
	Warnings []string `json:"warnings"`
}

type RetrieveRequest struct {
	DoZip    bool     `json:"zip"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Apps               []string `json:"apps"`
	AuthProviders      []string `json:"authproviders"`
	ComponentPacks     []string `json:"componentpacks"`
	DataServices       []string `json:"dataservices"`
	DataSources        []string `json:"datasources"`
	DesignSystems      []string `json:"designsystems"`
	Variables          []string `json:"variables"`
	Files              []string `json:"files"`
	Pages              []string `json:"pages"`
	PermissionSets     []string `json:"permissionsets"`
	SitePermissionSets []string `json:"sitepermissionsets"`
	Site               []string `json:"site"`
	Themes             []string `json:"themes"`
}

type RetrieveFilter struct {
	AppName string `json:"appName"`
}

type DeployFilter struct {
	AppName string `json:"appName"`
	Plan    []byte `json:"plan"`
}

// GetMetadataTypeDirNames returns the directory names for a type
func GetMetadataTypeDirNames() (types []string) {
	metadataType := reflect.TypeOf(Metadata{})

	for i := 0; i < metadataType.NumField(); i++ {
		field := metadataType.Field(i)
		types = append(types, field.Tag.Get("json"))
	}

	return types
}

// GetFieldNameForDirName returns the metadata field name for a given directory name
func GetFieldNameForDirName(dirName string) (fieldName string, err error) {
	metadataType := reflect.TypeOf(Metadata{})

	for i := 0; i < metadataType.NumField(); i++ {
		field := metadataType.Field(i)
		if field.Tag.Get("json") == dirName {
			fieldName = field.Name
			return
		}
	}

	err = fmt.Errorf("Field for dir '%v' not found", dirName)

	return
}

// GetNamesForType returns the item names provided in the metadata for a particular type
func (m Metadata) GetNamesForType(metadataType string) (names []string, err error) {
	fieldName, err := GetFieldNameForDirName(metadataType)
	if err != nil {
		return
	}

	value := reflect.ValueOf(m)
	field := value.FieldByName(fieldName)
	if field.IsValid() {
		names = field.Interface().([]string)
		return
	}

	err = fmt.Errorf("Names for type %v not found.", metadataType)

	return
}

func FromWindowsPath(path string) string {
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

// FilterMetadataItem returns true if the path meets the filter criteria, otherwise it returns false
func (m Metadata) FilterMetadataItem(relativeFilePath string) (keep bool) {
	cleanRelativeFilePath := FromWindowsPath(relativeFilePath)
	directory := filepath.Dir(cleanRelativeFilePath)
	baseName := filepath.Base(cleanRelativeFilePath)

	// Find the lowest level folder
	dirSplit := strings.Split(directory, string(filepath.Separator))
	metadataType, subFolders := dirSplit[0], dirSplit[1:]
	filePathArray := append(subFolders, baseName)
	filePath := strings.Join(filePathArray, string(filepath.Separator))

	validMetadataNames, err := m.GetNamesForType(metadataType)
	if validMetadataNames == nil || len(validMetadataNames) == 0 {
		// If we don't have valid names for this directory, just skip this file
		return
	}

	if err != nil {
		logging.VerboseError("FilterMetadataItem error", err)
		return
	}

	if StringSliceContainsAnyKey(validMetadataNames, []string{
		// Most common case --- check for our metadata with .json stripped
		strings.TrimSuffix(filePath, ".json"),
		// See if our filePath is in the valid metadata, if so, we're done
		filePath,
	}) {
		keep = true
		return
	}

	// Check for children of a component pack
	if metadataType == "componentpacks" {
		filePathParts := strings.Split(filePath, string(filepath.Separator))
		if len(filePathParts) == 2 && StringSliceContainsKey(validMetadataNames, filePathParts[0]) {
			keep = true
			return
		}
	}

	if StringSliceContainsAnyKey(validMetadataNames, []string{
		// Check for our metadata with .xml stripped
		strings.TrimSuffix(filePath, ".xml"),
		// Check for our metadata with .skuid.json stripped
		strings.TrimSuffix(filePath, ".skuid.json"),
		// Check for theme inline css
		strings.TrimSuffix(filePath, ".inline.css"),
	}) {
		keep = true
		return
	}

	return
}

func GetRetrievePlan(api *NlxApi, appName string) (map[string]Plan, error) {

	logging.VerboseSection("Getting Retrieve Plan")

	var postBody io.Reader
	if appName != "" {
		retFilter, err := json.Marshal(RetrieveFilter{
			AppName: appName,
		})
		if err != nil {
			return nil, err
		}
		postBody = bytes.NewReader(retFilter)
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

	logging.VerboseSuccess("Success Getting Retrieve Plan", planStart)

	defer (*planResult).Close()

	var plans map[string]Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func ExecuteRetrievePlan(api *NlxApi, plans map[string]Plan, noZip bool) (planResults []*io.ReadCloser, err error) {

	logging.VerboseSection("Executing Retrieve Plan")

	for _, plan := range plans {
		metadataBytes, err := json.Marshal(RetrieveRequest{
			Metadata: plan.Metadata,
			DoZip:    !noZip,
		})
		if err != nil {
			return nil, err
		}
		retrieveStart := time.Now()
		if plan.Host == "" {

			logging.VerboseF("Making Retrieve Request: URL: [%s] Type: [%s]\n", plan.URL, plan.Type)

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

			logging.VerboseF("Making Retrieve Request: URL: [%s] Type: [%s]\n", url, plan.Type)

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

		logging.VerboseSuccess("Success Retrieving from Source", retrieveStart)

	}
	return planResults, nil
}

func DeployModifiedFiles(api *NlxApi, targetDir, modifiedFile string) (err error) {

	// Create a buffer to write our archive to.
	bufPlan := new(bytes.Buffer)
	err = ArchivePartial(targetDir, bufPlan, modifiedFile)
	if err != nil {
		err = fmt.Errorf("Error creating deployment ZIP archive: %v", err)
		return
	}

	logging.VerboseLn("Getting deploy plan...")

	plan, err := api.GetDeployPlan(bufPlan, "application/zip")
	if err != nil {
		err = fmt.Errorf("Error getting deploy plan: %v", err)
		return
	}

	logging.VerboseLn("Retrieved deploy plan. Deploying...")

	_, err = api.ExecuteDeployPlan(plan, targetDir)
	if err != nil {
		err = fmt.Errorf("Error executing deploy plan: %v", err)
		return
	}

	successMessage := "Successfully deployed metadata to Skuid Site: " + modifiedFile
	logging.Println(successMessage)

	return
}
