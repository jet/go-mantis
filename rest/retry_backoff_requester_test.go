package rest

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestRetryRequesterDefaults(t *testing.T) {
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: http.StatusServiceUnavailable}.WithBody("unavailable"),
		},
		Test: t,
	}
	rr := RetryBackoffRequester{
		Requester: rs,
	}
	req := FakeRequest("POST", "http://example.com", []byte("hello"))
	rr.Do(req)
	if len(rs.Requests) != DefaultRetryAttempts+1 {
		t.Fatalf("expected %d requests to be made, got %d", DefaultRetryAttempts, len(rs.Requests))
	}
	for i, r := range rs.Requests {
		if ctx := r.Request.Context(); ctx != nil {
			if try := TryCount(r.Request); try != uint(i) {
				t.Errorf("expected request %d try count to be %[1]d, but was %d", i, try)
			}
		}
	}
}

func TestRetryRequesterBackoff(t *testing.T) {
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: http.StatusServiceUnavailable}.WithBody("unavailable"),
			FakeResponse{StatusCode: http.StatusInternalServerError}.WithBody("error"),
			FakeResponse{StatusCode: http.StatusBadGateway}.WithBody("bad gateway"),
			FakeResponse{StatusCode: http.StatusGatewayTimeout}.WithBody("timeout"),
			FakeResponse{StatusCode: http.StatusTooManyRequests}.WithBody("hold please"),
			FakeResponse{StatusCode: http.StatusOK}.WithBody("OK"),
		},
		Test: t,
	}
	boff := time.Millisecond * 100
	rr := RetryBackoffRequester{
		Attempts: 10,
		Backoff: func(try uint) time.Duration {
			return boff
		},
		ResponseTester: ResponseTesterStatusCodes(DefaultResponseTester, map[int]bool{
			// Fail but Retryable
			http.StatusServiceUnavailable:  false,
			http.StatusInternalServerError: false,
			http.StatusBadGateway:          false,
			http.StatusGatewayTimeout:      false,
			http.StatusTooManyRequests:     false,

			// Success
			http.StatusOK:        true,
			http.StatusNoContent: true,
		}),
		Requester: rs,
	}
	req := FakeRequest("POST", "http://example.com", []byte("hello"))
	rr.Do(req)
	if len(rs.Requests) != 6 {
		t.Fatalf("expected 6 requests to be made, got %d", len(rs.Requests))
	}
	for i, r := range rs.Requests {
		if ctx := r.Request.Context(); ctx != nil {
			if try := TryCount(r.Request); try != uint(i) {
				t.Errorf("expected request %d try count to be %[1]d, but was %d", i, try)
			}
		}
		if i > 0 {
			t1 := rs.Requests[i-1].Time
			t2 := rs.Requests[i].Time
			d := t2.Sub(t1)
			if d < boff {
				t.Fatalf("expected request interval to be >= backoff duration: %v < %v", d, boff)
			}
			t.Logf("actual back off: %v", d)
		}
	}
}

func ExampleRetryBackoffRequester() {
	rr := RetryBackoffRequester{
		Attempts: 10,
		Backoff: func(try uint) time.Duration {
			return 100 * time.Millisecond
		},
		ResponseTester: ResponseTesterStatusCodes(DefaultResponseTester, map[int]bool{
			// Fail but Retryable
			http.StatusServiceUnavailable:  false,
			http.StatusInternalServerError: false,
			http.StatusBadGateway:          false,
			http.StatusGatewayTimeout:      false,
			http.StatusTooManyRequests:     false,

			// Success
			http.StatusOK:        true,
			http.StatusNoContent: true,
		}),
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
