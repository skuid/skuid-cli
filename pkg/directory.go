package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gookit/color"

	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

func ClearDirectories(targetDir string) (err error) {
	if !filepath.IsAbs(targetDir) {
		err = fmt.Errorf("targetDir must be an absolute path")
		return
	}

	for _, member := range metadata.MetadataTypes.Members() {
		dirPath := filepath.Join(targetDir, member.DirName())
		logging.Get().Debugf("%v: %v", color.Yellow.Sprint("Removing"), dirPath)
		if err = os.RemoveAll(dirPath); err != nil {
			return
		}
	}

	return
}
