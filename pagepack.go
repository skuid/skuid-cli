package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var pagePackModule string
var outputFile string

// page-packCmd represents the page-pack command
var pagePackCmd = &cobra.Command{
	Use:   "page-pack",
	Short: "Retrieve a collection of Skuid Pages as a Page Pack.",
	Long:  "Retrieves all Pages in a given Module and returns in a Page Pack, which is a JSON-serialized array of Page objects.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// login to the Force API
		api, err := SalesforceLogin(
			ArgAppClientID,
			ArgAppClientSecret,
			ArgHost,
			ArgUsername,
			ArgPassword,
			ArgApiVersion,
		)

		if err != nil {
			Println(err.Error())
			return err
		}

		query := url.Values{
			"module": []string{pagePackModule},
			"as":     []string{"pagePack"},
		}

		result, err := api.Connection.Get("/skuid/api/v1/pages", query)

		if err != nil {
			Println(err.Error())
			return err
		}

		unquoted, _ := strconv.Unquote(string(result))

		pagePack := &PagePackResponse{}

		err = json.Unmarshal([]byte(unquoted), pagePack)

		if err != nil {
			Println(err.Error())
			return err
		}

		err = pagePack.WritePagePack(outputFile, pagePackModule)

		return err
	},
}

func init() {
	RootCmd.AddCommand(pagePackCmd)

	pagePackCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Filename of output Page Pack file")
}

type MasterPage struct {
	Attributes map[string]string `json:"attributes"`
	Name       string            `json:"Name"`
	UniqueID   string            `json:"skuid__UniqueId__c"`
}

type Page struct {
	ID                 string            `json:"Id"`
	Attributes         map[string]string `json:"attributes"`
	Name               string            `json:"Name"`
	Type               string            `json:"skuid__Type__c"`
	UniqueID           string            `json:"skuid__UniqueId__c"`
	Module             string            `json:"skuid__Module__c"`
	ComposerSettings   *string           `json:"skuid__Composer_Settings__c"`
	MasterPageID       *string           `json:"skuid__MasterPage__c"`
	MasterPageRelation *MasterPage       `json:"skuid__MasterPage__r"`
	IsMaster           bool              `json:"skuid__IsMaster__c"`
	Layout             *string           `json:"skuid__Layout__c"`
	Layout2            *string           `json:"skuid__Layout2__c"`
	Layout3            *string           `json:"skuid__Layout3__c"`
	Layout4            *string           `json:"skuid__Layout4__c"`
	Layout5            *string           `json:"skuid__Layout5__c"`
}

type PagePackResponse map[string][]Page

func (pack PagePackResponse) WritePagePack(filename, module string) error {
	definition := pack[module]

	str, err := json.MarshalIndent(definition, "", "    ")

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, []byte(str), 0644)
}
