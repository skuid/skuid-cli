package types

import (
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

type Metadata struct {
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
}

// GetMetadataTypeDirNames returns the directory names for a type
func GetMetadataTypeDirNames() []string {
	metadataType := reflect.TypeOf(Metadata{})
	fieldCount := metadataType.NumField()
	types := make([]string, fieldCount)
	for i := 0; i < fieldCount; i++ {
		field := metadataType.Field(i)
		types[i] = field.Tag.Get("json")
	}
	return types
}

// GetFieldNameForDirName returns the metadata field name for a given directory name
func GetFieldNameForDirName(dirName string) string {
	metadataType := reflect.TypeOf(Metadata{})
	fieldCount := metadataType.NumField()
	for i := 0; i < fieldCount; i++ {
		field := metadataType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == dirName {
			return field.Name
		}
	}
	return ""
}

// GetNamesForType returns the item names provided in the metadata for a particular type
func (m Metadata) GetNamesForType(metadataType string) []string {
	fieldName := GetFieldNameForDirName(metadataType)
	value := reflect.ValueOf(m)
	field := value.FieldByName(fieldName)
	if field.IsValid() {
		return field.Interface().([]string)
	}
	return nil
}

func fromWindowsPath(path string) string {
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

// FilterMetadataItem returns true if the path meets the filter criteria, otherwise it returns false
func (m Metadata) FilterMetadataItem(relativeFilePath string) bool {
	cleanRelativeFilePath := fromWindowsPath(relativeFilePath)
	directory := filepath.Dir(cleanRelativeFilePath)
	baseName := filepath.Base(cleanRelativeFilePath)

	// Find the lowest level folder
	dirSplit := strings.Split(directory, string(filepath.Separator))
	metadataType, subFolders := dirSplit[0], dirSplit[1:]
	filePathArray := append(subFolders, baseName)
	filePath := strings.Join(filePathArray, string(filepath.Separator))

	validMetadataNames := m.GetNamesForType(metadataType)
	if validMetadataNames == nil || len(validMetadataNames) == 0 {
		// If we don't have valid names for this directory, just skip this file
		return false
	}
	// Most common case --- check for our metadata with .json stripped
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(filePath, ".json")) {
		return true
	}
	// See if our filePath is in the valid metadata, if so, we're done
	if StringSliceContainsKey(validMetadataNames, filePath) {
		return true
	}
	// Check for children of a component pack
	if metadataType == "componentpacks" {
		filePathParts := strings.Split(filePath, string(filepath.Separator))
		if len(filePathParts) == 2 && StringSliceContainsKey(validMetadataNames, filePathParts[0]) {
			return true
		}
	}

	// Check for our metadata with .xml stripped
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(filePath, ".xml")) {
		return true
	}
	// Check for our metadata with .skuid.json stripped
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(filePath, ".skuid.json")) {
		return true
	}
	// Check for theme inline css
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(filePath, ".inline.css")) {
		return true
	}
	return false
}

// StringSliceContainsKey returns true if a string is contained in a slice
func StringSliceContainsKey(strings []string, key string) bool {
	for _, item := range strings {
		if item == key {
			return true
		}
	}
	return false
}
