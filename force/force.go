package force

import (
	"bytes"
	"io"

	"encoding/json"

	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/jpmonette/goforce"
)

type RestConnection struct {
	force.Connection
}

type RestApi struct {
	Connection *RestConnection
}

// Login is logging-in the User to the Salesforce organization
func Login(consumerKey string, consumerSecret string, instanceUrl string, username string, password string, apiVersion string) (api *RestApi, err error) {

	if apiVersion == "" {
		apiVersion = "39.0"
	}

	conn := RestConnection{force.Connection{
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

	api = &RestApi{
		Connection: &conn,
	}

	return api, nil
}

func (conn *RestConnection) Get(url string, query url.Values) (result []byte, err error) {
	return conn.MakeRequest("GET", url, nil, query)
}

func (conn *RestConnection) Post(url string, payload interface{}) (result []byte, err error) {
	return conn.MakeRequest("POST", url, payload, nil)
}

// Executes an HTTP Request against the Skuid Pages REST API
func (conn *RestConnection) MakeRequest(method string, url string, payload interface{}, params url.Values) (result []byte, err error) {

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
