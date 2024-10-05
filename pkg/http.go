package pkg

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/logging"
)

var (
	AcceptableProtocols = []string{
		"http", "https",
	}
)

const (
	DefaultApiVersion            = "v2"
	MaxAuthorizationAttempts     = 2 // 5 will lock you out
	RequestTotalAttemptsFieldKey = "totalAttempts"
)

func JsonBodyRequest[T any](
	route string,
	method string,
	body []byte,
	additionalHeaders map[string]string,
) (result *T, err error) {
	route = FixUrl(route)
	attempts := 0
	message := fmt.Sprintf("Executing JSON %v request to %v", method, route)
	fields := logging.Fields{
		"url":     route,
		"method":  method,
		"bodyLen": len(body),
	}
	logger := logging.WithTraceTracking("pkg.JsonBodyRequest", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	responseBody, err := requestHelper(route, method, body, additionalHeaders, &attempts, logger)
	logger = logger.WithFields(logging.Fields{"responseBodyLen": len(responseBody), RequestTotalAttemptsFieldKey: attempts})
	if err != nil {
		return nil, err
	}

	logger.Tracef("Unmarshalling data from %v request to %v", method, route)
	result = new(T)
	if err := json.Unmarshal(responseBody, result); err != nil {
		// Skuid Review Required - Do any of the APIs return confidential information besides the auth apis?
		//
		// TODO: For now, only logging body and `--diag` flag specified to minimize times when confidential information may be logged.
		// Once it is known whether or any other APIs can return confidential information, adjust to protect the confidential
		// information as appropriate.
		if logger.DiagEnabled() {
			logger.WithError(err).Debugf("Response from %v: %s", route, responseBody)
		}
		return nil, fmt.Errorf("unable to parse response from %v: %w", route, err)
	}

	logger = logger.WithSuccess()
	return result, nil
}

func Request(
	route string,
	method string,
	body []byte,
	headers RequestHeaders,
) (responseBody []byte, err error) {
	route = FixUrl(route)
	attempts := 0
	message := fmt.Sprintf("Executing %v request to %v", method, route)
	fields := logging.Fields{
		"url":     route,
		"method":  method,
		"bodyLen": len(body),
	}
	logger := logging.WithTraceTracking("pkg.Request", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	responseBody, err = requestHelper(route, method, body, headers, &attempts, logger)
	if err != nil {
		return nil, err
	}

	logger = logger.WithSuccess(logging.Fields{"responseBodyLen": len(responseBody), RequestTotalAttemptsFieldKey: attempts})
	return responseBody, nil
}

func FixUrl(route string) string {
	// only https
	if strings.HasPrefix(route, "http://") {
		return strings.Replace(route, "http://", "https://", 1)
	}
	if !strings.HasPrefix(route, "https://") {
		return fmt.Sprintf("https://%v", route)
	}

	return route
}

func requestHelper(
	route string,
	method string,
	body []byte,
	headers RequestHeaders,
	attempts *int,
	parentLogger *logging.Logger,
) (responseBody []byte, err error) {
	message := fmt.Sprintf("Making request to %v", route)
	*attempts++
	fields := logging.Fields{
		"url":        route,
		"method":     method,
		"bodyLen":    len(body),
		"attemptNum": *attempts,
	}
	logger := parentLogger.WithTraceTracking("requestHelper", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	req, err := createRequest(route, method, body, headers, logger)
	if err != nil {
		return nil, err
	}

	resp, err := sendRequest(req, logger)
	if err != nil {
		return nil, err
	}

	responseBody, err = readResponse(resp, logger)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		// we're good
	case http.StatusUnauthorized:
		if *attempts < MaxAuthorizationAttempts {
			logger.Tracef("Retrying request to %v (%v of %v)", route, (*attempts)+1, MaxAuthorizationAttempts)
			// intentionally passing parentLogger here since its not a child of the current logger
			return requestHelper(route, method, body, headers, attempts, parentLogger)
		} else {
			err = getError(resp, responseBody)
		}
	default:
		err = getError(resp, responseBody)
	}

	if err != nil {
		// Skuid Review Required - Do any of the APIs return confidential information in a failed response? What is the full
		// list of content types that can be returned from an API? Ideally, there would be logging of request & response bodies
		// (with conditional truncation based on current log level) but unsure of response payload and for sure the
		// auth APIs send/receive confidential info so those need to be separated out and/or protected in some way.  For now,
		// only logging response body upon failure when content type is JSON and requiring `--diag` flag to have been specified
		// to minimize times when confidential information may be logged and make it an intentional action by user.
		//
		// TODO: Once above is answered, implement logging request & response body (truncated based on some variable condition)
		// ensuring that any API that would contain confidential information in request or response is not leaked, at least not
		// without some type of protection around it (e.g., --diag flag).  In theory, this would only be auth APIs which could
		// be separated out completely.
		if logger.DiagEnabled() && isJSONResponse(resp) {
			logger.WithError(err).Debugf("Response from request to %v with status code %v: %s", route, resp.StatusCode, responseBody)
		}
		return nil, err
	}

	logger = logger.WithSuccess()
	return responseBody, nil
}

func createGzipRequest(route string, method string, body []byte, logger *logging.Logger) (req *http.Request, err error) {
	message := fmt.Sprintf("Creating gzip %v request to %v", method, route)
	logger = logger.WithTraceTracking("createGzipRequest", message).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(body); err != nil {
		return nil, fmt.Errorf("unable to write body for gzip %v request to %v: %w", method, route, err)
	}
	if err := g.Close(); err != nil {
		return nil, fmt.Errorf("unable to close writing of body for gzip %v request to %v: %w", method, route, err)
	}
	req, err = http.NewRequest(method, route, &buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create new gzip %v request to %v: %w", method, route, err)
	}

	logger = logger.WithSuccess(logging.Fields{"bufLen": buf.Len()})
	return req, nil
}

func createStandardRequest(route string, method string, body []byte, logger *logging.Logger) (req *http.Request, err error) {
	message := fmt.Sprintf("Creating standard %v request to %v", method, route)
	logger = logger.WithTraceTracking("createStandardRequest", message).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	req, err = http.NewRequest(method, route, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to create new %v request to %v: %w", method, route, err)
	}

	logger = logger.WithSuccess(logging.Fields{"bodyLen": len(body)})
	return req, nil
}

func createRequest(route string, method string, body []byte, headers RequestHeaders, logger *logging.Logger) (req *http.Request, err error) {
	message := fmt.Sprintf("Creating %v request to %v", method, route)
	logger = logger.WithTraceTracking("createRequest", message).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	if ContainsHeader(headers, HeaderContentEncoding, GzipContentEncoding) {
		req, err = createGzipRequest(route, method, body, logger)
	} else {
		req, err = createStandardRequest(route, method, body, logger)
	}
	if err != nil {
		return nil, err
	}

	for header, value := range headers {
		req.Header.Add(header, value)
	}

	// prep the request headers
	SkuidUserAgent := fmt.Sprintf("%s/%s", constants.ProjectName, constants.VersionName)
	req.Header.Set(HeaderUserAgent, SkuidUserAgent)

	logger = logger.WithSuccess()
	return req, nil
}

func sendRequest(req *http.Request, logger *logging.Logger) (resp *http.Response, err error) {
	message := fmt.Sprintf("Sending %v request to %v", req.Method, req.URL)
	logger = logger.WithTraceTracking("sendRequest", message).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	// SKuid Review Required - The comment on the last line existed in v0.6.7 but its not clear what it means.  As we know, errors
	// occur server side for any number of reasons (e.g., server not properly inspecting payload and returning HTTP 500) so I'm not
	// sure if the below is referring to something in the CLI or on the server?  Either way, seems incorrect but leaving it here
	// since not clear what its intended meaning/purpose is.  Can you clarify?
	// 		`perform the request. errors only pop up if there's an issue with assembly/resources.`
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed making request to %v: %w", req.URL, err)
	}

	logger = logger.WithSuccessMessage(fmt.Sprintf(" with status code %v", resp.StatusCode), logging.Fields{"statusCode": resp.StatusCode})
	return resp, nil
}

func readResponse(resp *http.Response, logger *logging.Logger) (responseBody []byte, err error) {
	message := fmt.Sprintf("Reading %v response from %v", resp.Request.Method, resp.Request.URL)
	logger = logger.WithTraceTracking("readResponse", message).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from %v: %w", resp.Request.URL, err)
	}
	defer resp.Body.Close()

	logger = logger.WithSuccess(logging.Fields{"responseBodyLen": len(responseBody)})
	return responseBody, nil
}

func isJSONResponse(resp *http.Response) bool {
	// very primitive check to determine if response is JSON
	contentType := resp.Header.Get("Content-Type")
	if len(contentType) == 0 {
		return false
	}
	media, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	return strings.HasPrefix(media, "application/json")
}

func getError(resp *http.Response, responseBody []byte) error {
	var body string
	if isJSONResponse(resp) {
		if jsonPath, err := json.CreatePath("$.message"); err == nil {
			var responseMsg string
			if err := jsonPath.Unmarshal(responseBody, &responseMsg); err == nil && strings.TrimSpace(responseMsg) != "" {
				// Skuid review required - Is there a consistent approach to the structure of "application" errors returned
				// from APIs across all APIs?  There doesn't seem to be a custom object of any type and just using the message
				// property but how to reliably differentiate between an "application error message" and a "general error
				// message"?  Is HTTP 400 always used when "application errors occur"?  Any other codes?  Is there any way to
				// do this reliably?
				// TODO: Depending on answer to above, may have to just return an error with full details rather than just
				// a "meaningful application message" if there is no reliable way to detect when the message is meaningful
				// vs system level error
				if resp.StatusCode == http.StatusBadRequest {
					return errors.New(responseMsg)
				} else {
					body = responseMsg
				}
			}
		}
	}
	if body == "" {
		body = string(responseBody)
	}

	return fmt.Errorf("request to %v failed with status code %v: '%v'", resp.Request.URL, resp.StatusCode, body)
}
