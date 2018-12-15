package rest

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"
	"time"
)

func readTestData(t *testing.T, file string) []byte {
	t.Helper()
	fd, err := os.Open("./testdata/" + file)
	if err != nil {
		t.Fatalf("could not open test data '%s': %v", file, err)
	}
	bs, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatalf("could not read test data '%s': %v", file, err)
	}
	return bs
}

func TestLoggingRequester(t *testing.T) {
	expectedReq := readTestData(t, "request.txt")
	expectedResp := readTestData(t, "response.txt")
	var dumpReq []byte
	var dumpResp []byte
	respHeaders := make(http.Header)
	respHeaders.Set("x-response-id", "def")
	rs := &FakeRequester{
		Responses: []FakeResponse{
			FakeResponse{StatusCode: 200, Headers: respHeaders},
		},
		Test: t,
	}
	rr := &LoggingRequester{
		RequestLogger: func(req *http.Request) {
			dumpReq, _ = httputil.DumpRequest(req, true)
			//ioutil.WriteFile("./testdata/request.txt", dumpReq, 0644)
		},
		// Log the message 'http response' on each response
		ResponseLogger: func(resp *http.Response, req *http.Request, dur time.Duration) {
			dumpResp, _ = httputil.DumpResponse(resp, true)
			//ioutil.WriteFile("./testdata/response.txt", dumpResp, 0644)
		},
		Requester: rs,
	}
	method := "POST"
	url := "http://example.com/test/path"
	req := FakeRequest(method, url, []byte("hello"))
	req.Header.Set("x-request-id", "abc")
	rr.Do(req)
	if !bytes.Equal(expectedReq, dumpReq) {
		t.Error("expected request does not match")
	}
	if !bytes.Equal(expectedResp, dumpResp) {
		t.Error("expected response does not match")
	}
}

func ExampleLoggingRequester() {
	rr := RetryAfterRequester{
		// Log requests and responses at debug level
		Requester: LoggingRequester{
			// Log the message 'http request'
			RequestLogger: func(req *http.Request) {
				log.Printf("request made to %s", req.URL)
			},
			// Log the message 'http response' on each response
			ResponseLogger: func(resp *http.Response, req *http.Request, dur time.Duration) {
				log.Printf("response returned: %s after %v", resp.Status, dur)
			},
			Requester: HTTPClient(),
		},
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
