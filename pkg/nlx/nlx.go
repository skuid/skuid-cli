package nlx

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/logging"
)

var (
	SkuidUserAgent = fmt.Sprintf("Skuid-CLI/%s", constants.VERSION_NAME)
)

const (
	DEFAULT_API_VERSION = "2"
)

func QuickGet[T any](
	route string,
	method string,
	body []byte,
	auth ...string,
	// optional
) (r T, err error) {
	logging.VerboseSeparator()
	logging.VerboseLn(color.Cyan.Sprintf("Skuid NLX Request: %v", route))

	// prepare the http request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	// Need to acquire and release resources from fasthttp
	// fasthttp handles multithreading for us and removes
	// allocs during processes. For instance, this login operation
	// takes about 44 allocs. The http standard lib requires about 221 allocs.

	// Determine the URI
	uri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(uri)
	// by passing `nil` as the first argument, the whole url will be parsed
	// by the uri.parse function.
	if err = uri.Parse(nil, []byte(route)); err != nil {
		return
	}
	req.SetURI(uri)

	// set the body
	// For login, this is the url encoded bytes that we prepared
	// with login credentials and information
	if len(body) > 0 {
		logging.VerboseF("With body: %v\n", string(body))
		req.SetBody(body)
	}

	// prep the request headers
	req.Header.SetMethod(method)
	req.Header.Add(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")
	req.Header.Add(fasthttp.HeaderUserAgent, SkuidUserAgent)

	// if it has authorization, put that as the header
	if len(auth) > 0 {
		req.Header.Add(fasthttp.HeaderAuthorization, fmt.Sprintf("Bearer %v", auth[0]))
	}

	// prep for the response
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// perform the request. errors only pop up if there's an issue
	// with assembly/resources.
	logging.VerboseLn("Login Attempt...")
	if err = fasthttp.Do(req, resp); err != nil {
		return
	}

	// check the validity of the body and grab the access token
	responseBody := resp.Body()
	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		err = fmt.Errorf("%s:\nStatus Code: %v\nBody: %v\n",
			color.Red.Sprint("ERROR"),
			color.Yellow.Sprint(statusCode),
			color.Cyan.Sprint(string(responseBody)),
		)
		return
	}

	logging.VerboseLn("Successful Request")

	if err = json.Unmarshal(responseBody, &r); err != nil {
		return
	}

	logging.VerboseF("(Response: %v)\n", r)

	return
}

func GetAccessToken(host, username, password string) (accessToken string, err error) {
	// prep the body
	urlValues := url.Values{}
	urlValues.Set("grant_type", "password")
	urlValues.Set("username", username)
	urlValues.Set("password", password)

	body := []byte(urlValues.Encode())

	type AccessTokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	var resp AccessTokenResponse
	if resp, err = QuickGet[AccessTokenResponse](
		host+"/auth/oauth/token",
		fasthttp.MethodPost,
		body,
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
	if resp, err = QuickGet[AuthorizationTokenResponse](
		host+"/api/v2/auth/token",
		fasthttp.MethodGet,
		[]byte{},
		accessToken,
	); err != nil {
		return
	}

	authToken = resp.AuthorizationToken

	return
}
