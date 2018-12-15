package rest

import (
	"testing"
	"time"
)

func TestRetryAfter(t *testing.T) {
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: 429}.Header("Retry-After", "1"),
			FakeResponse{StatusCode: 200},
		},
		Test: t,
	}
	rr := &RetryAfterRequester{
		Requester: rs,
	}
	req := FakeRequest("POST", "http://example.com", []byte("hello"))
	rr.Do(req)
	if len(rs.Requests) != 2 {
		t.Fatal("expected 2 requests to be made")
	}
	r1, r2 := rs.Requests[0], rs.Requests[1]
	td := r2.Time.Sub(r1.Time)
	if td < time.Second {
		t.Fatal("expected time between requests to exceed the Retry-After time")
	}
}
