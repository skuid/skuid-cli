package util

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	jsonpatch "github.com/skuid/json-patch"

	"github.com/skuid/tides/pkg/logging"
)

func CombineJSON(readCloser io.ReadCloser, fileReader FileReader, path string) (newReadCloser io.ReadCloser, err error) {
	fields := logrus.Fields{
		"function": "CombineJSON",
	}

	logging.WithFields(fields).Tracef("Augmenting File with more JSON Data: %v\n", color.Magenta.Sprint(path))
	existingBytes, err := fileReader(path)
	if err != nil {
		logging.Get().Warnf("existingFileReader: %v", err)
		return
	}

	newBytes, err := ioutil.ReadAll(readCloser)
	if err != nil {
		logging.Get().Warnf("ioutil.ReadAll: %v", err)
		return
	}

	// merge the files together using the json patch library
	combined, err := jsonpatch.MergePatch(existingBytes, newBytes)
	if err != nil {
		logging.Get().Warnf("jsonpatch.MergePatch: %v", err)
		return
	}

	// sort all of the keys in the json. custom sort logic.
	// this puts "name" first, then everything alphanumerically
	sorted, err := ReSortJsonIndent(combined, true)
	if err != nil {
		logging.Get().Warnf("ReSortJsonIndent: %v", err)
		return
	}

	var indented bytes.Buffer
	err = json.Indent(&indented, sorted, "", "\t")
	if err != nil {
		logging.Get().Warnf("Indent: %v", err)
		return
	}

	newReadCloser = ioutil.NopCloser(bytes.NewReader(indented.Bytes()))

	return
}
