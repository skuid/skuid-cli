package force

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Connection struct {
	AccessToken    string
	InstanceUrl    string
	ConsumerKey    string
	ConsumerSecret string
	Username       string
	APIVersion     string
	Password       string
}

// Used to refresh OAuth2 access_token
func (conn *Connection) Refresh() (err error) {
	urlValues := url.Values{}

	urlValues.Set("grant_type", "password")
	urlValues.Set("client_id", conn.ConsumerKey)
	urlValues.Set("client_secret", conn.ConsumerSecret)
	urlValues.Set("username", conn.Username)
	urlValues.Set("password", conn.Password)

	resp, err := http.PostForm(conn.InstanceUrl+"/services/oauth2/token", urlValues)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	result := OAuthResponse{}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return err
	}

	conn.AccessToken = result.AccessToken
	conn.InstanceUrl = result.InstanceUrl

	return nil
}

// Get is executing a HTTP request using GET verb
func (conn *Connection) Get(url string) (result []byte, err error) {
	return conn.Query("GET", url)
}

// Query is executing a HTTP request
func (conn *Connection) Query(method string, url string) (result []byte, err error) {

	endpoint := conn.InstanceUrl + "/services/data/v" + conn.APIVersion + url

	req, err := http.NewRequest(method, endpoint, nil)

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
