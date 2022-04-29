package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/skuid/tides/pkg/constants"
)

type NlxConnection struct {
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

type NlxAuthResponse struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

var (
	SkuidUserAgent = fmt.Sprintf("Skuid-CLI/%s", constants.VERSION_NAME)
)

// Refresh is used to obtain an OAuth2 access_token
func (conn *NlxConnection) Refresh() error {
	urlValues := url.Values{}

	urlValues.Set("grant_type", "password")
	urlValues.Set("username", conn.Username)
	urlValues.Set("password", conn.Password)

	req, err := http.NewRequest("POST", conn.Host+"/auth/oauth/token", strings.NewReader(urlValues.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", SkuidUserAgent)

	resp, err := getClientForProxyURL(conn.MetadataServiceProxy).Do(req)

	if err != nil {
		switch err.(type) {
		case *url.Error:
			return fmt.Errorf("%v\n%v", "Could not connect to Authorization Endpoint", err.Error())
		default:
			return err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NewHttpError(resp.Status, resp.Body)
	}

	result := NlxAuthResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return NewHttpError(resp.Status, resp.Body)
	}

	conn.AccessToken = result.AccessToken

	return nil
}

type JWTResponse struct {
	Token string `json:"token"`
}

func (conn *NlxConnection) GetJWT() error {
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
func (conn *NlxConnection) MakeRequest(method string, url string, payload io.Reader, contentType string) (result *io.ReadCloser, err error) {
	endpoint := fmt.Sprintf("%s/api/v%s%s", conn.Host, conn.APIVersion, url)

	req, err := http.NewRequest(method, endpoint, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.AccessToken)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	req.Header.Add("User-Agent", SkuidUserAgent)

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
				return nil, NewHttpError(newResp.Status, newResp.Body)
			}
			return &newResp.Body, nil
		}
	}

	if resp.StatusCode == 204 { // No content
		return nil, nil
	}

	if resp.StatusCode == 504 || resp.StatusCode == 503 {
		return nil, errors.Errorf("%v - %s", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, NewHttpError(resp.Status, resp.Body)
	}

	return &resp.Body, nil
}

// MakeJWTRequest Executes HTTP request using a jwt
func (conn *NlxConnection) MakeJWTRequest(method string, url string, payload io.Reader, contentType string) (result *io.ReadCloser, err error) {
	endpoint := fmt.Sprintf("https://%s", url)
	req, err := http.NewRequest(method, endpoint, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+conn.JWT)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	req.Header.Add("User-Agent", SkuidUserAgent)

	// Send the public key endpoint so that warden can configure a JWT key if needed
	req.Header.Add("x-skuid-public-key-endpoint", conn.Host+"/api/v1/site/verificationkey")

	resp, err := getClientForProxyURL(conn.DataServiceProxy).Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, NewHttpError(resp.Status, resp.Body)
	}

	return &resp.Body, nil
}
