package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// NewHttpError returns an error that formats as the given
func NewHttpError(status string, body io.Reader) error {
	var httpError HTTPError
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("%v - %v", err.Error(), "Could not parse error message")
	}
	bodyString := string(bodyBytes)
	// Attempt to parse the body
	err = json.Unmarshal(bodyBytes, &httpError)
	if err != nil {
		return fmt.Errorf("%v - %v", err.Error(), bodyString)
	}

	httpError.Body = bodyString
	httpError.Status = status

	return &httpError
}

// HTTPError generates a structure error from an http response
type HTTPError struct {
	StatusCode    int `json:"statusCode"`
	Status        string
	Body          string
	ParsedError   string `json:"error"`
	ParsedMessage string `json:"message"`
}

func (e *HTTPError) Error() string {
	if e.ParsedError != "" && e.ParsedMessage != "" {
		return fmt.Sprintf("%v - %v - %v", e.Status, e.ParsedError, e.ParsedMessage)
	}
	if e.ParsedError != "" {
		return fmt.Sprintf("%v - %v", e.Status, e.ParsedError)
	}
	if e.ParsedMessage != "" {
		return fmt.Sprintf("%v - %v", e.Status, e.ParsedMessage)
	}
	return fmt.Sprintf("%v - %v", e.Status, e.Body)
}
