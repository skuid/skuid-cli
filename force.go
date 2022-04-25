package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	salesforce "github.com/jpmonette/goforce"
)

type SalesforceRestConnection struct {
	salesforce.Connection
}

type SalesforceRestApi struct {
	Connection *SalesforceRestConnection
}

// SalesforceLogin is logging-in the User to the Salesforce organization
func SalesforceLogin(consumerKey string, consumerSecret string, instanceUrl string, username string, password string, apiVersion string) (api *SalesforceRestApi, err error) {

	if apiVersion == "" {
		apiVersion = "39.0"
	}

	conn := SalesforceRestConnection{salesforce.Connection{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		InstanceUrl:    instanceUrl,
		Username:       username,
		Password:       password,
		APIVersion:     apiVersion,
	}}

	err = conn.Refresh()

	if err != nil {
		return nil, err
	}

	api = &SalesforceRestApi{
		Connection: &conn,
	}

	return api, nil
}

func (conn *SalesforceRestConnection) Get(url string, query url.Values) (result []byte, err error) {
	return conn.MakeForceRequest("GET", url, nil, query)
}

func (conn *SalesforceRestConnection) Post(url string, payload interface{}) (result []byte, err error) {
	return conn.MakeForceRequest("POST", url, payload, nil)
}

// Executes an HTTP Request against the Skuid Pages REST API
func (conn *SalesforceRestConnection) MakeForceRequest(method string, url string, payload interface{}, params url.Values) (result []byte, err error) {

	endpoint := conn.InstanceUrl + "/services/apexrest" + url

	if params != nil && len(params) != 0 {
		endpoint += "?" + params.Encode()
	}

	var body io.Reader

	if payload != nil {
		jsonBytes, err := json.Marshal(payload)

		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(jsonBytes)

	}

	req, err := http.NewRequest(method, endpoint, body)

	if err != nil {
		return result, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return result, err
	}

	return result, nil
}
