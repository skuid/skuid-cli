package nlx

import (
	"encoding/json"
	"fmt"

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
	logging.VerboseSeparator()
	logging.VerboseLn(color.Gray.Sprint("Skuid NLX Request:"), color.Cyan.Sprint(route))

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
	if additionalHeaders != nil {
		for headerName, headerValue := range additionalHeaders {
			req.Header.Add(headerName, headerValue)
		}
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
