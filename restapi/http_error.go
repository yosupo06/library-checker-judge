package main

import (
	"errors"
	"fmt"
)

// HTTPError decorates an error with an HTTP status code.
type HTTPError struct {
	Status  int
	Message string
}

func (e HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("http status %d", e.Status)
}

func newHTTPError(status int, message string) error {
	return HTTPError{Status: status, Message: message}
}

func getHTTPError(err error) (HTTPError, bool) {
	if err == nil {
		return HTTPError{}, false
	}
	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		return HTTPError{}, false
	}
	if httpErr.Status == 0 {
		httpErr.Status = 500
	}
	return httpErr, true
}
