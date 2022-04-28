package pkg

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
)

// Page Pack Command Arguments
var (
	// page-packCmd represents the page-pack command
	pagePackCmd = &cobra.Command{
		Use:   "page-pack",
		Short: "Retrieve a collection of Skuid Pages as a Page Pack.",
		Long:  "Retrieves all Pages in a given Module and returns in a Page Pack, which is a JSON-serialized array of Page objects.",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			// login to the Force API
			api, err := SalesforceLogin(cmd)

			if err != nil {
				return err
			}

			var module string
			if module, err = cmd.Flags().GetString(flags.Module.Name); err != nil {
				return err
			}

			query := url.Values{
				"module": []string{module}, // this was an unused string before
				"as":     []string{"pagePack"},
			}

			result, err := api.Connection.Get("/skuid/api/v1/pages", query)

			if err != nil {
				return err
			}

			unquoted, _ := strconv.Unquote(string(result))

			pagePack := PagePackResponse{}

			err = json.Unmarshal([]byte(unquoted), &pagePack)

			if err != nil {
				return err
			}

			var outputfileName string
			if outputfileName, err = cmd.Flags().GetString(flags.OutputFile.Name); err != nil {
				return
			}

			err = WritePagePack(pagePack, outputfileName, module)

			return
		},
	}
)

func init() {
	RootCmd.AddCommand(pagePackCmd)

	flags.AddFlagFunctions(pagePackCmd, flags.SalesforceLoginFlags...)
	flags.AddFlags(pagePackCmd, flags.OutputFile)
}

type PagePackResponse map[string][]struct {
	ID                 string            `json:"Id"`
	Attributes         map[string]string `json:"attributes"`
	Name               string            `json:"Name"`
	Type               string            `json:"skuid__Type__c"`
	UniqueID           string            `json:"skuid__UniqueId__c"`
	Module             string            `json:"skuid__Module__c"`
	ComposerSettings   *string           `json:"skuid__Composer_Settings__c"`
	MasterPageID       *string           `json:"skuid__MasterPage__c"`
	MasterPageRelation *struct {
		Attributes map[string]string `json:"attributes"`
		Name       string            `json:"Name"`
		UniqueID   string            `json:"skuid__UniqueId__c"`
	} `json:"skuid__MasterPage__r"`
	IsMaster bool    `json:"skuid__IsMaster__c"`
	Layout   *string `json:"skuid__Layout__c"`
	Layout2  *string `json:"skuid__Layout2__c"`
	Layout3  *string `json:"skuid__Layout3__c"`
	Layout4  *string `json:"skuid__Layout4__c"`
	Layout5  *string `json:"skuid__Layout5__c"`
}

func WritePagePack(pack PagePackResponse, filename string, module string) error {
	definition := pack[module]
	if str, err := json.MarshalIndent(definition, "", "    "); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filename, []byte(str), 0644)
	}
}
