package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
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
		types[i] = field.Tag.Get("json")
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
		VerboseError("FilterMetadataItem error", err)
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
