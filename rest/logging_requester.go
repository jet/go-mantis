package rest

import (
	"net/http"
	"time"
)

// LoggingRequester may log the request and response
type LoggingRequester struct {
	// RequestLogger is given the request before it is executed
	RequestLogger func(*http.Request)
	// ResponseLogger may log the response
	ResponseLogger func(*http.Response, *http.Request, time.Duration)
	// Requester to call with the wrapped request. If nil, then Do will panic
	Requester Requester
}

// Do executes the request and logs requests / responses
func (r LoggingRequester) Do(req *http.Request) (*http.Response, error) {
	if r.Requester == nil {
		panic("Requester is nil")
	}
	if r.RequestLogger != nil {
		r.RequestLogger(req)
	}
	t0 := time.Now()
	resp, err := r.Requester.Do(req)
	if err != nil {
		return nil, err
	}
	if r.ResponseLogger != nil {
		r.ResponseLogger(resp, req, time.Since(t0))
	}
	return resp, nil
}
