package pkg

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/skuid/skuid-cli/pkg/flags"
)

type Authorization struct {
	Host               string
	AccessToken        string
	AuthorizationToken string
}

type AuthorizeOptions struct {
	Host     string
	Username string
	Password *flags.Value[string]
}

func GetAccessToken(opts *AuthorizeOptions) (accessToken string, err error) {
	// prep the body
	body := []byte(url.Values{
		"grant_type": []string{"password"},
		"username":   []string{opts.Username},
		"password":   []string{opts.Password.Unredacted().String()},
	}.Encode())

	type AccessTokenResponse struct {
		ExpiresIn   int    `json:"expires_in"`
		AccessToken string `json:"access_token"`
	}

	var resp AccessTokenResponse
	if resp, err = JsonBodyRequest[AccessTokenResponse](
		opts.Host+"/auth/oauth/token",
		http.MethodPost,
		body,
		map[string]string{
			HeaderContentType: URL_ENCODED_CONTENT_TYPE,
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
	if resp, err = JsonBodyRequest[AuthorizationTokenResponse](
		fmt.Sprintf("%v/api/%v/auth/token", host, DEFAULT_API_VERSION),
		http.MethodGet,
		[]byte{},
		map[string]string{
			HeaderAuthorization: fmt.Sprintf("Bearer %v", string(accessToken)),
		},
	); err != nil {
		return
	}

	authToken = resp.AuthorizationToken

	return
}

func Authorize(opts *AuthorizeOptions) (info *Authorization, err error) {
	info = &Authorization{
		Host: opts.Host,
	}

	if info.AccessToken, err = GetAccessToken(opts); err != nil {
		return
	}

	info.AuthorizationToken, err = GetAuthorizationToken(info.Host, info.AccessToken)

	return
}
