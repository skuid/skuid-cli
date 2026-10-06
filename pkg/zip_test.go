package pkg_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/util"
)

func TestZip(t *testing.T) {
	util.SkipIntegrationTest(t)
	cd, _ := os.Getwd()
	relpath := filepath.Join(cd, ".", ".", "_deploy")
	bb, err := pkg.Archive(relpath, nil)
	if err != nil {
		logging.Get().Fatal(err)
	}
	logging.Get().Info(len(bb))

}

func TestArchiveMetadata(t *testing.T) {
	dir := t.TempDir()
	for _, file := range []string{
		"manifest.json",
		"pages/MyPage.json",
		"pages/MyPage.xml",
		"README.md",
		"skuid",
		"node_modules/left-pad/index.js",
	} {
		path := filepath.Join(dir, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(file), 0644); err != nil {
			t.Fatal(err)
		}
	}

	archive, err := pkg.ArchiveMetadata(dir)
	if err != nil {
		t.Fatal(err)
	}

	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatal(err)
	}
	var archived []string
	for _, file := range reader.File {
		archived = append(archived, filepath.ToSlash(file.Name))
	}

	assert.ElementsMatch(t, []string{
		"manifest.json",
		"pages/MyPage.json",
		"pages/MyPage.xml",
	}, archived)
}
