package rest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"
)

type FakeRequester struct {
	Test      *testing.T
	Index     int
	Responses []FakeResponse
	Requests  []RequestRecord
}

type RequestRecord struct {
	Request *http.Request
	Time    time.Time
}

type MatchRequestFn func(r *http.Request) bool

type FakeResponse struct {
	Err         error
	StatusCode  int
	ContentType string
	Body        []byte
	BodyReader  io.Reader
	Headers     http.Header
}

func FakeRequest(method string, url string, body []byte) *http.Request {
	if body == nil {
		req, _ := http.NewRequest(method, url, nil)
		return req
	}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(body)), nil
	}
	return req
}

func (r FakeResponse) Status(code int) FakeResponse {
	r.StatusCode = code
	return r
}

func (r FakeResponse) WithBody(s string) FakeResponse {
	r.Body = []byte(s)
	return r
}

func (r FakeResponse) Header(key, value string) FakeResponse {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Add(key, value)
	return r
}

func (r FakeResponse) Do(req *http.Request) (*http.Response, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Make(), nil
}

type FakeReadCloser struct {
	br  *bytes.Buffer
	err error
}

func (r FakeReadCloser) Read(bs []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.br.Read(bs)
}

func (r FakeReadCloser) Close() error {
	return nil
}

func (r FakeResponse) Make() *http.Response {
	resp := &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	if r.BodyReader != nil {
		bbs, err := ioutil.ReadAll(r.BodyReader)
		if err != nil {
			r.Body = bbs
		}
		resp.Body = FakeReadCloser{br: bytes.NewBuffer(r.Body), err: err}
	} else {
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(r.Body))
	}
	if r.StatusCode != 0 {
		resp.Status = http.StatusText(r.StatusCode)
		resp.StatusCode = r.StatusCode
	}
	if r.ContentType != "" {
		resp.Header.Set("Content-Type", r.ContentType)
	}
	if r.StatusCode == 0 {
		if len(r.Body) > 0 {
			resp.Status = http.StatusText(r.StatusCode)
			resp.StatusCode = http.StatusNoContent
		} else {
			resp.Status = http.StatusText(r.StatusCode)
			resp.StatusCode = http.StatusOK
		}
	}
	resp.Status = http.StatusText(r.StatusCode)
	resp.ContentLength = int64(len(r.Body))
	date := time.Date(2018, time.December, 25, 9, 30, 59, 0, time.UTC)
	resp.Header.Set("Date", date.Format(http.TimeFormat))
	resp.Header.Set("Server", "fake-server")
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(r.Body)))
	if len(r.Body) > 0 {
		if r.ContentType == "" {
			resp.Header.Set("Content-Type", "application/octet-stream")
		} else {
			resp.Header.Set("Content-Type", r.ContentType)
		}
	}
	// Clone headers
	for hn, hvs := range r.Headers {
		for _, hv := range hvs {
			resp.Header.Add(hn, hv)
		}
	}
	return resp
}

func (r *FakeRequester) Do(req *http.Request) (*http.Response, error) {
	if r.Index >= len(r.Responses) {
		r.Index = 0
	}
	r.Requests = append(r.Requests, RequestRecord{Request: req, Time: time.Now()})
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	rs := r.Responses[r.Index]
	r.Index++
	rb, _ := httputil.DumpRequest(req, true)
	r.Test.Logf("==== request ====>>>\n%s", string(rb))
	resp, err := rs.Do(req)
	if err != nil {
		return nil, err
	}
	rb, _ = httputil.DumpResponse(resp, true)
	r.Test.Logf("<<<==== response ====\n%s", string(rb))
	return resp, nil
}

func TestHTTPClient(t *testing.T) {
	if h1, h2 := HTTPClient(), HTTPClient(); h1 == h2 {
		t.Fatal("expected successive calls of HTTPClient to return different clients")
	}
}
