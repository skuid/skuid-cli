package pkg

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/errors"
	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/util"
)

var (
	SkuidUserAgent = fmt.Sprintf("Tides/%s", constants.VERSION_NAME)

	AcceptableProtocols = []string{
		"http", "https",
	}
)

const (
	DEFAULT_API_VERSION        = "v2"
	MAX_AUTHORIZATION_ATTEMPTS = 3 // 5 will lock you out iirc
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

	logging.Logger.Debug("Unmarshalling data")

	if err = json.Unmarshal(responseBody, &r); err != nil {
		return
	}

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

	logging.Logger.Debug("Assembling Request.")

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
	logging.Logger.Debugf("URI: %v", uri.String())

	// Set the body for the request (if found)
	// (empty bodies will be discarded)
	//
	// ...For login: this is the url encoded bytes that we prepared beforehand
	// ...along with the grant_type: password
	if len(body) > 0 {
		if len(body) < 1e4 {
			logging.Logger.Debugf("With body: %v\n", string(util.RemovePasswordBytes(body)))
		} else {
			logging.Logger.Debugf("(With large body, too large to print)\n")
		}
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
	logging.Logger.Debug("Making Request.")
	if err = fasthttp.Do(req, resp); err != nil {
		return
	}

	// check the validity of the body and grab the access token
	responseBody := resp.Body()
	statusCode := resp.StatusCode()

	httpError := func() error {
		return errors.Critical("%s:\nStatus Code: %v\nBody: %v\n",
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

	logging.Logger.Debug("Successful Request")

	response = responseBody

	// don't output something huge
	if len(response) < 1e3 {
		// Debug and output the marshalled body
		if isJson := resp.Header.Peek(fasthttp.HeaderContentType); strings.Contains(string(isJson), JSON_CONTENT_TYPE) {
			var prettyMarshal interface{}
			json.Unmarshal(response, &prettyMarshal)
			pretty, _ := json.MarshalIndent(prettyMarshal, "", " ")
			logging.Logger.Tracef("Pretty Response Body: %v", color.Cyan.Sprint(string(pretty)))
		}
	}

	return
}
