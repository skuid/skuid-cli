package pkg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync/atomic"

	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/skuid/skuid-cli/pkg/util"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

/*
	TODO: Skuid Review Required - The previous version of Archive had some (potential, but likely) shortcomings:

		1. It would not filter out files that "looked" like metadata objects but are not valid metadata objects.
		   For example, if a file (other than a Skuid File metadata object which can be extensionless) had no file extension but valid
		   object name. This is of particular note during the ExecuteDeployPlan step since there could be a file "pages/test_page"
		   that will match a page "test_page" but the file itself does not contain .xml/.json extension so it will be ignored by server.
		   More importantly, there could be a file pages/test_page.xml & pages/test_page and they will both be included in the payload.
		   This all seems rather odd given we are walking the file system during archive.  The reason it would not filter the file was
		   because it relied on NlxMetadata.FilterItem which matches only on an item name of the item passed in (it strips any known
		   extension to derive the name of the item). Keeping the same behavior for now for consistency as I don't know enough about
		   what is valid and what is not valid but likely the updated Archive function below should be changed to ensure the filename
		   is a  valid filename based on its metadata type (e.g., pages files would have to have an extension of either .xml & .json).

		2. In ExecuteDeployPlan, the result from GetDeployPlan is used to filter out files that aren't present in the generated plan.
		   It does this by looking only at known metadata directories (obtained from NlxMetadata tag via GetFieldValueByName function)
		   which results in anything not in a known metadata directory to not be included in the ExecuteDeployPlan payload.  However,
		   when generating the payload for GetDeployPlan, all directories are walked, not just known metadata directories. Given how
		   ExecuteDeployPlan works in 0.67.0, there does not seem to be any reason to "walk" any directories other than known metadata
		   directories for all requests to Archive.  Limiting to known metadata directories eliminates several issues that are present
		   in 0.67.0.  Given this, the version of Archive below (used to create payload for GenerateDeployPlan and ExecuteDeployPlan).
		   The approach I've taken should be validated since without internal knowledge of how Skuid works I can't say for certain that
		   the modified approach I've taken is valid (although it seems likely given the way ExecuteDeployPlan works in 0.67.0).

	see related issues:
		https://github.com/skuid/skuid-cli/issues/159
		https://github.com/skuid/skuid-cli/issues/163
		https://github.com/skuid/skuid-cli/issues/164

*/

// true if relativePath meets criteria, false otherwise
type ArchiveFilter func(metadata.MetadataEntityFile) bool

type archiveFile struct {
	Bytes      []byte
	EntityFile metadata.MetadataEntityFile
}

var (
	ErrArchiveNoFiles = errors.New("unable to create archive, no files were found")
)

// filters files not in NlxMetadata
// TODO: Skuid Review Required - See comment above ArchiveFilter type in this file
func MetadataArchiveFilter(filter *metadata.NlxMetadata) ArchiveFilter {
	return func(item metadata.MetadataEntityFile) bool {
		return filter != nil && filter.FilterItem(item)
	}
}

// filters files not in list of files
func MetadataEntityArchiveFilter(entities []metadata.MetadataEntity) ArchiveFilter {
	return func(item metadata.MetadataEntityFile) bool {
		return slices.Contains(entities, item.Entity)
	}
}

// filters files whose metadata type is in excludedMetadataDirs
// TODO: Skuid Review Required - See comment above ArchiveFilter type
func MetadataTypeArchiveFilter(excludedMetadataTypes []metadata.MetadataType) ArchiveFilter {
	hasExcludedDirs := len(excludedMetadataTypes) > 0
	return func(item metadata.MetadataEntityFile) bool {
		if !hasExcludedDirs {
			return true
		}
		return !slices.Contains(excludedMetadataTypes, item.Entity.Type)
	}
}

func Archive(fsys fs.FS, fileUtil util.FileUtil, filterArchive ArchiveFilter) ([]byte, []string, []metadata.MetadataEntity, error) {
	buffer := new(bytes.Buffer)
	zipWriter := fileUtil.NewZipWriter(buffer)
	g, ctx := errgroup.WithContext(context.Background())

	entityFiles := getFiles(ctx, g, fsys, fileUtil, filterArchive)
	filesToArchive := readFiles(ctx, g, fsys, fileUtil, entityFiles)
	archivedFiles := addFiles(ctx, g, zipWriter, filesToArchive)

	var archivedFilePaths []string
	archivedEntities := make(map[metadata.MetadataEntity][]archiveFile)
	for archivedFile := range archivedFiles {
		entity := archivedFile.EntityFile.Entity
		archivedEntities[entity] = append(archivedEntities[entity], archivedFile)
		archivedFilePaths = append(archivedFilePaths, archivedFile.EntityFile.Path)
	}

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}

	if len(archivedEntities) == 0 {
		return nil, nil, nil, ErrArchiveNoFiles
	}

	if err := zipWriter.Close(); err != nil {
		return nil, nil, nil, err
	}

	if result, err := io.ReadAll(buffer); err != nil {
		return nil, nil, nil, err
	} else {
		return result, archivedFilePaths, maps.Keys(archivedEntities), nil
	}
}

func getFiles(ctx context.Context, g *errgroup.Group, fsys fs.FS, fileUtil util.FileUtil, filterArchive ArchiveFilter) <-chan metadata.MetadataEntityFile {
	entityFiles := make(chan metadata.MetadataEntityFile)

	g.Go(func() error {
		var fileWorkers atomic.Int32
		fileWorkers.Store(int32(metadata.MetadataTypes.Len()))
		for _, mdt := range metadata.MetadataTypes.Members() {
			metadataType := mdt
			g.Go(func() error {
				defer func() {
					// Last one out closes shop
					if fileWorkers.Add(-1) == 0 {
						close(entityFiles)
					}
				}()
				metadataDirName := metadataType.DirName()
				if exists, err := fileUtil.DirExists(fsys, metadataDirName); err != nil {
					return fmt.Errorf("unable to detect if metadata path %q is a directory: %w", metadataDirName, err)
				} else if !exists {
					logging.Get().Tracef("metadata directory does not exist, skipping: %v", metadataDirName)
					return nil
				}
				return getMetadataEntityFiles(ctx, fsys, fileUtil, metadataType.DirName(), entityFiles, filterArchive)
			})
		}

		return nil
	})

	return entityFiles
}

func getMetadataEntityFiles(ctx context.Context, fsys fs.FS, fileUtil util.FileUtil, metadataDirName string, entityFiles chan<- metadata.MetadataEntityFile, filterArchive ArchiveFilter) error {
	return fileUtil.WalkDir(fsys, metadataDirName, func(filePath string, dirEntry fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		// exit if cancelled - see note below in select
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if dirEntry.IsDir() {
			return nil
		}

		// create item
		entityFile, err := metadata.NewMetadataEntityFile(filePath)
		if err != nil {
			return err
		}

		// TODO: Skuid Review Required - See comment above ArchiveFilter type in this file
		if filterArchive != nil && !filterArchive(*entityFile) {
			return nil
		}

		// NOTE: select will pseudo-randomly choose a case when both are ready so we could send even though we are cancelled at this point.  See
		// above ctx.Err() to ensure we don't start any new files once cancel occurred.  See https://go.dev/ref/spec#Select_statements.
		select {
		case entityFiles <- *entityFile:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

func readFiles(ctx context.Context, g *errgroup.Group, fsys fs.FS, fileUtil util.FileUtil, entityFiles <-chan metadata.MetadataEntityFile) <-chan archiveFile {
	filesToArchive := make(chan archiveFile)

	/*
		TODO: Skuid Review Required

		In version 0.67.0, a separate go routine was started for every file to read it and send to a channel.  However, this approach
		had issues (e.g., channels weren't closed, errors didn't cancel context, etc.).  In the version below, a fixed number of workers
		is used to read files.  This approach was taken for three reasons:

		  1. Starting a separate go routine for every file, while possible, would require a separate, but nested, errgroup to be created.
		     The nested errgroup is required because it would be the only way to know when to close the sending channel unless we wait
			 for all filenames to be known then keep track of how many have been read.  While technically possible, it makes the code more
			 complex (although fairly striaghtforward).
		  2. Writing to the zip file is restricted to a single routine given that archive/zip does not seem to support concurrent
		     writes (see note in addFiles function).  Given this and with writing to a zip of likely bottleneck anyway, 20 workers to
			 read seems sufficient.
		  3. Without the knowledge of typical site size (e.g., number of files) and any target SLA, optimizing this code right now
		     would just be overengineering since it would be doing so absent of the information on what to optmize for.


		If information is provided on SLA and/or optimization target, this code can be easily adjusted to accomidate (e.g., change the
		number of workers and/or move to nested errgroup)
	*/

	const numItemWorkers = 20
	var itemWorkers atomic.Int32
	itemWorkers.Store(numItemWorkers)
	for i := 0; i < numItemWorkers; i++ {
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if itemWorkers.Add(-1) == 0 {
					close(filesToArchive)
				}
			}()

			return readFile(ctx, fsys, fileUtil, entityFiles, filesToArchive)
		})
	}

	return filesToArchive
}

func addFiles(ctx context.Context, g *errgroup.Group, zipWriter util.ZipWriter, filesToArchive <-chan archiveFile) <-chan archiveFile {
	archivedFiles := make(chan archiveFile)

	// add files to zip - must be single worker as archive/zip package does not appear to be thread-safe (e.g., must
	// write file contents to io.Writer prior to calling Create again - see https://pkg.go.dev/archive/zip#Writer.Create
	g.Go(func() error {
		defer func() {
			close(archivedFiles)
		}()

		for archiveFile := range filesToArchive {
			// exit if cancelled - see note below in select
			if ctx.Err() != nil {
				return ctx.Err()
			}

			var zipFileWriter io.Writer

			zipFileWriter, err := zipWriter.Create(archiveFile.EntityFile.Path)
			if err != nil {
				return err
			}
			_, err = zipFileWriter.Write(archiveFile.Bytes)
			if err != nil {
				return err
			}

			// NOTE: select will pseudo-randomly choose a case when both are ready so we could send even though we are cancelled at this point.  See
			// above ctx.Err() to ensure we don't start any new files once cancel occurred.  See https://go.dev/ref/spec#Select_statements.
			select {
			case archivedFiles <- archiveFile:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	return archivedFiles
}

func readFile(ctx context.Context, fsys fs.FS, fileUtil util.FileUtil, entityFiles <-chan metadata.MetadataEntityFile, filesToArchive chan archiveFile) error {
	for entityFile := range entityFiles {
		// exit if cancelled - see note below in select
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fileBytes, err := fileUtil.ReadFile(fsys, entityFile.Path)
		if err != nil {
			return err
		}

		// NOTE: select will pseudo-randomly choose a case when both are ready so we could send even though we are cancelled at this point.  See
		// above ctx.Err() to ensure we don't start any new files once cancel occurred.  See https://go.dev/ref/spec#Select_statements.
		select {
		case filesToArchive <- archiveFile{fileBytes, entityFile}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
