package rest

import (
	"net/http"
	"time"
)

const (
	// DefaultRetryAttempts is used when no attempt count is given
	DefaultRetryAttempts = 10
)

// Backoff policy implements this interface which returns the duration that
// should be waited between each try at a given 'try' number
type Backoff func(try uint) time.Duration

// RetryBackoffRequester performs a request with retries,
// using the provided backoff strategy between each attempt
type RetryBackoffRequester struct {
	// Attempts is the number of times a request will be re-tried in addition to the initial attempt
	Attempts uint
	// Backoff is the strategy used to delay between attempts
	Backoff Backoff
	// ResponseTester is a function that returns true if the request should be retried
	ResponseTester ResponseTester
	// Requester performs the actual request for each attempt
	Requester Requester
}

// ResponseTester tests the results returned by Requester.Do
// If the response is successful, it will return true, otherwise it returns false.
// If an error is returned, the response is considered a fatal error and will not be retried
type ResponseTester func(*http.Response, error) (bool, error)

type retryResponse struct {
	Response *http.Response
	Err      error
}

func (r retryResponse) Error() string {
	return r.Err.Error()
}

// Do performs a request
func (r RetryBackoffRequester) Do(req *http.Request) (*http.Response, error) {
	mustBeReplayable(req)
	var resp *http.Response
	var err error
	var backoff = r.Backoff
	var tester = r.ResponseTester
	attempts := r.Attempts
	if backoff == nil {
		// no backoff
		backoff = func(try uint) time.Duration {
			return time.Duration(0)
		}
	}
	if attempts == 0 {
		attempts = DefaultRetryAttempts
	}
	if tester == nil {
		tester = DefaultResponseTester
	}
	ctx := req.Context()
	for i := uint(0); i <= attempts; i++ {
		// attempt to request
		resp, err = r.Requester.Do(req)
		// test the response
		success, ferr := tester(resp, err)
		if success {
			return resp, nil
		}
		if ferr != nil {
			return nil, ferr
		}
		resp.Body.Close()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff(i + 1)):
			replayBody(req)
			req = WithTryCount(req, i+1)
		}
	}
	return resp, err
}
