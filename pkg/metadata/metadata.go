package metadata

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/orsinium-labs/enum"

	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var (
	// Skuid Review Required - These rules are based on what the Web UI appears to allow for the various metadata types.  They should be reviewed for accuracy
	// and adjusted as needed.
	// Allow only use spaces, letters, numbers, underscores, and dashes
	entityNameValidator = regexp.MustCompile(`^[a-zA-Z0-9_\- ]+$`)
	// Allow only use spaces, letters, numbers, underscores, dashes, parenthesis and periods
	fileEntityNameValidator = regexp.MustCompile(`^[a-zA-Z0-9_\-\(\)\. ]+$`)

	metadataTypeBuilder            = enum.NewBuilder[MetadataTypeValue, MetadataType]()
	MetadataTypeApps               = metadataTypeBuilder.Add(MetadataType{"Apps"})
	MetadataTypeAuthProviders      = metadataTypeBuilder.Add(MetadataType{"AuthProviders"})
	MetadataTypeComponentPacks     = metadataTypeBuilder.Add(MetadataType{"ComponentPacks"})
	MetadataTypeDataServices       = metadataTypeBuilder.Add(MetadataType{"DataServices"})
	MetadataTypeDataSources        = metadataTypeBuilder.Add(MetadataType{"DataSources"})
	MetadataTypeDesignSystems      = metadataTypeBuilder.Add(MetadataType{"DesignSystems"})
	MetadataTypeVariables          = metadataTypeBuilder.Add(MetadataType{"Variables"})
	MetadataTypeFiles              = metadataTypeBuilder.Add(MetadataType{"Files"})
	MetadataTypePages              = metadataTypeBuilder.Add(MetadataType{"Pages"})
	MetadataTypePermissionSets     = metadataTypeBuilder.Add(MetadataType{"PermissionSets"})
	MetadataTypeSitePermissionSets = metadataTypeBuilder.Add(MetadataType{"SitePermissionSets"})
	MetadataTypeSessionVariables   = metadataTypeBuilder.Add(MetadataType{"SessionVariables"})
	MetadataTypeSite               = metadataTypeBuilder.Add(MetadataType{"Site"})
	MetadataTypeThemes             = metadataTypeBuilder.Add(MetadataType{"Themes"})
	MetadataTypes                  = metadataTypeBuilder.Enum()
)

const (
	MetadataSubTypeNone MetadataSubType = iota + 1
	MetadataSubTypeSiteLogo
	MetadataSubTypeSiteFavicon

	EntityNameSite                     string = "site"
	metadataTypeEntityPathNotSupported string = "metadata type %q does not support the entity path: %q"
)

type MetadataTypeValue string
type MetadataSubType int

func (s MetadataTypeValue) Equal(other MetadataTypeValue) bool {
	return strings.EqualFold(string(s), string(other))
}

type MetadataType enum.Member[MetadataTypeValue]

func (m MetadataType) Name() string {
	return string(m.Value)
}

func (m MetadataType) DirName() string {
	return strings.ToLower(m.Name())
}

type entityPathDetails struct {
	Type         MetadataType
	SubType      MetadataSubType
	Name         string // The name of the entity/file (e.g., my_page.xml, my_page)
	Path         string // Path to the entity/file (e.g., pages/my_page.xml, pages/my_page)
	PathRelative string // Path to the entity/file relative to the metadata directory (e.g., my_page.xml, my_page, logo/somelogo.png)
}

type MetadataEntity struct {
	Type         MetadataType    // The of the entity (e.g., Pages)
	SubType      MetadataSubType // The subtype of the entity type (e.g., MetadataSubTypeNone for pages/my_page, MetadataSubTypeSiteLogo for site/logo/my_logo)
	Name         string          // Name of the entity (e.g., my_page)
	Path         string          // Path to the entity (e.g., pages/my_page)
	PathRelative string          // Path to the entity relative to the metadata directory (e.g., my_page, logo/my_logo)
}

type MetadataEntityFile struct {
	Entity                 MetadataEntity // Entity the file is associated to
	Name                   string         // Name of the file (e.g., my_page.xml)
	Path                   string         // Path to the file (e.g., pages/my_page.xml)
	PathRelative           string         // Path to the file relative to the metadata directory (e.g., my_page.xml, logo/my_logo.png)
	IsEntityDefinitionFile bool           // true if the file contains the entity definition (e.g., pages/my_page.json, apps/my_app.json, files/my_file.txt.skuid.json), false otherwise
}

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
	SessionVariables   []string `json:"sessionvariables"`
	Site               []string `json:"site"`
	Themes             []string `json:"themes"`
}

func (from NlxMetadata) GetFieldValue(target MetadataType) []string {
	name := target.Name()
	value := reflect.ValueOf(from)
	field := value.FieldByName(name)
	// should never occur in production
	errors.MustConditionf(field.IsValid(), "unable to locate metadata field for metadata type %q", name)
	return field.Interface().([]string)
}

// FilterItem returns true if the path meets the filter criteria, otherwise it returns false
func (from NlxMetadata) FilterItem(item MetadataEntityFile) bool {
	validMetadataNames := from.GetFieldValue(item.Entity.Type)
	if slices.Contains(validMetadataNames, item.Entity.PathRelative) {
		logging.Get().Tracef("keeping metadata file: %v", item.Path)
		return true
	}

	// We can get here for the following reasons:
	// 1. The entity was filtered via NlxMetadata filter (e.g., app name) - this scenario would be expected
	// 2. We detected a valid metadata file locally, included it in the GetDeployPayload request but the server, for some reason, did not include
	//    it in the deploy plan - this scenario is NOT expected.
	//
	// Skuid Review Required - What are the valid scenarios where #2 would occur?  If we have a valid metadata file locally
	// why would the server ever not accept it and more specifically, if we request GetDeployPayload and include a file(s) that
	// the server isn't going to accept, why doesn't GetDeployPayload fail with an HTTP 400?  GetDeployPayload shouldn't accept
	// requests to obtain a payload that contain unexpected files.  Three examples of this I've found thus far
	// are https://github.com/skuid/skuid-cli/issues/184 && https://github.com/skuid/skuid-cli/issues/163 && https://github.com/skuid/skuid-cli/issues/158
	//
	// For scenario 1, this should be Tracef, for Scenario 2 this should be Warnf at minimum but we currently have no way to detect which scenario
	// it is at this point in the code and really there is no easy way to do this client side anywhere since we would have to parse things like "app"
	// info to identify it's permission sets, pages, etc. in the scenario where we are filtering by app.
	// TODO: The server API should be updated to fail if an invalid payload (e.g., a file it doesn't like) is submitted in GetDeployPlan so that we only
	// hit this code due to a filter having been applied
	logging.Get().Tracef("skipping metadata file, possibly due to a filter being applied: %q", item.Path)
	return false
}

// Accepts a relative entity path (e.g, pages/my_page, ) and returns a *MetadataEntity or error if
// unable to parse the path
func NewMetadataEntity(entityPath string) (*MetadataEntity, error) {
	details, err := parseEntityPath(entityPath)
	if err != nil {
		return nil, err
	}
	valid := validateEntityName(details)
	if !valid {
		return nil, fmt.Errorf(metadataTypeEntityPathNotSupported, details.Type.Name(), entityPath)
	}
	entity := MetadataEntity{
		Type:         details.Type,
		SubType:      details.SubType,
		Name:         details.Name,
		Path:         details.Path,
		PathRelative: details.PathRelative,
	}
	return &entity, nil
}

// Accepts a relative entity file path (e.g, pages/my_page.xml, pages/my_page.json) and returns a
// *MetadataEntityFile or error if unable to parse the path
func NewMetadataEntityFile(entityFilePath string) (*MetadataEntityFile, error) {
	details, err := parseEntityPath(entityFilePath)
	if err != nil {
		return nil, err
	}
	entityName, entityRelativePath, isEntityDefinitionFile, valid := entityNameFromFilePath(details)
	if !valid {
		return nil, fmt.Errorf(metadataTypeEntityPathNotSupported, details.Type.Name(), entityFilePath)
	}

	// not really necessary but a santity check for future proofing against code adjustments that don't have full test coverage
	// recognizing that just because it has a .json extension doesn't mean its a json file :(
	// should never happen in production
	errors.MustConditionf(!isEntityDefinitionFile || path.Ext(details.Path) == ".json", "entity definition file does not have .json extension: %q", details.Path)

	item := &MetadataEntityFile{
		Entity: MetadataEntity{
			Type:         details.Type,
			SubType:      details.SubType,
			Name:         entityName,
			Path:         path.Join(details.Type.DirName(), entityRelativePath),
			PathRelative: entityRelativePath,
		},
		Name:                   details.Name,
		Path:                   details.Path,
		PathRelative:           details.PathRelative,
		IsEntityDefinitionFile: isEntityDefinitionFile,
	}
	return item, nil
}

// returns true if path is in the form <metadatadirectory>/<any>/** and <metadatadirectory> is a valid metadata type, false otherwise
// does not validate the that <any>/** is valid, only that the metadata type is valid.  For example, pages/my_page.txt
// will return true since pages is a known metadata type even though my_page.txt is not a valid pages entity file.
func IsMetadataTypePath(path string) bool {
	if _, _, _, err := parseMetadataType(path); err != nil {
		return false
	} else {
		return true
	}
}

// returns true if two slices are equal ignoring the order of the elements, false otherwise. If there are duplicate elements,
// the number of appearances of each of them in both lists should match.
func EntitiesMatch[S ~[]MetadataEntity](s1 S, s2 S) bool {
	// don't mutate the originals
	a := slices.Clone(s1)
	b := slices.Clone(s2)

	// sort and then use built-in compare to keep things simpler rather than manually
	// checking lengths and iterating both outer an inner since that would require
	// keeping track of visited items to detect duplicates, etc.  For the situations
	// where we need to match, the lists should be relatively short, however if perf
	// becomes an issue, analysis can be performed against both approaches using
	// identified use cases to determine which approach yields better performance
	slices.SortFunc(a, sortEntities)
	slices.SortFunc(b, sortEntities)
	return slices.Equal(a, b)
}

// returns the unique metadata entities found in the specified slice
func UniqueEntities[S ~[]MetadataEntity](s S) []MetadataEntity {
	// don't mutate the originals
	a := slices.Clone(s)

	// multiple ways this can be achieved - the below assumes slices are small
	// and therefore does not optimize for performance.  If perf becomes an issue,
	// use cases should be evaluated against available methods for identifying unique
	// values and appropriate approach implemented
	slices.SortFunc(a, sortEntities)
	return slices.Compact(a)
}

// parses either a entity name path (e.g., pages/my_page) or an entity file path (e.g., pages/my_page.xml)
// all paths returned are normalized to entity path format (`/` separator)
func parseEntityPath(originalEntityPath string) (*entityPathDetails, error) {
	normalizedEntityPath, metadataType, subFolders, err := parseMetadataType(originalEntityPath)
	if err != nil {
		return nil, err
	}
	baseName := path.Base(normalizedEntityPath)
	relativePathSegments := append(subFolders, baseName)
	relativeEntityPath := path.Join(relativePathSegments...)
	subType, ok := parseMetadataSubType(*metadataType, relativeEntityPath)
	if !ok {
		return nil, fmt.Errorf(metadataTypeEntityPathNotSupported, (*metadataType).Name(), originalEntityPath)
	}

	details := &entityPathDetails{
		Type:         *metadataType,
		SubType:      subType,
		Name:         baseName,
		Path:         normalizedEntityPath,
		PathRelative: relativeEntityPath,
	}

	return details, nil
}

// parses a path in the form <metadatadirectory>/<any>/** returning the following with all paths
// normalized to entity path format (`/` separator)
// string - given path normalized to entity path format or empty string if path invalid
// *MetadataType - MetadataType detected or nil if path invalid
// []string - any remaining path folder segments relative to <metadatadirectory> or nil if path invalid
// error - nil or error containing issue with path if path invalid
func parseMetadataType(originalEntityPath string) (string, *MetadataType, []string, error) {
	normalizedEntityPath := filepath.ToSlash(filepath.Clean(originalEntityPath))
	directory := path.Dir(normalizedEntityPath)
	if path.IsAbs(directory) || directory == "" || directory == "." {
		return "", nil, nil, fmt.Errorf("directory name matching a valid metadata type name must exist in entity path: %q", originalEntityPath)
	}

	// Find the top/root level folder
	dirSplit := strings.Split(directory, "/")
	metadataName, subFolders := dirSplit[0], dirSplit[1:]
	metadataType := enum.Parse(MetadataTypes, MetadataTypeValue(metadataName))
	if metadataType == nil {
		return "", nil, nil, fmt.Errorf("invalid metadata type name %q for entity path: %q", metadataName, originalEntityPath)
	}

	return normalizedEntityPath, metadataType, subFolders, nil
}

// parses a path in the form <any>/** returning the following:
// MetadataSubType - the subtype for the specified metadata type
// bool - true if the path is valid, false otherwise
//
// Except for ComponentPacks, validation will be performed to ensure that the path is located in a valid directory
// relative to the metadata directory itself for the metadata type specified.  Validation is not performed on the entity name
// only the relative directory location of the entity. For example:
//  1. For the metadata type pages, a relativeEntityPath of "my_page" will return MetadataSubTypeNone, true since pages does
//     not have any subtypes and "my_page" is located in the root of pages.
//  2. For the metadata type pages, a relativeEntityPath of "foobar/my_page" will return MetadataSubTypeNone, false since
//     pages does not allow subdirectories
//  3. For the metadata type site, a relativeEntityPath of "favicon/my_icon.ico" will return MetadataSubTypeFavicon, true
//  4. For the metadata type site, a relativeEntityPath of "foobar/my_icon.ico" will return MetadataSubTypeFavicon, false since
//     foobar is not a valid subdirectory for site
//  5. For the metadata type site, a relativeEntityPath of "foobar.json" will return MetadataSubTypeNone, true since site
//     allows files in the root of site directory
//
// For Component Packs, unlike all other types, the entity name comes from the directory name as opposed to the filename
// so MetadataSubTypeNone, true is returned regardless of relativeEntityPath structure
func parseMetadataSubType(mdt MetadataType, relativeEntityPath string) (MetadataSubType, bool) {
	subdir := path.Dir(relativeEntityPath)
	switch mdt {
	case MetadataTypeSite:
		if subdir == "favicon" {
			return MetadataSubTypeSiteFavicon, true
		} else if subdir == "logo" {
			return MetadataSubTypeSiteLogo, true
		} else {
			return MetadataSubTypeNone, subdir == "."
		}
	case MetadataTypeComponentPacks:
		// Skuid Review Required - the entity name is the name of the subdir which can be any value
		// so when parsing the entity name, the directory will be "." but when parsing an entity file
		// the dir will be <any> where <any> can be 1 to N layers deep.  Given this, we need to allow
		// for "." and <any> for directory and the corresponding Entity & Entity File parsing validation
		// will ensure the directory is valid.  Not sure there is any other way to enforce this here?
		return MetadataSubTypeNone, true
	default:
		return MetadataSubTypeNone, subdir == "."
	}
}

// Skuid Review Required - This code uses concepts in the code at https://github.com/skuid/skuid-cli/blob/master/pkg/metadata.go#L68
// along with applying knowledge from observations of retrieving site metadata obtained via retrieve.  For some metadata types, I have
// no way to test behavior on a real site (e.g., componentpacks are only supported on v1 but unclear how to create a v1 page - see
// https://github.com/skuid/skuid-cli/issues/196) so for those types, while I have included unit tests, I can't guarantee that the
// behavior tested in accurate.  In addition to componentpacks, the old code references *.inline.css as "theme inline css" but I'm
// unable to determine where/how those files would exist. In short, this code should be throughly reviewed for accuracy and
// completeness across all metadata types.
func validateEntityName(details *entityPathDetails) bool {
	entityName := details.Name
	ext := path.Ext(entityName)
	entityNameWithoutExtension := strings.TrimSuffix(entityName, ext)

	switch details.Type {
	case MetadataTypeFiles:
		return details.SubType == MetadataSubTypeNone && fileEntityNameValidator.MatchString(entityName)
	case MetadataTypeSite:
		// Skuid Review Required - For favicon & logo, the Web UI indicates that only ico & png/jpg/gif are supported respectively.
		// How does skuid evaluate validty - by extension and if so, which are valid?  by mime-type and if so, which are valid?
		// TODO: Modify below and update tests based on answers to these questions
		if details.SubType == MetadataSubTypeNone {
			return entityName == EntityNameSite
		} else if details.SubType == MetadataSubTypeSiteFavicon {
			return ext == ".ico" && fileEntityNameValidator.MatchString(entityNameWithoutExtension)
		} else if details.SubType == MetadataSubTypeSiteLogo {
			return (ext == ".png" || ext == ".jpg" || ext == ".gif") && fileEntityNameValidator.MatchString(entityNameWithoutExtension)
		} else {
			return false
		}
	case MetadataTypeComponentPacks:
		// Skuid Review Required - Component pack entity names come from the directory name as opposed to the filename so parseMetadataSubType
		// must allow both and therefore we must validate the directory name to ensure its in the root of the Metadata Type
		// directory.  Not sure there is a better way to do this?
		directory := path.Dir(details.PathRelative)
		return details.SubType == MetadataSubTypeNone && directory == "." && ext == "" && entityNameValidator.MatchString(entityNameWithoutExtension)

	default:
		//return directory == "." && ext == "" && entityNameValidator.MatchString(entityNameWithoutExtension)
		return details.SubType == MetadataSubTypeNone && ext == "" && entityNameValidator.MatchString(entityNameWithoutExtension)
	}
}

// Skuid Review Required - This code is based on the code at https://github.com/skuid/skuid-cli/blob/master/pkg/metadata.go#L68.
// Even though the old code knew the metadata type, it took a "if it looks like a duck, quacks like a duck or even remotely
// resembles a duck, assume its a duck" approach to satisfying the filter which could lead to inaccuracies.  For example, if a file
// pages/invalid.skuid.json exists, it would have passed the filter check even though its not an expected file for pages. Since the
// metadata type is known, we know what files to expect for that type and how to resolve the entity name.  For some metadata types,
// I have no way to test behavior on a real site (e.g., componentpacks are only supported on v1 but unclear how to create a v1 page -
// see https://github.com/skuid/skuid-cli/issues/196) so for those types, while I have included unit tests, I can't guarantee that
// the behavior tested in accurate.  In addition to componentpacks, the old code references *.inline.css as "theme inline css" but
// I'm unable to determine where/how those files would exist. In short, this code should be throughly reviewed for accuracy and
// completeness across all metadata types.
//
// if successful in parsing the details, will return the entity name, path to the entity relative to the metadata type directory,
// true/false if the file contains the metadata definition for the entity and true for valid.  If unsuccessful in parsing the details
// will return "", "", false, false - for example:
//  1. pages/my_page.xml will return my_page, my_page, false, true
//  2. pages/my_page.json will return my_page, my_page, true, true
//  3. pages/my_page.txt will return "", "", false, false
//  4. site/favicon/my_icon.ico will return my_icon.ico, favicon/my_icon.ico, false, true
//  5. site/favicon/my_icon.ico.skuid.json will return my_icon.ico, favicon/my_icon.ico, true, true
//  6. site/favicon/my_icon.ico.json will return "", "", false, false
//  7. componentpacks/mypack/runtime.js will return mypack, mypack and false
func entityNameFromFilePath(details *entityPathDetails) (string, string, bool, bool) {
	directory := path.Dir(details.PathRelative)
	fileName := details.Name
	ext := path.Ext(fileName)
	fileNameWithoutExtension := strings.TrimSuffix(fileName, ext)

	switch details.Type {
	case MetadataTypeFiles:
		if details.SubType != MetadataSubTypeNone {
			return "", "", false, false
		} else if !fileEntityNameValidator.MatchString(fileName) {
			return "", "", false, false
		} else if s, found := strings.CutSuffix(fileName, ".skuid.json"); found {
			return s, s, true, true
		} else {
			return fileName, fileName, false, true
		}
	// Skuid Review Required - For favicon & logo, the Web UI indicates that only ico & png/jpg/gif are supported respectively.
	// How does skuid evaluate validty - by extension and if so, which are valid?  by mime-type and if so, which are valid?
	// TODO: Modify below and update tests based on answers to these questions
	case MetadataTypeSite:
		switch details.SubType {
		case MetadataSubTypeNone:
			if fileName == (EntityNameSite + ".json") {
				return EntityNameSite, EntityNameSite, true, true
			} else {
				return "", "", false, false
			}
		case MetadataSubTypeSiteFavicon:
			if !fileEntityNameValidator.MatchString(fileName) {
				return "", "", false, false
			} else if strings.HasSuffix(fileName, ".ico.skuid.json") {
				entityName := strings.TrimSuffix(fileName, ".skuid.json")
				return entityName, path.Join(directory, entityName), true, true
			} else if ext == ".ico" {
				return fileName, path.Join(directory, fileName), false, true
			} else {
				return "", "", false, false
			}
		case MetadataSubTypeSiteLogo:
			if !fileEntityNameValidator.MatchString(fileName) {
				return "", "", false, false
			} else if hasSuffix([]string{".png.skuid.json", ".jpg.skuid.json", ".gif.skuid.json"}, fileName) {
				entityName := strings.TrimSuffix(fileName, ".skuid.json")
				return entityName, path.Join(directory, entityName), true, true
			} else if slices.Contains([]string{
				".png",
				".jpg",
				".gif",
			}, ext) {
				return fileName, path.Join(directory, fileName), false, true
			} else {
				return "", "", false, false
			}
		default:
			return "", "", false, false
		}
	case MetadataTypePages:
		if details.SubType != MetadataSubTypeNone {
			return "", "", false, false
		} else if !entityNameValidator.MatchString(fileNameWithoutExtension) {
			return "", "", false, false
		} else if ext == ".json" || ext == ".xml" {
			return fileNameWithoutExtension, fileNameWithoutExtension, ext == ".json", true
		} else {
			return "", "", false, false
		}
	case MetadataTypeComponentPacks:
		// each component pack should be in its own directory
		if details.SubType != MetadataSubTypeNone || directory == "." {
			return "", "", false, false
		}
		// Skuid Review Required - The previous logic would match on any file as long as it was in a subdirectory
		// under componentpacks.  This would include js, css, json, extensionless, txt, jpg, png, etc. as well as any
		// character in any portion of the path and any level of subdirectories, all of which seems rather broad and
		// unlikely to be correct.  What are the accepted file formats/extensions/directory structures for component packs?
		// TODO: Adjust validation and update tests based on answers to above questions
		dirSplit := strings.Split(directory, "/")
		if entityNameValidator.MatchString(dirSplit[0]) {
			// Skuid Review Required - Unclear if there is a "definition file" for Component packs.  From reviewing docs,
			// there seems to be a skuid_runtime.json & skuid_builders.json but these are user controlled and also
			// not required to be these names per https://docs.skuid.com/latest/v1/en/skuid/components/component-packs/build/#runtime-definition-manifest
			// and https://docs.skuid.com/latest/v1/en/skuid/components/component-packs/build/#builder-definition-manifest.
			// Is there a "definition" file for a component pack that is controlled by Skuid?
			// TODO: Adjust the "IsEntityDefinitionFile" return value based on answers to above.  For now, given above since it
			// appears that the "manifests" are the "definition files" and both are user controlled, not marking any files
			// in component packs as "definition file" for now.
			return dirSplit[0], dirSplit[0], false, true
		} else {
			return "", "", false, false
		}
	case MetadataTypeThemes:
		// Skuid Review Required - Not sure what metadata type supports .inline.css files but the old code referenced it related
		// to theems.
		// TODO: If .inline.css is no longer supported, remove it, else adjust as needed based on current metadata types and scenarios
		// it is supported in
		if details.SubType != MetadataSubTypeNone {
			return "", "", false, false
		} else if s, found := strings.CutSuffix(fileName, ".inline.css"); found {
			return s, s, false, true
		} else if !entityNameValidator.MatchString(fileNameWithoutExtension) {
			return "", "", false, false
		} else if ext == ".json" {
			return fileNameWithoutExtension, fileNameWithoutExtension, true, true
		} else {
			return "", "", false, false
		}
	default:
		if details.SubType != MetadataSubTypeNone {
			return "", "", false, false
		} else if !entityNameValidator.MatchString(fileNameWithoutExtension) {
			return "", "", false, false
		} else if ext == ".json" {
			return fileNameWithoutExtension, fileNameWithoutExtension, true, true
		} else {
			return "", "", false, false
		}
	}
}

func hasSuffix[S []string](s S, v string) bool {
	for _, x := range s {
		if strings.HasSuffix(v, x) {
			return true
		}
	}
	return false
}

func sortEntities(a, b MetadataEntity) int {
	if a.Type != b.Type {
		return strings.Compare(a.Type.Name(), b.Type.Name())
	}

	return strings.Compare(a.Path, b.Path)
}
