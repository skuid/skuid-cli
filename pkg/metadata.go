package pkg

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gookit/color"

	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
)

type NlxMetadata struct {
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
	Profiles           []string `json:"profiles"`
	Site               []string `json:"site"`
	Themes             []string `json:"themes"`
}

func GetFieldValueByNameError(target string) error {
	return errors.Error("GetFieldValueByName('%v') failed", target)
}

func (from NlxMetadata) GetFieldValueByName(target string) (names []string, err error) {
	mType := reflect.TypeOf(NlxMetadata{})

	var name string
	fieldCount := mType.NumField()
	for i := 0; i < fieldCount; i++ {
		field := mType.Field(i)
		if field.Tag.Get("json") == target {
			name = field.Name
			break
		}
	}

	if name == "" {
		err = GetFieldValueByNameError(target)
		return
	}

	value := reflect.ValueOf(from)
	field := value.FieldByName(name)
	if field.IsValid() {
		names = field.Interface().([]string)
		return
	}

	logging.Get().Tracef("Somehow able to find field name %v but not its value as []string in the metadata", name)
	err = GetFieldValueByNameError(target)

	return
}

// FilterItem returns true if the path meets the filter criteria, otherwise it returns false
func (from NlxMetadata) FilterItem(item string) (keep bool) {
	cleanRelativeFilePath := util.FromWindowsPath(item)
	directory := filepath.Dir(cleanRelativeFilePath)
	baseName := filepath.Base(cleanRelativeFilePath)

	// Find the lowest level folder
	dirSplit := strings.Split(directory, string(filepath.Separator))
	metadataType, subFolders := dirSplit[0], dirSplit[1:]
	filePathArray := append(subFolders, baseName)
	filePath := strings.Join(filePathArray, string(filepath.Separator))

	validMetadataNames, err := from.GetFieldValueByName(metadataType)
	if len(validMetadataNames) == 0 {
		logging.Get().Tracef("No valid names for this directory: %v", color.Gray.Sprint(item))
		return
	}

	if err != nil {
		logging.Get().Errorf("Metadata Filter Error: %v", err)
		return
	}

	if util.StringSliceContainsAnyKey(validMetadataNames, []string{
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
		if len(filePathParts) >= 2 && util.StringSliceContainsKey(validMetadataNames, filePathParts[0]) {
			logging.Get().Tracef("Keeping componentpack metadata file: %v", filePath)
			keep = true
			return
		}
	}

	if util.StringSliceContainsAnyKey(validMetadataNames, []string{
		// Check for our metadata with .xml stripped
		strings.TrimSuffix(filePath, ".xml"),
		// Check for our metadata with .skuid.json stripped
		strings.TrimSuffix(filePath, ".skuid.json"),
		// Check for theme inline css
		strings.TrimSuffix(filePath, ".inline.css"),
	}) {
		logging.Get().Tracef("Keeping metadata file: %v", filePath)
		keep = true
		return
	}

	return
}

// GetMetadataTypeDirNames returns the directory names for a type
func GetMetadataTypeDirNames() (types []string) {
	metadataType := reflect.TypeOf(NlxMetadata{})

	for i := 0; i < metadataType.NumField(); i++ {
		field := metadataType.Field(i)
		types = append(types, field.Tag.Get("json"))
	}

	return types
}
