package rest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	errPrefix     = "mantis/rest: "
	httpErrPrefix = "mantis/rest.http"
	// ErrorHTTPResponseBodyBytes is the default maximum amount of bytes read from a response when constructing an HTTP Error
	ErrorHTTPResponseBodyBytes = int64(1000000) // 1MB
)

// ErrorHTTPResponse contains an error response from an HTTP request
type ErrorHTTPResponse struct {
	response *http.Response
	body     []byte
	status   int
}

// NewErrorHTTPResponseLimitBody converts a http response into an error. It can limit the number of bytes that are read from the response body.
// If the limit is <= 0, then there is no limit, and the entire body will be read.
func NewErrorHTTPResponseLimitBody(resp *http.Response, limit int64) error {
	if resp.Body != nil {
		var err error
		var body []byte
		if limit <= 0 {
			body, err = ioutil.ReadAll(resp.Body)
		} else {
			body, err = ioutil.ReadAll(io.LimitReader(resp.Body, limit))
		}
		if err != nil {
			return &ErrorHTTPResponse{response: resp, status: resp.StatusCode}
		}
		defer resp.Body.Close()
		return &ErrorHTTPResponse{body: body, response: resp, status: resp.StatusCode}
	}
	return &ErrorHTTPResponse{status: resp.StatusCode, response: resp}
}

// NewErrorHTTPResponse converts a http response into an error. This will read the entire response body.
func NewErrorHTTPResponse(resp *http.Response) error {
	return NewErrorHTTPResponseLimitBody(resp, -1)
}

// Error returns a string representation of the http error response
func (e *ErrorHTTPResponse) Error() string {
	if e == nil {
		return httpErrPrefix
	}
	if len(e.body) > 0 {
		return fmt.Sprintf("%s: %d %s; %s", httpErrPrefix, e.status, http.StatusText(e.status), string(e.body))
	}
	return fmt.Sprintf("%s: %d %s", httpErrPrefix, e.status, http.StatusText(e.status))
}

// Status returns the http status code of the error response
func (e *ErrorHTTPResponse) Status() int {
	return e.status
}

// Body returns the body of the response read (up to 1MB)
func (e *ErrorHTTPResponse) Body() []byte {
	return e.body
}
