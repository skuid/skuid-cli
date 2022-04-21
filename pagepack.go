package main

import (
	"encoding/json"
	"io/ioutil"
)

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
