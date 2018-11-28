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
}

type Metadata struct {
	Pages        []string `json:"pages"`
	Apps         []string `json:"apps"`
	DataServices []string `json:"dataservices"`
	DataSources  []string `json:"datasources"`
	Profiles     []string `json:"profiles"`
	Files        []string `json:"files"`
	Themes       []string `json:"themes"`
	DesignSystems []string `json:"designsystems"`
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

// FilterMetadataItem returns true if the path meets the filter criteria, otherwise it returns false
func (m Metadata) FilterMetadataItem(relativeFilePath string) bool {
	metadataType := filepath.Dir(relativeFilePath)
	baseName := filepath.Base(relativeFilePath)
	validMetadataNames := m.GetNamesForType(metadataType)
	if validMetadataNames == nil || len(validMetadataNames) == 0 {
		// If we don't have valid names for this directory, just skip this file
		return false
	}
	// See if our baseName is in the valid metadata, if so, we're done
	if StringSliceContainsKey(validMetadataNames, baseName) {
		return true
	}
	// Check for our metadata with .xml stripped
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(baseName, ".xml")) {
		return true
	}
	// Check for our metadata with .json stripped
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(baseName, ".json")) {
		return true
	}
	// Check for our metadata with .skuid.json stripped
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(baseName, ".skuid.json")) {
		return true
	}
	// Check for theme inline css
	if StringSliceContainsKey(validMetadataNames, strings.TrimSuffix(baseName, ".inline.css")) {
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
