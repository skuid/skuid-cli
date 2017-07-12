package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type OAuthResponse struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type RestConnection struct {
	AccessToken  string
	APIVersion   string
	ClientId     string
	ClientSecret string
	Host         string
	Password     string
	Username     string
}

type RestApi struct {
	Connection *RestConnection
}

// Logs a given user into a target Skuid Platform site and returns a RestApi connection
// that can be used to make HTTP requests
func Login(host string, username string, password string, apiVersion string, verboseLogs bool) (api *RestApi, err error) {

	if apiVersion == "" {
		apiVersion = "1"
	}

	if verboseLogs {
		fmt.Println("Logging in to Skuid Platform... ")
		fmt.Println("Host: " + host)
		fmt.Println("Username: " + username)
		fmt.Println("Password: " + password)
		fmt.Println("API Version: " + apiVersion)
	}

	conn := RestConnection{
		Host:       host,
		Username:   username,
		Password:   password,
		APIVersion: apiVersion,
	}

	err = conn.Refresh()

	if err != nil {
		return nil, err
	}

	api = &RestApi{
		Connection: &conn,
	}

	if verboseLogs {
		fmt.Println("Login successful! Access Token: " + conn.AccessToken)
	}

	return api, nil
}

// Used to obtain an OAuth2 access_token
func (conn *RestConnection) Refresh() error {
	urlValues := url.Values{}

	urlValues.Set("grant_type", "password")
	urlValues.Set("username", conn.Username)
	urlValues.Set("password", conn.Password)

	resp, err := http.PostForm(conn.Host+"/auth/oauth/token", urlValues)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	result := OAuthResponse{}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	conn.AccessToken = result.AccessToken

	return nil
}

// Executes an HTTP request
func (conn *RestConnection) MakeRequest(method string, url string, payload interface{}) (result []byte, err error) {

	endpoint := conn.Host + "/api/v" + conn.APIVersion + url

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
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)

}
