package pkg

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
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

func GetAccessToken(opts *AuthorizeOptions) (string, error) {
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

	if resp, err := JsonBodyRequest[AccessTokenResponse](
		opts.Host+"/auth/oauth/token",
		http.MethodPost,
		body,
		map[string]string{
			HeaderContentType: URL_ENCODED_CONTENT_TYPE,
		},
	); err != nil {
		return "", err
	} else {
		return resp.AccessToken, nil
	}
}

func GetAuthorizationToken(host, accessToken string) (string, error) {
	type AuthorizationTokenResponse struct {
		AuthorizationToken string `json:"token"`
	}

	if resp, err := JsonBodyRequest[AuthorizationTokenResponse](
		fmt.Sprintf("%v/api/%v/auth/token", host, DEFAULT_API_VERSION),
		http.MethodGet,
		[]byte{},
		map[string]string{
			HeaderAuthorization: fmt.Sprintf("Bearer %v", string(accessToken)),
		},
	); err != nil {
		return "", err
	} else {
		return resp.AuthorizationToken, nil
	}
}

func AuthorizeOnce(opts *AuthorizeOptions) (*Authorization, error) {
	auth, err := Authorize(opts)
	// we don't need it anymore - very inelegant approach but at least it is something for now
	// Clearing it here instead of in auth package which is the only place its accessed because the tests that exist
	// for auth rely on package global variables so clearing in there would break those tests as they currently exist.
	//
	// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
	// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
	// intentionally ignoring error since there is nothing we can do and we should fail entirely as a result
	_ = opts.Password.Set("")
	return auth, err
}

func Authorize(opts *AuthorizeOptions) (auth *Authorization, err error) {
	message := "Requesting authentication"
	fields := logging.Fields{
		"host":     opts.Host,
		"username": opts.Username,
	}
	logger := logging.WithTracking("pkg.Authorize", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	auth = &Authorization{Host: opts.Host}
	if auth.AccessToken, err = GetAccessToken(opts); err != nil {
		return nil, err
	}
	if auth.AuthorizationToken, err = GetAuthorizationToken(auth.Host, auth.AccessToken); err != nil {
		return nil, err
	}

	logger = logger.WithSuccess()
	return auth, nil
}
