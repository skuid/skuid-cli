package pkg

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bobg/go-generics/v4/set"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	pathMap = make(map[string]bool, 0)
)

type WritePayload struct {
	PlanName PlanName
	PlanData []byte
}

func WriteResultsToDisk(targetDirectory string, result WritePayload) (set.Of[string], error) {
	return WriteResults(targetDirectory, result, util.CopyToFile, util.CreateDirectoryDeep, os.ReadFile)
}

func WriteResults(targetDirectory string, result WritePayload, copyToFile util.FileCreator, createDirectoryDeep util.DirectoryCreator, ioutilReadFile util.FileReader) (entityPaths set.Of[string], err error) {
	message := fmt.Sprintf("Writing entity results for plan %v to directory %v", logging.QuoteText(result.PlanName), logging.ColorResource.QuoteText(targetDirectory))
	fields := logging.Fields{
		"targetDirectory":   targetDirectory,
		"planName":          result.PlanName,
		"resultPlanDataLen": len(result.PlanData),
	}
	logger := logging.WithTraceTracking("pkg.WriteResults", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if !filepath.IsAbs(targetDirectory) {
		return nil, fmt.Errorf("targetDirectory %v must be an absolute path", logging.QuoteText(targetDirectory))
	}

	if err := createDirectoryDeep(targetDirectory, 0755); err != nil {
		return nil, fmt.Errorf("unable to create directory %v: %w", logging.QuoteText(targetDirectory), err)
	}

	tmpFileName, err := util.CreateTemporaryFile(string(result.PlanName), ".zip", result.PlanData)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpFileName) // intentionally ignoring any error

	// unzip the contents of our temp zip file
	entityPaths, err = unzipArchive(
		tmpFileName,
		targetDirectory,
		copyToFile,
		createDirectoryDeep,
		ioutilReadFile,
		logger,
	)
	if err != nil {
		return nil, err
	}

	logger = logger.WithSuccess(logging.Fields{
		"entityPathLen": len(entityPaths),
	})
	return entityPaths, nil
}

func ResetPathMap() {
	pathMap = make(map[string]bool)
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func unzipArchive(sourceFileLocation string, targetLocation string, fileCreator util.FileCreator, directoryCreator util.DirectoryCreator, existingFileReader util.FileReader, logger *logging.Logger) (entityPaths set.Of[string], err error) {
	message := fmt.Sprintf("Unzipping archive %v to %v", logging.QuoteText(sourceFileLocation), logging.ColorResource.QuoteText(targetLocation))
	fields := logging.Fields{
		"sourceFileLocation": sourceFileLocation,
		"targetLocation":     targetLocation,
	}
	logger = logger.WithTraceTracking("unzipArchive", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	var reader *zip.ReadCloser
	if reader, err = zip.OpenReader(sourceFileLocation); err != nil {
		return nil, fmt.Errorf("unable to open archive %v: %w", logging.QuoteText(sourceFileLocation), err)
	}
	defer reader.Close()

	if err = directoryCreator(targetLocation, 0755); err != nil {
		return nil, err
	}

	entityPaths = set.New[string]()
	for _, file := range reader.File {
		if entityFile, err := extractArchiveFile(sourceFileLocation, targetLocation, fileCreator, directoryCreator, existingFileReader, file, logger); err != nil {
			// Skuid Review Required - This code is maintains the behavior of v0.6.7 by simply skipping files in the archive that are not valid entity files.  However,
			// it raises questions around leaving the local file system in an unexpected state.  The server gave us a file and we can't interpret it so something more
			// than logging and continuing seems the more prudent approach.
			//
			// TODO: We need to do something more than just "skip" and log an error - not sure what can be done but we received an invalid file
			// here and it warrants more than just "ok, thanks but no thanks, moving on".  Possibly the approach to unzipping changes where the payload
			// is inspected prior to any local file changes (e.g, don't even delete the existing dirs until payload is validated) and only if valid, then local
			// file changes made
			var pErr *metadata.MetadataPathError
			if errors.As(err, &pErr) && pErr.PathType == metadata.MetadataPathTypeEntityFile {
				logger.WithError(err).Warnf("Unable to write file to disk, could not resolve metadata information for file %v in archive %v: %v", logging.QuoteText(file.Name), logging.QuoteText(sourceFileLocation), err)
			} else {
				return nil, err
			}
		} else {
			entityPaths.Add(entityFile.Entity.Path)
		}
	}

	logger = logger.WithSuccess()
	return entityPaths, nil
}

func extractArchiveFile(sourceFileLocation, targetLocation string, fileCreator util.FileCreator, directoryCreator util.DirectoryCreator, existingFileReader util.FileReader, file *zip.File, logger *logging.Logger) (entityFile *metadata.MetadataEntityFile, err error) {
	entityFileArchivePath := filepath.FromSlash(file.Name)
	message := fmt.Sprintf("Extracting file from archive %v to %v", logging.QuoteText(sourceFileLocation), logging.ColorResource.QuoteText(filepath.Join(targetLocation, file.Name)))
	fields := logging.Fields{
		"sourceFileLocation":    sourceFileLocation,
		"targetLocation":        targetLocation,
		"entityFileArchivePath": entityFileArchivePath,
	}
	logger = logger.WithTraceTracking("extractArchiveFile", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	// Check to see if we've already written to this file in this retrieve.  This can occur when a metadata service result contains an entity and the data service result contains
	// the same entity.  The results need to be merged together to produce the final file locally.
	_, fileAlreadyWritten := pathMap[entityFileArchivePath]
	if !fileAlreadyWritten {
		pathMap[entityFileArchivePath] = true
	}

	entityFile, err = metadata.NewMetadataEntityFile(entityFileArchivePath)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(targetLocation, entityFile.Path)
	fileDir := filepath.Dir(filePath)
	if err := directoryCreator(fileDir, 0755); err != nil {
		return nil, err
	}

	// Skuid Review Required - See https://github.com/skuid/skuid-cli/issues/199
	// In 0.6.7, if a directory was encountered in the archive, the code (https://github.com/skuid/skuid-cli/blob/master/pkg/util/archive.go#L79)
	// would call directoryCreator and return the value from it resulting in the entire function exiting leaving any remaining files yet to be
	// processed in the archive unprocessed.  Assuming directoryCreator successfully created the directory on disk, no error message would be
	// emitted to the user leaving them to think everything was retrieved when it wasn't. The question is was this intentional or a bug?
	// Assuming this is a bug since it wouldn't make sense to abort the unzip simply because a naked directory was unencountered (unless naked
	// directories are not expected to be in the archive and if one is it should be considered an "invalid" archive?), I am leaving the check
	// for a directory and if detected, creating it, but altering the behavior to continue processing the archive unless directoryCreator itself
	// returns an error.
	//
	// TODO: Adjust as needed based on answer to above
	if file.FileInfo().IsDir() {
		logger.Warnf("Detected directory %v in archive %v", logging.QuoteText(file.Name), logging.QuoteText(sourceFileLocation))
		// no file to process, just return result of creating the directory so calling function can handle accordingly
		return nil, directoryCreator(filePath, file.Mode())
	}

	fileReader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open file %v in archive %v: %w", logging.QuoteText(file.Name), logging.QuoteText(sourceFileLocation), err)
	}
	defer fileReader.Close()

	// Skuid Review Required - This code has been modified to applying "SanitizeZip" to all entity definition files.  The previous
	// code did not apply this to all definition files which I believe was/is the intent of what is being done here.  The question
	// is to confirm what the intent was/is and if the change made is correct to account for the issue described in
	// https://github.com/skuid/skuid-cli/issues/198?
	// Sanitize all metadata definition
	if entityFile.IsEntityDefinitionFile {
		if sanitizedFileReader, err := sanitizeZipFile(file.Name, fileReader, logger); err != nil {
			return nil, fmt.Errorf("unable to sanitize file %v in archive %v: %w", logging.QuoteText(file.Name), logging.QuoteText(sourceFileLocation), err)
		} else {
			fileReader.Close()
			fileReader = sanitizedFileReader
		}
	}

	// A retrieval can result in 0, 1 or 2 different archives (no results, metadata/data service result or both) that all need
	// merged and written to since they can contain information regarding the same entities.  We keep track of whether or
	// not a file has already been written from a previously processed archive and if so, merge
	// the current in with it
	if fileAlreadyWritten {
		if combinedFileReader, err := util.CombineJSON(fileReader, existingFileReader, filePath); err != nil {
			return nil, err
		} else {
			fileReader.Close()
			fileReader = combinedFileReader
		}
	}

	if err := fileCreator(fileReader, filePath); err != nil {
		return nil, err
	}

	logger = logger.WithSuccess()
	return entityFile, nil
}

func sanitizeZipFile(path string, inReader io.ReadCloser, logger *logging.Logger) (outReader io.ReadCloser, err error) {
	message := fmt.Sprintf("Sanitizing zip file %v", logging.ColorResource.QuoteText(path))
	fields := logging.Fields{"path": path}
	logger = logger.WithTraceTracking("sanitizeZipFile", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	b, err := io.ReadAll(inReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %v: %w", logging.QuoteText(path), err)
	}

	b, err = util.ReSortJson(b, true)
	if err != nil {
		return nil, err
	}

	logger = logger.WithSuccess()
	return io.NopCloser(bytes.NewBuffer(b)), nil
}
