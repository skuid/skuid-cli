package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

func ClearDirectories(targetDirectory string) (err error) {
	message := fmt.Sprintf("Removing directory %v", logging.ColorResource.QuoteText(targetDirectory))
	fields := logging.Fields{
		"targetDirectory": targetDirectory,
	}
	logger := logging.WithTracking("pkg.ClearDirectories", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if !filepath.IsAbs(targetDirectory) {
		return fmt.Errorf("targetDirectory %v must be an absolute path", logging.QuoteText(targetDirectory))
	}

	for _, member := range metadata.MetadataTypes.Members() {
		dirPath := filepath.Join(targetDirectory, member.DirName())
		logger.Tracef("Removing directory %v", logging.ColorResource.QuoteText(dirPath))
		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("unable to remove directory %v: %w", logging.QuoteText(dirPath), err)
		}
	}

	logger = logger.WithSuccess()
	return nil
}
