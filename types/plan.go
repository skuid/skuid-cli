package types

import "reflect"

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
