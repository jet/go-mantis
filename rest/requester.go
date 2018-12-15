package rest

import (
	"net/http"
)

// Requester executes HTTP requests and returns a response
// An error is returned if the request could not be made
type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

// ResponseTesterStatusCodes wraps a `ResponseTester` that will test status codes
// If the response exists, it return the corresponding boolean that matches the status code in the map provided
// If there is no match, it will delegate to the given `tester`
// If no `tester` was given:
// - Request errors are returned immediately without testing the response code (it is assumed that response is nil in this case)
// - Unmatched response codes will convert the entire response to `ErrorHTTPResponse`.
//   Errors converted in this way will read up to `ErrorHTTPResponseBodyBytes` bytes from the response and close the response body.
func ResponseTesterStatusCodes(tester ResponseTester, sc map[int]bool) ResponseTester {
	return func(resp *http.Response, err error) (bool, error) {
		if err != nil {
			if tester != nil {
				return tester(resp, err)
			}
			return false, err
		}
		// Status code matched
		if s, ok := sc[resp.StatusCode]; ok {
			return s, nil
		}
		if tester != nil {
			return tester(resp, err)
		}
		return false, NewErrorHTTPResponseLimitBody(resp, ErrorHTTPResponseBodyBytes)
	}
}

// DefaultResponseTester returns true if the response status code is between 200 and 399 (inclusive)
func DefaultResponseTester(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return resp.StatusCode >= 200 && resp.StatusCode < 400, nil
}
