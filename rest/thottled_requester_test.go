package rest

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type constantLimit time.Duration

func (l constantLimit) Delay() time.Duration {
	return time.Duration(l)
}

func TestThrottledRequesterContextDeadline(t *testing.T) {
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: 200},
		},
		Test: t,
	}
	delay := 100 * time.Millisecond
	rr := &ThrottledRequester{
		Limiter:   constantLimit(delay),
		Requester: rs,
	}
	req := FakeRequest("POST", "http://example.com", []byte("hello"))
	ctx, cancel := context.WithTimeout(context.Background(), delay/2)
	defer cancel()
	req = req.WithContext(ctx)
	_, err := rr.Do(req)
	if err == nil {
		t.Fatal("expected response error")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatalf("expected 'context deadline exceeded', got '%v'", err)
	}
	t.Logf("%v", err)
}

func TestThrottledRequesterContextCancel(t *testing.T) {
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: 200},
		},
		Test: t,
	}
	delay := 100 * time.Millisecond
	rr := &ThrottledRequester{
		Limiter:   constantLimit(delay),
		Requester: rs,
	}
	req := FakeRequest("POST", "http://example.com", []byte("hello"))
	ctx, cancel := context.WithTimeout(context.Background(), delay*2)
	cancel()
	req = req.WithContext(ctx)
	_, err := rr.Do(req)
	if err == nil {
		t.Fatal("expected response error")
	}
	if err.Error() != "context canceled" {
		t.Fatalf("expected 'context deadline exceeded', got '%v'", err)
	}
	t.Logf("%v", err)
}

func TestThrottledRequester(t *testing.T) {
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: 200},
		},
		Test: t,
	}
	delay := 100 * time.Millisecond
	rr := &ThrottledRequester{
		Limiter:   constantLimit(delay),
		Requester: rs,
	}
	for i := 0; i < 10; i++ {
		req := FakeRequest("POST", "http://example.com", []byte("hello"))
		req.Header.Set("Request-Num", fmt.Sprintf("%d", i))
		rr.Do(req)
	}
	for i := 1; i < len(rs.Requests); i++ {
		t1 := rs.Requests[i-1].Time
		t2 := rs.Requests[i].Time
		d := t2.Sub(t1)
		if d < delay {
			t.Fatalf("expected request interval to be >= delay duration: %v < %v", d, delay)
		}
	}
}
