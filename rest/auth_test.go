package rest

import (
	"net/http"
	"testing"
)

func TestBasicAuthorizer(t *testing.T) {
	req := FakeRequest("GET", "http://example.com", nil)
	user, pass, ok := req.BasicAuth()
	if ok {
		t.Fatalf("expected no basic auth before applying authorizer")
	}
	auth := BasicAuthorizer("user", "pass")
	req, err := auth(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user, pass, ok = req.BasicAuth()
	if !ok {
		t.Fatalf("expected basic auth after applying authorizer")
	}
	if user != "user" {
		t.Error("expected basic auth user = 'user'")
	}
	if pass != "pass" {
		t.Error("expected basic auth password = 'pass'")
	}
}

func TestHeaderAuth(t *testing.T) {
	header, token := "X-Auth-Token", "token-id"
	req := FakeRequest("GET", "http://example.com", nil)

	if hv := req.Header.Get(header); hv != "" {
		t.Fatalf("expected no header value for %s auth before applying authorizer", header)
	}
	auth := HeaderAuthorizer(header, token)
	req, err := auth(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hv := req.Header.Get(header); hv != token {
		t.Errorf("expected %s = %s; but was %s", header, token, hv)
	}
}

func TestAuthorizedRequester(t *testing.T) {

	header, token := "X-Auth-Token", "token-id"
	req := FakeRequest("GET", "http://example.com", nil)
	if hv := req.Header.Get(header); hv != "" {
		t.Fatalf("expected no header value for %s auth before applying authorizer", header)
	}
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: http.StatusOK}.WithBody("OK"),
		},
		Test: t,
	}
	ar := AuthorizedRequester{
		Requester:  rs,
		Authorizer: HeaderAuthorizer(header, token),
	}
	resp, err := ar.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("unexpected nil response")
	}
	req2 := rs.Requests[0].Request
	if hv := req2.Header.Get(header); hv != token {
		t.Errorf("expected %s = %s; but was %s", header, token, hv)
	}
}
