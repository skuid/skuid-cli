package platform

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/skuid/skuid/httperror"
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
func Login(host string, username string, password string, apiVersion string, verbose bool) (api *RestApi, err error) {

	if apiVersion == "" {
		apiVersion = "1"
	}

	if verbose {
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

	if verbose {
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

	req, err := http.NewRequest("POST", conn.Host+"/auth/oauth/token", strings.NewReader(urlValues.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Skuid-CLI/0.2")

	client := &http.Client{}
	resp, err := client.Do(req)

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
func (conn *RestConnection) MakeRequest(method string, url string, payload io.Reader, contentType string) (result []byte, err error) {

	endpoint := fmt.Sprintf("%s/api/v%s%s", conn.Host, conn.APIVersion, url)

	req, err := http.NewRequest(method, endpoint, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("User-Agent", "Skuid-CLI/0.2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, httperror.New(resp.Status, string(body))
	}

	return body, nil

}
