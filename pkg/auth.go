package pkg

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"
)

type Authorization struct {
	username string // private
	password string // private

	Host               string
	AccessToken        string
	AuthorizationToken string
}

func (a *Authorization) Refresh() (err error) {
	access, err := GetAccessToken(a.Host, a.username, a.password)
	if err != nil {
		return
	}
	a.AccessToken = access

	auth, err := GetAuthorizationToken(a.Host, a.AccessToken)
	if err != nil {
		return
	}
	a.AuthorizationToken = auth

	return
}

// GetAccessToken
func GetAccessToken(host, username, password string) (accessToken string, err error) {
	// prep the body
	body := []byte(url.Values{
		"grant_type": []string{"password"},
		"username":   []string{username},
		"password":   []string{password},
	}.Encode())

	type AccessTokenResponse struct {
		ExpiresIn   int    `json:"expires_in"`
		AccessToken string `json:"access_token"`
	}

	var resp AccessTokenResponse
	if resp, err = FastJsonBodyRequest[AccessTokenResponse](
		host+"/auth/oauth/token",
		fasthttp.MethodPost,
		body,
		map[string]string{
			fasthttp.HeaderContentType: "application/x-www-form-urlencoded",
		},
	); err != nil {
		return
	}

	accessToken = resp.AccessToken

	return
}

func GetAuthorizationToken(host, accessToken string) (authToken string, err error) {
	type AuthorizationTokenResponse struct {
		AuthorizationToken string `json:"token"`
	}

	var resp AuthorizationTokenResponse
	if resp, err = FastJsonBodyRequest[AuthorizationTokenResponse](
		fmt.Sprintf("%v/api/%v/auth/token", host, DEFAULT_API_VERSION),
		fasthttp.MethodGet,
		[]byte{},
		map[string]string{
			fasthttp.HeaderAuthorization: fmt.Sprintf("Bearer %v", string(accessToken)),
		},
	); err != nil {
		return
	}

	authToken = resp.AuthorizationToken

	return
}

func Authorize(host, username, password string) (info *Authorization, err error) {
	info = &Authorization{
		Host:     host,
		username: username,
		password: password,
	}

	if info.AccessToken, err = GetAccessToken(host, username, password); err != nil {
		return
	}

	info.AuthorizationToken, err = GetAuthorizationToken(host, info.AccessToken)

	return
}
