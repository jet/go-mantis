package rest

import (
	"net/http"
	"time"
)

// ThrottledRequester waits before sending a request using the associated Waiter
type ThrottledRequester struct {
	Limiter   Limiter
	Requester Requester
}

// Limiter calculates the delay between successive requets.
// Implementers of Limiter can be stateful, and should expect that Delay is called before the request is made.
type Limiter interface {
	Delay() time.Duration
}

// Do request with throttling
func (r ThrottledRequester) Do(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(r.Limiter.Delay()):
		return r.Requester.Do(req)
	}
}
