package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// RetryAfterRequester will retry requests
type RetryAfterRequester struct {
	// StatusCodes that this requester will retry on, given a Retry-After header is present.
	// If this is not set, it defaults to `[]int{http.StatusTooManyRequests}`
	StatusCodes []int
	// HeaderName from which the delay in seconds will be read.
	// If the header name is missing, the request will not be retried
	// If this is unset, it defaults to "Retry-After"
	HeaderName string
	// Requester makes the actual http request
	// This must be set
	Requester Requester
}

const headerRetryAfter = "Retry-After"

var defaultRetryAfterStatusCodes = []int{http.StatusTooManyRequests}

type retryAfterErr time.Duration

func (r retryAfterErr) Error() string {
	return fmt.Sprintf("retry-after: %v", time.Duration(r))
}

func parseDelay(str string) (time.Duration, error) {
	secs, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(secs) * time.Second, nil
}

// Do performs a request
func (r RetryAfterRequester) Do(req *http.Request) (*http.Response, error) {
	if r.Requester == nil {
		panic("Requester is nil")
	}
	mustBeReplayable(req)
	ctx := req.Context()
	header := r.HeaderName
	if header == "" {
		header = headerRetryAfter
	}
	var resp *http.Response
	var err error
	codes := make(map[int]struct{})
	for _, v := range r.StatusCodes {
		codes[v] = struct{}{}
	}
	if len(codes) == 0 {
		for _, v := range defaultRetryAfterStatusCodes {
			codes[v] = struct{}{}
		}
	}
	if err != nil {
		return nil, err
	}
	for { // Keep trying as needed
		resp, err = r.Requester.Do(req)
		if err != nil {
			return nil, err
		}
		if _, ok := codes[resp.StatusCode]; ok {
			hv := resp.Header.Get(header)
			if hv == "" { // no Retry-After header
				return resp, nil
			}
			var delay time.Duration
			if delay, err = parseDelay(hv); err != nil { //Parser error
				return resp, nil
			}
			resp.Body.Close()
			select {
			case <-time.After(delay):
				continue // retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return resp, err
	}
}

// ExampleRetryAfterRequester uses a requester that will detect a Retry-After header
// and retry the request after the provided number of seconds.
func ExampleRetryAfterRequester() {
	rr := RetryAfterRequester{
		Requester: HTTPClient(),
	}
	req, _ := NewRetryableRequest("POST", "http://example.com", "hello")
	resp, err := rr.Do(req)
	if err != nil {
		log.Println("error executing request: ", err)
		return
	}
	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading body: ", err)
		return
	}
	log.Println("body:", string(rb))
}
