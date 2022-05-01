package nlx

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"
)

// GetAccessToken
func GetAccessToken(host, username, password string) (accessToken string, err error) {
	// prep the body
	body := []byte(url.Values{
		"grant_type": []string{"password"},
		"username":   []string{username},
		"password":   []string{password},
	}.Encode())

	type AccessTokenResponse struct {
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
		host+"/api/v2/auth/token",
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
