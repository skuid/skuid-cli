package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

type SkuidPullResponse struct {
	APIVersion         string  `json:"apiVersion"`
	Name               string  `json:"name"`
	UniqueID           string  `json:"uniqueId"`
	Type               string  `json:"type"`
	Module             string  `json:"module"`
	MaxAutoSaves       int     `json:"maxAutoSaves"`
	MasterPageUniqueID string  `json:"masterPageUniqueId,omitempty"`
	IsMasterPage       bool    `json:"isMasterPage"`
	ComposerSettings   *string `json:"composerSettings"`
	Body               string  `json:"body,omitempty"`
}

type SkuidPagePost struct {
	Changes   []SkuidPullResponse `json:"changes"`
	Deletions []SkuidPullResponse `json:"deletions"`
}

type SkuidPagePostResult struct {
	OrgName string   `json:"orgName"`
	Success bool     `json:"success"`
	Errors  []string `json:"upsertErrors,omitempty"`
}

// Create a file basename from the pull response
// I have no idea why we were using a buffer before
func (page *SkuidPullResponse) FileBasename() string {
	if page.Module != "" {
		return fmt.Sprintf("%v_%v", page.Module, page.Name)
	} else {
		return page.Name
	}
}

func (page *SkuidPullResponse) WriteAtRest(path string) (err error) {

	// if the desired directory isn't there, create it
	if _, err = os.Stat(path); err != nil {
		err = os.Mkdir(path, 0700)
		// if we can't make the directory return the error
		if err != nil {
			return err
		}
	}

	// create a copy of the page to keep the Body out of the json
	clone := *page
	clone.Body = ""
	str, _ := json.MarshalIndent(clone, "", "    ")
	//write the json metadata
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.json", path, page.FileBasename()), str, 0644)

	if err != nil {
		return err
	}
	pageXml := page.Body
	//write the body to the file
	formatted, err := FormatXml(pageXml)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.xml", path, page.FileBasename()), []byte(formatted), 0644)

	if err != nil {
		return err
	}

	return nil

}

func FormatXml(toFormat string) (string, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromString(toFormat)

	if err != nil {
		return "", err
	}

	doc.Indent(2)
	return doc.WriteToString()

}

func FilterByGlob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func FilterByModule(dir string, moduleFilter []string) ([]string, error) {

	filter := &bytes.Buffer{}

	if len(moduleFilter) > 0 {
		// TODO: make sure this works
		// this module filter (if there are more)
		// should be inclusive. we had "module(s)" as
		// part of the argument, so I turned it into an array of strings
		filter.WriteString(strings.Join(moduleFilter, "_") + "_")
	}

	filter.WriteString("*")

	pattern := filepath.Join(dir, filter.String())
	return filepath.Glob(pattern)
}

func filterOutXmlFiles(files []string) []string {
	filtered := []string{}

	for _, path := range files {
		if filepath.Ext(path) == ".xml" {
			continue
		}

		filtered = append(filtered, path)
	}

	return filtered
}

func ReadFiles(dir string, moduleFilter []string, file string) ([]SkuidPullResponse, error) {

	var files []string

	if file != "" {
		var err error
		files, err = FilterByGlob(file)

		if err != nil {
			return nil, err
		}

	} else {
		var err error
		if _, err := os.Stat(dir); err != nil {
			return nil, err
		}

		files, err = FilterByModule(dir, moduleFilter)

		if err != nil {
			return nil, err
		}
	}

	files = filterOutXmlFiles(files)

	definitions := []SkuidPullResponse{}

	for _, file := range files {

		metadataFilename := file

		bodyFilename := strings.Replace(file, ".json", ".xml", 1)
		//read the metadata file
		metadataFile, _ := ioutil.ReadFile(metadataFilename)
		//read the page xml
		bodyFile, _ := ioutil.ReadFile(bodyFilename)

		pullRes := &SkuidPullResponse{}

		_ = json.Unmarshal(metadataFile, pullRes)

		pullRes.Body = string(bodyFile)

		definitions = append(definitions, *pullRes)
	}

	return definitions, nil
}
