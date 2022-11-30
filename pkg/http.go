package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gookit/color"

	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var VERSION_NAME string

var (
	SkuidUserAgent = fmt.Sprintf("skuid-cli/%s", VERSION_NAME)

	AcceptableProtocols = []string{
		"http", "https",
	}
)

const (
	DEFAULT_API_VERSION        = "v2"
	MAX_AUTHORIZATION_ATTEMPTS = 1 // 5 will lock you out
)

func JsonBodyRequest[T any](
	route string,
	method string,
	body []byte,
	additionalHeaders map[string]string,
) (r T, err error) {
	var responseBody []byte
	if responseBody, err = Request(route, method, body, additionalHeaders); err != nil {
		return
	}

	logging.Get().Trace("Unmarshalling data")

	if err = json.Unmarshal(responseBody, &r); err != nil {
		return
	}

	return
}

func Request(
	route string,
	method string,
	body []byte,
	headers RequestHeaders,
) (response []byte, err error) {
	return RequestHelper(route, method, body, headers, 0)
}

func RequestHelper(
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
	if !strings.HasPrefix(route, "https://") {
		route = fmt.Sprintf("https://%v", route)
	}

	req, err := http.NewRequest(method, route, bytes.NewReader(body))
	for header, value := range headers {
		req.Header.Add(header, value)
	}
	logging.Get().Tracef("URI: %v", color.Blue.Sprint(route))

	// prep the request headers
	req.Header.Set(HeaderUserAgent, SkuidUserAgent)

	// perform the request. errors only pop up if there's an issue with assembly/resources.
	logging.Get().Trace(color.Blue.Sprint("Making Request"))
	var resp *http.Response
	if resp, err = (&http.Client{}).Do(req); err != nil {
		return
	}

	// // check the validity of the body and grab the access token
	var responseBody []byte
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode

	httpError := func() error {
		return errors.Critical("%s:\nStatus Code: %v\nBody: %v\nURI: %v\nHeaders: %v",
			color.Red.Sprint("ERROR"),
			color.Yellow.Sprint(statusCode),
			color.Cyan.Sprint(string(responseBody)),
			color.Red.Sprint(route),
			color.Cyan.Sprint(headers),
		)
	}

	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		// we're good
	case http.StatusUnauthorized:
		if attempts < MAX_AUTHORIZATION_ATTEMPTS {
			return RequestHelper(route, method, body, headers, attempts+1)
		} else {
			err = httpError()
		}
	default:
		err = httpError()
	}

	if err != nil {
		return
	}

	logging.Get().Trace(color.Green.Sprint("Successful Request"))

	response = responseBody

	return
}
