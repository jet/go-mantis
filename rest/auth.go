package rest

import (
	"net/http"
)

// Authorizer adds authorization information to a request. Authorize receives an unauthorized http request and returns an http request with authorization information added. the request may be the same object, or it might be a new request.
type Authorizer func(req *http.Request) (*http.Request, error)

// AuthorizedRequester adds authorization to outgoing requests
type AuthorizedRequester struct {
	Requester
	Authorizer
}

func (c *AuthorizedRequester) request(req *http.Request) (*http.Response, error) {
	if c.Requester == nil {
		return http.DefaultClient.Do(req)
	}
	return c.Requester.Do(req)
}

func (c *AuthorizedRequester) authorize(req *http.Request) (*http.Request, error) {
	if c.Authorizer == nil {
		return req, nil
	}
	return c.Authorizer(req)
}

// Do applies authorization to the request before executing it
func (c *AuthorizedRequester) Do(req *http.Request) (*http.Response, error) {
	r, err := c.authorize(req)
	if err != nil {
		return nil, err
	}
	return c.request(r)
}

// HeaderAuthorizer creates an Authorizer for arbitrary header values
// This is used for apis that implement a static auth token
// Like `X-Auth-Token: 4200322b-2073-4e85-9186-959b2f4c49d1`
func HeaderAuthorizer(key string, value string) Authorizer {
	return func(req *http.Request) (*http.Request, error) {
		req.Header.Set(key, value)
		return req, nil
	}
}

// BasicAuthorizer creates an Authorizer for HTTP Basic Auth
// https://en.wikipedia.org/wiki/Basic_access_authentication
func BasicAuthorizer(user string, pass string) Authorizer {
	return func(req *http.Request) (*http.Request, error) {
		req.SetBasicAuth(user, pass)
		return req, nil
	}
}
