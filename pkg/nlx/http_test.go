package nlx_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
)

const (
	DEFAULT_API_VERSION = "2"
)

func GetAccessToken(
	host, username, password, metadataServiceProxy string, // optional
) (accessToken string, err error) {
	logging.VerboseSeparator()
	logging.VerboseLn("Starting NLX Login")

	// if apiVersion == "" {
	// 	logging.VerboseF("Setting API Version to DEFAULT_API_VERSION: %v\n", DEFAULT_API_VERSION)
	// 	apiVersion = DEFAULT_API_VERSION
	// }

	metadataServiceProxyUri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(metadataServiceProxyUri)
	if metadataServiceProxy != "" {
		if err = metadataServiceProxyUri.Parse(nil, []byte(metadataServiceProxy)); err == nil {
			logging.VerboseF("Using Metadata Service Proxy: %v", color.Green.Sprint(metadataServiceProxy))
		}
	}

	// Need to acquire and release resources from fasthttp
	// fasthttp handles multithreading for us and removes
	// allocs during processes. For instance, this login operation
	// takes about 44 allocs. The http standard lib requires about 221 allocs.

	// Determine the URI
	uri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(uri)

	// by passing `nil` as the first argument, the whole url will be parsed
	// by the uri.parse function.
	if err = uri.Parse(nil, []byte(host+"/auth/oauth/token")); err != nil {
		return
	}

	// Let's get the url values we're going to encode in the body
	urlValues := url.Values{}
	urlValues.Set("grant_type", "password")
	urlValues.Set("username", username)
	urlValues.Set("password", password)

	// prepare the http request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// set the uri to the one we calculated earlier
	req.SetURI(uri)

	// set the body to the url encoded bytes that we prepared
	// with login credentials and information
	req.SetBody([]byte(urlValues.Encode()))

	// prep the request itself
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Add(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")
	req.Header.Add(fasthttp.HeaderUserAgent, fmt.Sprintf("Skuid-CLI/%s", constants.VERSION_NAME))

	// prep for the response
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// TODO: Log that we're attempting to login
	logging.VerboseLn("Login Attempt...")

	// perform the request. errors only pop up if there's an issue
	// with assembly/resources.

	if err = fasthttp.Do(req, resp); err != nil {
		return
	}

	// check the validity of the body and grab the access token
	body := resp.Body()
	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		err = fmt.Errorf("%s:\nStatus Code: %v\nBody: %v\n",
			color.Red.Sprint("ERROR"),
			color.Yellow.Sprint(statusCode),
			color.Cyan.Sprint(string(body)),
		)
		return
	}

	logging.VerboseLn("Successfully logged in.")

	// anonymous struct for the response body unmarshalling from
	// json. the field we want is "access_token" from the response
	// body
	responseJson := struct {
		AccessToken string `json:"access_token"`
	}{}

	if err = json.Unmarshal(body, &responseJson); err != nil {
		return
	}

	accessToken = responseJson.AccessToken

	logging.VerboseLn(color.Green.Sprintf("Received Access Token: %v", accessToken))

	return
}

func TestFasthttpMethods(t *testing.T) {
	logging.SetVerbose(true)
	if _, err := GetAccessToken("https://jredhoop-subdomain.pliny.webserver:3000", "jredhoop", "SkuidLocalDevelopment", ""); err != nil {
		color.Red.Println(err)
		t.FailNow()
	}
}

func BenchmarkFasthttpLogin(b *testing.B) {
	logging.SetVerbose(false)
	for i := 0; i < b.N; i++ {
		if _, err := GetAccessToken("https://jredhoop-subdomain.pliny.webserver:3000", "jredhoop", "SkuidLocalDevelopment", ""); err != nil {
			color.Red.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkHttpLogin(b *testing.B) {
	logging.SetVerbose(false)
	for i := 0; i < b.N; i++ {
		if _, err := pkg.SkuidNlxLogin2("https://jredhoop-subdomain.pliny.webserver:3000", "jredhoop", "SkuidLocalDevelopment", "", "", ""); err != nil {
			color.Red.Println(err)
			b.FailNow()
		}
	}
}
