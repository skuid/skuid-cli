package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gookit/color"

	"github.com/skuid/skuid-cli/pkg/logging"
)

func ClearDirectories(targetDir string) (err error) {
	if !filepath.IsAbs(targetDir) {
		err = fmt.Errorf("targetDir must be an absolute path")
		return
	}

	for _, dirName := range GetMetadataTypeDirNames() {
		dirPath := filepath.Join(targetDir, dirName)
		logging.Get().Debugf("%v: %v", color.Yellow.Sprint("Removing"), dirPath)
		if err = os.RemoveAll(dirPath); err != nil {
			return
		}
	}

	return
}
