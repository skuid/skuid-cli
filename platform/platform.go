package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/skuid/skuid-cli/httperror"
	"github.com/skuid/skuid-cli/text"
	"github.com/skuid/skuid-cli/types"
	"github.com/skuid/skuid-cli/ziputils"
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
	req.Header.Add("User-Agent", "Skuid-CLI/0.3")

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
		return httperror.New(resp.Status, resp.Body)
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

	req.Header.Add("Authorization", "Bearer " + conn.AccessToken)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	req.Header.Add("User-Agent", "Skuid-CLI/0.3")

	resp, err := getClientForProxyURL(conn.MetadataServiceProxy).Do(req)

	if err != nil {
		return nil, err
	}

	// Attempt to seamlessly grab a new access token if our current token has expired
	if resp.StatusCode == 401 {
		defer resp.Body.Close()
		conn.AccessToken = ""
		refreshErr := conn.Refresh()
		if refreshErr == nil && conn.AccessToken != "" {
			newResp, newErr := getClientForProxyURL(conn.MetadataServiceProxy).Do(req)
			if newErr != nil {
				return nil, newErr
			}
			if newResp.StatusCode != 200 {
				defer newResp.Body.Close()
				return nil, httperror.New(newResp.Status, newResp.Body)
			}
			return &newResp.Body, nil
		}
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
	req.Header.Add("User-Agent", "Skuid-CLI/0.3")

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

// GetDeployPlan fetches a deploymnent plan from Skuid Platform API
func (api *RestApi) GetDeployPlan(payload io.Reader, verbose bool) (map[string]types.Plan, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Getting Deploy Plan"))
	}

	planStart := time.Now()
	// Get a deploy plan
	planResult, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/metadata/deploy/plan",
		payload,
		"application/zip",
	)

	if err != nil {
		return nil, err
	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success Getting Deploy Plan", planStart))
	}

	defer (*planResult).Close()

	var plans map[string]types.Plan
	if err := json.NewDecoder(*planResult).Decode(&plans); err != nil {
		return nil, err
	}

	return plans, nil
}

// ExecuteDeployPlan executes a map of plan items in a deployment plan
func (api *RestApi) ExecuteDeployPlan(plans map[string]types.Plan, targetDir string, verbose bool) ([]*io.ReadCloser, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Executing Deploy Plan"))
	}
	planResults := []*io.ReadCloser{}
	for _, plan := range plans {
		planResult, err := api.ExecutePlanItem(plan, targetDir, verbose)
		if err != nil {
			return nil, err
		}
		planResults = append(planResults, planResult)
	}
	return planResults, nil
}

// ExecutePlanItem executes a particular item in a deployment plan
func (api *RestApi) ExecutePlanItem(plan types.Plan, targetDir string, verbose bool) (*io.ReadCloser, error) {
	// Create a buffer to write our archive to.
	var planResult *io.ReadCloser
	bufDeploy := new(bytes.Buffer)
	err := ziputils.Archive(targetDir, bufDeploy, &plan.Metadata)
	if err != nil {
		log.Print("Error creating deployment ZIP archive")
		log.Fatal(err)
	}

	deployStart := time.Now()

	if plan.Host == "" {
		if verbose {
			fmt.Println(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", plan.URL, plan.Type))
		}
		planResult, err = api.Connection.MakeRequest(
			http.MethodPost,
			plan.URL,
			bufDeploy,
			"application/zip",
		)
		if err != nil {
			return nil, err
		}
	} else {

		url := fmt.Sprintf("%s:%s/api/v2%s", plan.Host, plan.Port, plan.URL)
		if verbose {
			fmt.Println(fmt.Sprintf("Making Deploy Request: URL: [%s] Type: [%s]", url, plan.Type))
		}
		planResult, err = api.Connection.MakeJWTRequest(
			http.MethodPost,
			url,
			bufDeploy,
			"application/zip",
		)
		if err != nil {
			return nil, err
		}

	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success Deploying to Source", deployStart))
	}
	defer (*planResult).Close()
	return planResult, nil

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
