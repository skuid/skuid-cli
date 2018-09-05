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

type JWTResponse struct {
	Token string `json:"token"`
}

type RestConnection struct {
	AccessToken  string
	APIVersion   string
	ClientId     string
	ClientSecret string
	Host         string
	Password     string
	Username     string
	JWT          string
}

type RestApi struct {
	Connection *RestConnection
}

// Logs a given user into a target Skuid Platform site and returns a RestApi connection
// that can be used to make HTTP requests
func Login(host string, username string, password string, apiVersion string, verbose bool) (api *RestApi, err error) {

	if apiVersion == "" {
		apiVersion = "2"
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

	err = conn.GetJWT()

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

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return httperror.New(resp.Status, string(body))
	}

	defer resp.Body.Close()

	result := OAuthResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	conn.AccessToken = result.AccessToken

	return nil
}

func (conn *RestConnection) GetJWT() error {
	jwtResult, err := conn.MakeRequest(
		http.MethodGet,
		"/auth/token",
		nil,
		"",
	)

	if err != nil {
		return err
	}

	result := JWTResponse{}

	if err := json.NewDecoder(jwtResult).Decode(&result); err != nil {
		return err
	}

	conn.JWT = result.Token

	return nil
}

// MakeRequest Executes an HTTP request using a session token
func (conn *RestConnection) MakeRequest(method string, url string, payload io.Reader, contentType string) (result io.ReadCloser, err error) {

	endpoint := fmt.Sprintf("%s/api/v%s%s", conn.Host, conn.APIVersion, url)

	req, err := http.NewRequest(method, endpoint, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.AccessToken)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	req.Header.Add("User-Agent", "Skuid-CLI/0.2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, httperror.New(resp.Status, string(body))
	}

	return resp.Body, nil
}

// MakeJWTRequest Executes HTTP request using a jwt
func (conn *RestConnection) MakeJWTRequest(method string, url string, payload io.Reader, contentType string) (result io.ReadCloser, err error) {
	fmt.Println("MAKING JWT REQUEST")
	fmt.Println(url)
	endpoint := fmt.Sprintf("https://%s", url)
	req, err := http.NewRequest(method, endpoint, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.JWT)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	req.Header.Add("User-Agent", "Skuid-CLI/0.2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, httperror.New(resp.Status, string(body))
	}

	return resp.Body, nil
	/*
		return ziputils.CreateTestZip([]ziputils.TestFile{
			{
				Name: "readme.txt",
				Body: "This archive contains some text files.",
			},
			{
				Name: "gopher.txt",
				Body: "Gopher names:\nGeorge\nGeoffrey\nGonzo",
			},
			{
				Name: "dataservices/BenLocalDataService.json",
				Body: "{\"anotherkey\":\"anothervalue\"}",
			},
		})
	*/
}
