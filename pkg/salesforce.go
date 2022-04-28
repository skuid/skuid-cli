package pkg

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	salesforce "github.com/jpmonette/goforce"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/flags"
)

type SalesforceRestConnection struct {
	salesforce.Connection
}

type SalesforceRestApi struct {
	Connection *SalesforceRestConnection
}

// SalesforceLogin is logging-in the User to the Salesforce organization
func SalesforceLogin(cmd *cobra.Command) (api *SalesforceRestApi, err error) {
	var consumerKey, consumerSecret, instanceUrl, username, password, apiVersion string

	// TODO: Rename ClientId to "Consumer Key" or relate that to the Salesforce term for this argument.
	if consumerKey, err = cmd.Flags().GetString(flags.ClientId.Name); err != nil {
		return
	}

	// TODO: Rename Client Secret to "Consumer Secret" or relate that to the Salesforce term for this argument.
	if consumerSecret, err = cmd.Flags().GetString(flags.ClientSecret.Name); err != nil {
		return
	}

	// TODO: Rename Host to "Instance Url" or relate that to the Salesforce term for this argument.
	if instanceUrl, err = cmd.Flags().GetString(flags.Host.Name); err != nil {
		return
	}

	if username, err = cmd.Flags().GetString(flags.Username.Name); err != nil {
		return
	}

	if password, err = cmd.Flags().GetString(flags.Password.Name); err != nil {
		return
	}

	if apiVersion, err = cmd.Flags().GetString(flags.ApiVersion.Name); err != nil {
		return
	}

	if apiVersion == "" {
		apiVersion = constants.DEFAULT_SALESFORCE_API_VERSION
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
