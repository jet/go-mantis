package rest

import (
	"bytes"
	"encoding"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// ErrBodyNotReReadable is returned by NewRetryableRequest when the given body cannot be re-read
var ErrBodyNotReReadable = errors.New(errPrefix + "body is not re-readable")

// NewRetryableRequest helps create http requests that can be retried
// It is preferable to give a body which implements `Len() int`
// Known Content-Length:
// - `nil`
// - `*bytes.Buffer`
// - `*bytes.Reader`
// - `*strings.Reader`
// - `string`
// - `[]rune`
// - `[]byte`
// - `encoding.BinaryMarshaler`
// - `encoding.TextMarshaler`
//
// Potentially unknown Content-Length:
// - `io.ReadSeeker`
// - `func() (io.ReadCloser, error)`
// - `io.Reader`: This is tested last. This will read the entire body into a buffer and use that buffer for repeating requests.
//   this is only acceptable for small bodies.
func NewRetryableRequest(method string, url string, body interface{}) (*http.Request, error) {
	if body == nil || body == http.NoBody {
		return http.NewRequest(method, url, nil)
	}
	switch b := body.(type) {
	case *bytes.Buffer: // Special case already accounted for in http.NewRequest
		return http.NewRequest(method, url, b)
	case *bytes.Reader: // Special case already accounted for in http.NewRequest
		return http.NewRequest(method, url, b)
	case *strings.Reader: // Special case already accounted for in http.NewRequest
		return http.NewRequest(method, url, b)
	case string: // convert strings to *strings.Reader
		return http.NewRequest(method, url, strings.NewReader(b))
	case []rune: // convert []rune to *strings.Reader
		return http.NewRequest(method, url, strings.NewReader(string(b)))
	case []byte: // convert []byte to *bytes.Reader
		return http.NewRequest(method, url, bytes.NewReader(b))
	case encoding.BinaryMarshaler:
		bs, err := b.MarshalBinary()
		if err != nil {
			return nil, errors.Wrapf(err, errPrefix+"could not create retryable request")
		}
		return http.NewRequest(method, url, bytes.NewReader(bs))
	case encoding.TextMarshaler:
		bs, err := b.MarshalText()
		if err != nil {
			return nil, errors.Wrapf(err, errPrefix+"could not create retryable request")
		}
		return http.NewRequest(method, url, bytes.NewReader(bs))
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	type lengther interface {
		Len() int
	}
	switch b := body.(type) {
	case io.ReadSeeker:
		// Can seek back, so don't have to buffer.
		// Assumes that closing is done by the caller
		if l, ok := b.(lengther); ok {
			req.ContentLength = int64(l.Len())
		}
		req.Body = ioutil.NopCloser(b)
		req.GetBody = func() (io.ReadCloser, error) {
			_, err := b.Seek(0, 0)
			if err != nil {
				return nil, err
			}
			return ioutil.NopCloser(b), nil
		}
	case func() (io.ReadCloser, error):
		// Matches "GetBody" signature; so just use that. Takes ownership of closing the body.
		rc, err := b()
		if err != nil {
			return nil, err
		}
		if l, ok := rc.(lengther); ok {
			req.ContentLength = int64(l.Len())
		}
		req.Body = rc
		req.GetBody = b
	case io.Reader: // Worst case; buffer it all in memory
		buf, err := ioutil.ReadAll(b)
		if err != nil {
			return nil, err
		}
		req.ContentLength = int64(len(buf))
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(buf)), nil
		}
	default:
		return nil, ErrBodyNotReReadable
	}
	return req, nil
}

func replayBody(req *http.Request) {
	if req.Body != nil {
		// Close and ignore the error (if it was already closed)
		req.Body.Close()
		nb, err := req.GetBody()
		if err != nil {
			// Returns error on next read
			req.Body = ioutil.NopCloser(errReader{Err: err})
			return
		}
		// Re-read with new body
		req.Body = nb
	}
}

func isReplayable(req *http.Request) bool {
	if req.GetBody == nil {
		if req.Body != nil && req.Body != http.NoBody && req.ContentLength != 0 {
			return false
		}
	}
	return true
}

func mustBeReplayable(req *http.Request) {
	if !isReplayable(req) {
		panic(errPrefix + "request with non-empty body is not retryable. Must have GetBody set.")
	}
}

type errReader struct {
	Err error
}

func (e errReader) Read(bs []byte) (int, error) {
	return 0, e.Err
}
