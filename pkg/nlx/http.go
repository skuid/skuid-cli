package nlx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/logging"
)

var (
	SkuidUserAgent = fmt.Sprintf("Tides/%s", constants.VERSION_NAME)

	AcceptableProtocols = []string{
		"http", "https",
	}
)

const (
	DEFAULT_API_VERSION        = "v2"
	MAX_AUTHORIZATION_ATTEMPTS = 3
)

// FastJsonBodyRequest takes a type parameter and automatically
// Unmarshals the body of the payload response into a value of that type.
//
// Leveraging the fasthttp lib, this should outperform the stdlib http package.
//
// Some patterns require acquiring and releasing resources from fasthttp.
//
// The package 'fasthttp' handles multithreading for us and removes
// allocs during processes.
//
// For instance, this login operation
// takes about 44 allocs. The http standard lib requires about 221 allocs.
func FastJsonBodyRequest[T any](
	route string,
	method string,
	body []byte,
	additionalHeaders map[string]string,
) (r T, err error) {
	var responseBody []byte
	if responseBody, err = FastRequest(route, method, body, additionalHeaders); err != nil {
		return
	}

	logging.VerboseLn("Unmarshalling data")

	if err = json.Unmarshal(responseBody, &r); err != nil {
		return
	}

	logging.VerboseF("(Response: %v)\n", r)

	return
}

func FastRequest(
	route string,
	method string,
	body []byte,
	headers RequestHeaders,
) (response []byte, err error) {
	return FastRequestHelper(route, method, body, headers, 0)
}

func FastRequestHelper(
	route string,
	method string,
	body []byte,
	headers RequestHeaders,
	attempts int,
) (response []byte, err error) {
	// only https
	if strings.HasPrefix(route, "http://") {
		route = strings.Replace(route, "http://", "https://", 1)
	}
	// only https
	if !strings.HasPrefix(route, "https://") {
		route = fmt.Sprintf("https://%v", route)
	}

	logging.VerboseSeparator()
	logging.VerboseLn(color.Gray.Sprint("HTTPS Request:"), color.Cyan.Sprint(route))

	// Prepare resources for the http request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Prepare resources for the URI
	uri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(uri)
	// by passing `nil` as the first argument, the whole url will be parsed
	// by the uri.parse function.
	if err = uri.Parse(nil, []byte(route)); err != nil {
		return
	}
	// Set the URI for the request
	req.SetURI(uri)

	// Set the body for the request (if found)
	// (empty bodies will be discarded)
	//
	// ...For login: this is the url encoded bytes that we prepared beforehand
	// ...along with the grant_type: password
	if len(body) > 0 {
		logging.VerboseF("With body: %v\n", string(body))
		req.SetBody(body)
	}

	// prep the request headers
	req.Header.SetMethod(method)
	req.Header.Add(fasthttp.HeaderUserAgent, SkuidUserAgent)
	if headers != nil {
		for headerName, headerValue := range headers {
			req.Header.Add(headerName, headerValue)
		}
	}

	// prep for the response
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// perform the request. errors only pop up if there's an issue
	// with assembly/resources.
	logging.VerboseLn("Making Request")
	if err = fasthttp.Do(req, resp); err != nil {
		return
	}

	// check the validity of the body and grab the access token
	responseBody := resp.Body()
	statusCode := resp.StatusCode()

	httpError := func() error {
		return fmt.Errorf("%s:\nStatus Code: %v\nBody: %v\n",
			color.Red.Sprint("ERROR"),
			color.Yellow.Sprint(statusCode),
			color.Cyan.Sprint(string(responseBody)),
		)
	}

	switch statusCode {
	case fasthttp.StatusUnauthorized:
		if attempts < MAX_AUTHORIZATION_ATTEMPTS {
			return FastRequestHelper(route, method, body, headers, attempts+1)
		} else {
			err = httpError()
		}
	case fasthttp.StatusOK:
		// we're good
	default:
		err = httpError()
	}

	if err != nil {
		return
	}

	logging.VerboseLn("Successful Request")

	response = responseBody

	// Debug and output the marshalled body
	if isJson := resp.Header.Peek(fasthttp.HeaderContentType); strings.Contains(string(isJson), JSON_CONTENT_TYPE) {
		var prettyMarshal interface{}
		json.Unmarshal(response, &prettyMarshal)
		pretty, _ := json.MarshalIndent(prettyMarshal, "", " ")
		logging.DebugF("Pretty Response Body: %v", color.Cyan.Sprint(string(pretty)))
	}

	return
}
