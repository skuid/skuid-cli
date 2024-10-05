package util

import (
	"bytes"
	"fmt"
	"io"

	jsonpatch "github.com/skuid/json-patch"

	"github.com/goccy/go-json"
	"github.com/skuid/skuid-cli/pkg/logging"
)

func CombineJSON(inReader io.ReadCloser, fileReader FileReader, path string) (outReader io.ReadCloser, err error) {
	message := fmt.Sprintf("Adding additional JSON data to file %v", logging.QuoteText(path))
	fields := logging.Fields{
		"path": path,
	}
	logger := logging.WithTraceTracking("CombineJSON", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	existingBytes, err := fileReader(path)
	if err != nil {
		return nil, err
	}

	newBytes, err := io.ReadAll(inReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read new bytes to merge with file %v: %w", logging.QuoteText(path), err)
	}

	// merge the files together using the json patch library
	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to patch new bytes with file %v: %w", logging.QuoteText(path), err)
	}

	// sort all of the keys in the json. custom sort logic.
	// this puts "name" first, then everything alphanumerically
	sorted, err := ReSortJson(combined, true)
	if err != nil {
		return nil, err
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, sorted, "", "\t"); err != nil {
		return nil, fmt.Errorf("unable to indent merged json for file %v: %w", logging.QuoteText(path), err)
	}

	logger = logger.WithSuccess()
	return io.NopCloser(bytes.NewReader(indented.Bytes())), nil
}
