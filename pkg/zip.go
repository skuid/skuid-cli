package pkg

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"sync/atomic"

	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
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
type ArchiveFilter func(relativePath string) bool

type archiveItem struct {
	Bytes    []byte
	FilePath string
}

var (
	ErrArchiveNoMetadataTypeDirs = errors.New("unable to create archive, no metadata directories found")
	ErrArchiveNoFiles            = errors.New("unable to create archive, no files were found")
)

// filters files not in NlxMetadata
// TODO: Skuid Review Required - See comment above ArchiveFilter type in this file
func MetadataArchiveFilter(filter *NlxMetadata) ArchiveFilter {
	return func(relativePath string) bool {
		return filter != nil && filter.FilterItem(relativePath)
	}
}

// filters files not in list of files
func FileNameArchiveFilter(files []string) ArchiveFilter {
	return func(relativePath string) bool {
		// NOTE - As of golang v1.21, slices package includes a Contains method (slices.Contains(files, relativePath))
		// however in order to support >= v1.20, unable to use it.
		// TODO: If/When skuid-cli states an official minimum supported go version and if/when that version
		// is >= v1.21, the slices Contains can be called directly instead of using custom util StringSliceContainsKey
		return util.StringSliceContainsKey(files, relativePath)
	}
}

// filters files whose metadata type is in excludedMetadataDirs
// TODO: Skuid Review Required - See comment above ArchiveFilter type
func MetadataDirArchiveFilter(excludedMetadataDirs []string) ArchiveFilter {
	hasExcludedDirs := len(excludedMetadataDirs) > 0
	return func(relativePath string) bool {
		if !hasExcludedDirs {
			return true
		}
		// NOTE - As of golang v1.21, slices package includes a Contains method (slices.Contains(files, relativePath))
		// however in order to support >= v1.20, unable to use it.
		// TODO: If/When skuid-cli states an official minimum supported go version and if/when that version
		// is >= v1.21, the slices Contains can be called directly instead of using custom util StringSliceContainsKey
		metadataType, _ := GetEntityDetails(relativePath)
		return !util.StringSliceContainsKey(excludedMetadataDirs, metadataType)
	}
}

func Archive(fsys fs.FS, fileUtil util.FileUtil, filterArchive ArchiveFilter) ([]byte, int, error) {
	buffer := new(bytes.Buffer)
	zipWriter := fileUtil.NewZipWriter(buffer)
	g, ctx := errgroup.WithContext(context.Background())

	filePaths := getFiles(ctx, g, fsys, fileUtil, filterArchive)
	fileContents := readFiles(ctx, g, fsys, fileUtil, filePaths)
	archivedFiles := addFiles(ctx, g, zipWriter, fileContents)

	var fileCount int
	for range archivedFiles {
		fileCount++
	}

	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	if fileCount == 0 {
		return nil, 0, ErrArchiveNoFiles
	}

	if err := zipWriter.Close(); err != nil {
		return nil, 0, err
	}

	if result, err := io.ReadAll(buffer); err != nil {
		return nil, 0, err
	} else {
		return result, fileCount, nil
	}
}

// returns the list of metadata type directory names that exist based on all known metadata type dir names
func getMetadataTypeDirNamesForArchive(fsys fs.FS, fileUtil util.FileUtil) (dirNames []string, err error) {
	for _, mddir := range GetMetadataTypeDirNames() {
		var exists bool
		exists, err = fileUtil.DirExists(fsys, mddir)
		if err != nil {
			logging.Get().Errorf("unable to detect if metadata path is a directory: %v", mddir)
			return
		}
		if !exists {
			logging.Get().Tracef("metadata directory does not exist, skipping: %v", mddir)
			continue
		}
		dirNames = append(dirNames, mddir)
	}

	if len(dirNames) == 0 {
		err = ErrArchiveNoMetadataTypeDirs
		return
	}

	return
}

func getFiles(ctx context.Context, g *errgroup.Group, fsys fs.FS, fileUtil util.FileUtil, filterArchive ArchiveFilter) <-chan string {
	archivePaths := make(chan string)

	g.Go(func() error {
		metadataTypeDirNames, err := getMetadataTypeDirNamesForArchive(fsys, fileUtil)
		if err != nil {
			close(archivePaths)
			return err
		}

		var fileWorkers atomic.Int32
		fileWorkers.Store(int32(len(metadataTypeDirNames)))
		for i, mddir := range metadataTypeDirNames {
			workerNum := i
			metadataDirName := mddir
			g.Go(func() error {
				defer func() {
					// Last one out closes shop
					if fileWorkers.Add(-1) == 0 {
						close(archivePaths)
					}
				}()
				return getMetadataFiles(ctx, fsys, fileUtil, metadataDirName, archivePaths, filterArchive, workerNum)
			})

		}

		return nil
	})

	return archivePaths
}

func getMetadataFiles(ctx context.Context, fsys fs.FS, fileUtil util.FileUtil, metadataDirName string, archivePaths chan<- string, filterArchive ArchiveFilter, workerNum int) (err error) {
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

		// TODO: Skuid Review Required - See comment above ArchiveFilter type in this file
		if filterArchive != nil && !filterArchive(filePath) {
			return nil
		}

		// NOTE: select will pseudo-randomly choose a case when both are ready so we could send even though we are cancelled at this point.  See
		// above ctx.Err() to ensure we don't start any new files once cancel occurred.  See https://go.dev/ref/spec#Select_statements.
		select {
		case archivePaths <- filePath:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

func readFiles(ctx context.Context, g *errgroup.Group, fsys fs.FS, fileUtil util.FileUtil, archivePaths <-chan string) <-chan archiveItem {
	archiveItems := make(chan archiveItem)

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
		workerNum := i
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if itemWorkers.Add(-1) == 0 {
					close(archiveItems)
				}
			}()

			return readFile(ctx, fsys, fileUtil, archivePaths, archiveItems, workerNum)
		})
	}

	return archiveItems
}

func addFiles(ctx context.Context, g *errgroup.Group, zipWriter util.ZipWriter, archiveItems <-chan archiveItem) <-chan struct{} {
	// we do not need the actual filename so use empty struct to optimize perf
	archivedFiles := make(chan struct{})

	// add files to zip - must be single worker as archive/zip package does not appear to be thread-safe (e.g., must
	// write file contents to io.Writer prior to calling Create again - see https://pkg.go.dev/archive/zip#Writer.Create
	g.Go(func() error {
		defer func() {
			close(archivedFiles)
		}()

		for item := range archiveItems {
			// exit if cancelled - see note below in select
			if ctx.Err() != nil {
				return ctx.Err()
			}

			var zipFileWriter io.Writer

			zipFileWriter, err := zipWriter.Create(item.FilePath)
			if err != nil {
				return err
			}
			_, err = zipFileWriter.Write(item.Bytes)
			if err != nil {
				return err
			}

			// NOTE: select will pseudo-randomly choose a case when both are ready so we could send even though we are cancelled at this point.  See
			// above ctx.Err() to ensure we don't start any new files once cancel occurred.  See https://go.dev/ref/spec#Select_statements.
			select {
			case archivedFiles <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	return archivedFiles
}

func readFile(ctx context.Context, fsys fs.FS, fileUtil util.FileUtil, archivePaths <-chan string, archiveItems chan archiveItem, workerNum int) error {
	for path := range archivePaths {
		// exit if cancelled - see note below in select
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fileBytes, err := fileUtil.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		// NOTE: select will pseudo-randomly choose a case when both are ready so we could send even though we are cancelled at this point.  See
		// above ctx.Err() to ensure we don't start any new files once cancel occurred.  See https://go.dev/ref/spec#Select_statements.
		select {
		case archiveItems <- archiveItem{fileBytes, path}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
