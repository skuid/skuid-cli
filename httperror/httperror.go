package httperror

import "fmt"

// New returns an error that formats as the given text.
func New(statusCode string, body string) error {
	return &httpError{statusCode, body}
}

type httpError struct {
	statusCode string
	body       string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%v - %v", e.statusCode, e.body)
}
