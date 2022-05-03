package util_test

import (
	"testing"
)

// type ArchiveFilter struct {
// }

// func (a *ArchiveFilter) FilterItem(item string) bool {
// 	cleanRelativeFilePath := util.FromWindowsPath(item)
// 	directory := filepath.Dir(cleanRelativeFilePath)
// 	baseName := filepath.Base(cleanRelativeFilePath)

// 	// Find the lowest level folder
// 	dirSplit := strings.Split(directory, string(filepath.Separator))
// 	metadataType, subFolders := dirSplit[0], dirSplit[1:]
// 	filePathArray := append(subFolders, baseName)
// 	filePath := strings.Join(filePathArray, string(filepath.Separator))

// 	validMetadataNames, err := a.GetNamesForType(metadataType)
// 	if validMetadataNames == nil || len(validMetadataNames) == 0 {
// 		// If we don't have valid names for this directory, just skip this file
// 		return
// 	}

// 	if err != nil {
// 		logging.VerboseError("FilterMetadataItem error", err)
// 		return
// 	}

// 	if StringSliceContainsAnyKey(validMetadataNames, []string{
// 		// Most common case --- check for our metadata with .json stripped
// 		strings.TrimSuffix(filePath, ".json"),
// 		// See if our filePath is in the valid metadata, if so, we're done
// 		filePath,
// 	}) {
// 		keep = true
// 		return
// 	}

// 	// Check for children of a component pack
// 	if metadataType == "componentpacks" {
// 		filePathParts := strings.Split(filePath, string(filepath.Separator))
// 		if len(filePathParts) == 2 && StringSliceContainsKey(validMetadataNames, filePathParts[0]) {
// 			keep = true
// 			return
// 		}
// 	}

// 	if StringSliceContainsAnyKey(validMetadataNames, []string{
// 		// Check for our metadata with .xml stripped
// 		strings.TrimSuffix(filePath, ".xml"),
// 		// Check for our metadata with .skuid.json stripped
// 		strings.TrimSuffix(filePath, ".skuid.json"),
// 		// Check for theme inline css
// 		strings.TrimSuffix(filePath, ".inline.css"),
// 	}) {
// 		keep = true
// 		return
// 	}

// 	return
// }

// // Archive compresses a file/directory to a writer
// //
// // If the path ends with a separator, then the contents of the folder at that path
// // are at the root level of the archive, otherwise, the root of the archive contains
// // the folder as its only item (with contents inside).
// //
// // If progress is not nil, it is called for each file added to the archive.
// func Archive(inFilePath string, writer io.Writer, filter *ArchiveFilter) error {
// 	return ArchiveWithFilterFunc(inFilePath, writer, func(relativePath string) bool {
// 		// if there was a metadata filter, apply it.
// 		if filter != nil {
// 			return filter.FilterItem(relativePath)
// 		}
// 		return true
// 	})
// }

// // ArchivePartial compresses all files in a file/directory matching a relative prefix to a writer
// //
// // If the path ends with a separator, then the contents of the folder at that path
// // are at the root level of the archive, otherwise, the root of the archive contains
// // the folder as its only item (with contents inside).
// //
// // If progress is not nil, it is called for each file added to the archive.
// func ArchivePartial(inFilePath string, writer io.Writer, basePrefix string) error {
// 	return ArchiveWithFilterFunc(inFilePath, writer, func(relativePath string) bool {
// 		return strings.HasPrefix(relativePath, basePrefix)
// 	})
// }

// func ArchiveWithFilterFunc(inFilePath string, writer io.Writer, filter func(string) bool) error {
// 	zipWriter := zip.NewWriter(writer)

// 	basePath := filepath.Dir(inFilePath)

// 	err := filepath.Walk(inFilePath, func(filePath string, fileInfo os.FileInfo, err error) error {
// 		if err != nil || fileInfo.IsDir() {
// 			return err
// 		}

// 		relativeFilePath, err := filepath.Rel(basePath, filePath)
// 		if err != nil {
// 			return err
// 		}

// 		archivePath := path.Join(filepath.SplitList(relativeFilePath)...)

// 		if strings.HasPrefix(archivePath, ".") {
// 			return nil
// 		}

// 		if !filter(relativeFilePath) {
// 			return nil
// 		}

// 		file, err := os.Open(filePath)
// 		if err != nil {
// 			return err
// 		}
// 		defer func() {
// 			_ = file.Close()
// 		}()

// 		zipFileWriter, err := zipWriter.Create(archivePath)
// 		if err != nil {
// 			return err
// 		}

// 		_, err = io.Copy(zipFileWriter, file)
// 		return err
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	return zipWriter.Close()
// }

func TestZip(t *testing.T) {

}
