package platform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/skuid/skuid/httperror"
	"github.com/skuid/skuid/text"
)

type OAuthResponse struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type JWTResponse struct {
	Token string `json:"token"`
}

type RestConnection struct {
	AccessToken          string
	APIVersion           string
	ClientId             string
	ClientSecret         string
	Host                 string
	Password             string
	Username             string
	JWT                  string
	MetadataServiceProxy *url.URL
	DataServiceProxy     *url.URL
}

type RestApi struct {
	Connection *RestConnection
}

// Login logs a given user into a target Skuid Platform site and returns a RestApi connection
// that can be used to make HTTP requests
func Login(host, username, password, apiVersion, metadataServiceProxy, dataServiceProxy string, verbose bool) (api *RestApi, err error) {

	if apiVersion == "" {
		apiVersion = "2"
	}

	loginStart := time.Now()

	if verbose {
		fmt.Println(fmt.Sprintf("Logging in to Skuid Platform as user: %v\n%v", username, host)) 
		fmt.Println("API Version: " + apiVersion)
	}

	conn := RestConnection{
		Host:       host,
		Username:   username,
		Password:   password,
		APIVersion: apiVersion,
	}

	if metadataServiceProxy != "" {
		proxyURL, err := url.Parse(metadataServiceProxy)
		if err != nil {
			return nil, err
		}
		conn.MetadataServiceProxy = proxyURL
	}

	if dataServiceProxy != "" {
		proxyURL, err := url.Parse(dataServiceProxy)
		if err != nil {
			return nil, err
		}
		conn.DataServiceProxy = proxyURL
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
		fmt.Println(text.SuccessWithTime("Login Success", loginStart))
		fmt.Println("Access Token: " + conn.AccessToken)
	}

	return api, nil
}

// Refresh is used to obtain an OAuth2 access_token
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

	resp, err := getClientForProxyURL(conn.MetadataServiceProxy).Do(req)

	if err != nil {
		switch err.(type) {
		case *url.Error:
			return text.AugmentError("Could not connect to authorization endpoint", err)
		default:
			return err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return httperror.New(resp.Status, resp.Body)
	}

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

	if err := json.NewDecoder(*jwtResult).Decode(&result); err != nil {
		return err
	}

	conn.JWT = result.Token

	return nil
}

// MakeRequest Executes an HTTP request using a session token
func (conn *RestConnection) MakeRequest(method string, url string, payload io.Reader, contentType string) (result *io.ReadCloser, err error) {
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

	resp, err := getClientForProxyURL(conn.MetadataServiceProxy).Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, httperror.New(resp.Status, resp.Body)
	}

	return &resp.Body, nil
}

// MakeJWTRequest Executes HTTP request using a jwt
func (conn *RestConnection) MakeJWTRequest(method string, url string, payload io.Reader, contentType string) (result *io.ReadCloser, err error) {
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

	// Send the public key endpoint so that warden can configure a JWT key if needed
	req.Header.Add("x-skuid-public-key-endpoint", conn.Host+"/api/v1/site/verificationkey")

	resp, err := getClientForProxyURL(conn.DataServiceProxy).Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, httperror.New(resp.Status, resp.Body)
	}

	return &resp.Body, nil
}

func getClientForProxyURL(url *url.URL) *http.Client {
	if url != nil {
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(url),
			},
		}
	}
	return http.DefaultClient
}
