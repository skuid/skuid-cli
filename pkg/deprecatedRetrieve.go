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

	"github.com/gookit/color"

	"github.com/skuid/tides/pkg/logging"
)

type Plan struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	Url      string   `json:"url"`
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

// for backwards compatibility
func (m *Metadata) UnmarshalJSON(data []byte) error {
	// unmarshal the old fields
	old := struct {
		Apps           []string `json:"apps"`
		AuthProviders  []string `json:"authproviders"`
		ComponentPacks []string `json:"componentpacks"`
		DataServices   []string `json:"dataservices"`
		DataSources    []string `json:"datasources"`
		DesignSystems  []string `json:"designsystems"`
		Variables      []string `json:"variables"`
		Files          []string `json:"files"`
		Pages          []string `json:"pages"`
		PermissionSets []string `json:"permissionsets"`
		Profiles       []string `json:"profiles"`
		Site           []string `json:"site"`
		Themes         []string `json:"themes"`
	}{}
	err := json.Unmarshal(data, &old)
	if err != nil {
		return err
	} else {
		m.Apps = old.Apps
		m.AuthProviders = old.AuthProviders
		m.ComponentPacks = old.ComponentPacks
		m.DataServices = old.DataServices
		m.DataSources = old.DataSources
		m.DesignSystems = old.DesignSystems
		m.Variables = old.Variables
		m.Files = old.Files
		m.Pages = old.Pages
		m.PermissionSets = old.PermissionSets
		m.Site = old.Site
		m.Themes = old.Themes
		m.SitePermissionSets = old.Profiles
	}

	// unmarshal the current fields and join them
	current := struct {
		Profiles []string `json:"sitepermissionsets"`
	}{}
	err = json.Unmarshal(data, &current)
	if err != nil {
		return err
	} else {
		// just append
		m.SitePermissionSets = append(m.SitePermissionSets, current.Profiles...)
	}

	return nil
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

func ExecuteRetrievePlan(api *NlxApi, plans map[string]Plan, noZip bool) (planResults []*io.ReadCloser, err error) {

	logging.VerboseSection("Executing Skuid NLX Retrieve Plan")

	for _, plan := range plans {
		var metadataBytes []byte
		if metadataBytes, err = json.Marshal(RetrieveRequest{
			Metadata: plan.Metadata,
			DoZip:    !noZip,
		}); err != nil {
			return
		}

		retrieveStart := time.Now()

		var url string
		var req func(string, string, io.Reader, string) (*io.ReadCloser, error)
		if plan.Host != "" {
			url = fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.Url)
			req = api.Connection.MakeAuthorizationBearerRequest
		} else {
			url = plan.Url
			req = api.Connection.MakeAccessTokenRequest
		}

		logging.VerboseF("Retrieval => %s (%s)\n", color.Yellow.Sprint(url), color.Cyan.Sprint(plan.Type))

		var planResult *io.ReadCloser
		if planResult, err = req(
			http.MethodPost,
			url,
			bytes.NewReader(metadataBytes),
			"application/json",
		); err != nil {
			return
		}

		planResults = append(planResults, planResult)

		logging.VerboseSuccess("Retrieve Plan Executed", retrieveStart)

	}
	return planResults, nil
}

func GetRetrievePlan(api *NlxApi, appName string) (results map[string]Plan, err error) {

	logging.VerboseSection("Skuid NLX Retrieval Plan")

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
	planResult, err := api.Connection.MakeAccessTokenRequest(
		http.MethodPost,
		"/metadata/retrieve/plan",
		postBody,
		"application/json",
	)

	if err != nil {
		return
	}

	logging.VerboseSuccess("Plan Retrieved", planStart)

	defer (*planResult).Close()

	err = json.NewDecoder(*planResult).Decode(&results)

	return
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
