package rest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type textMarshaller struct {
	text string
}

func (t *textMarshaller) MarshalText() (data []byte, err error) {
	return []byte(t.text), nil
}

type binaryMarshaller struct {
	bytes []byte
}

func (b *binaryMarshaller) MarshalBinary() (data []byte, err error) {
	return []byte(b.bytes), nil
}

type readSeeker struct {
	reader *bytes.Reader
}

func (rs *readSeeker) Seek(offset int64, whence int) (int64, error) {
	return rs.reader.Seek(offset, whence)
}

func (rs *readSeeker) Read(b []byte) (int, error) {
	return rs.reader.Read(b)
}

func (rs *readSeeker) Len() int {
	return rs.reader.Len()
}

func (rs *readSeeker) Close() error {
	return nil
}

type getBody struct {
	bytes []byte
}

func (gb *getBody) GetBody() (io.ReadCloser, error) {
	return &readSeeker{reader: bytes.NewReader(gb.bytes)}, nil
}

type reader struct {
	reader *bytes.Reader
}

func (rs *reader) Read(b []byte) (int, error) {
	return rs.reader.Read(b)
}

func TestReplayGetBodyError(t *testing.T) {
	err := fmt.Errorf("error reading body")
	rdr := strings.NewReader("hola")
	req, _ := http.NewRequest("POST", "http://example.com", rdr)
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, err
	}
	replayBody(req)
	_, err2 := ioutil.ReadAll(req.Body)
	if err != err2 {
		t.Fatalf("expected replay to result in a read error")
	}
}
func TestNonReplayableRequest(t *testing.T) {
	rdr := strings.NewReader("hola")
	req, _ := http.NewRequest("POST", "http://example.com", nil)
	req.Body = ioutil.NopCloser(rdr)
	req.ContentLength = int64(rdr.Len())
	if isReplayable(req) {
		t.Fatalf("expected not replayable")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	mustBeReplayable(req)
}

func TestRetryableRequestBodies(t *testing.T) {
	tests := []struct {
		body interface{}
		len  int64
		err  error
	}{
		{body: nil},
		{body: strings.NewReader("hi"), len: 2},
		{body: bytes.NewBuffer([]byte("hola")), len: 4},
		{body: bytes.NewReader([]byte("γεια")), len: 8},
		{body: "Привет", len: 12},
		{body: []rune("你好"), len: 6},
		{body: []byte("こんにちは"), len: 15},
		{body: &binaryMarshaller{bytes: []byte{0x32, 0x34, 0x36, 0x38}}, len: 4},
		{body: &textMarshaller{text: "Bonjour"}, len: 7},
		{body: &readSeeker{reader: bytes.NewReader([]byte{0x32, 0x34, 0x36, 0x38})}, len: 4},
		{body: (&getBody{bytes: []byte{0x32, 0x34, 0x36, 0x38}}).GetBody, len: 4},
		{body: &reader{reader: bytes.NewReader([]byte("ciao"))}, len: 4},
		{body: 14, len: 0, err: ErrBodyNotReReadable},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%T", i, test.body), func(t *testing.T) {
			req, err := NewRetryableRequest("POST", "http://example.com", test.body)
			if err != test.err {
				if test.err == nil {
					t.Fatalf("unexpected error: %v", err)
				} else if err == nil {
					t.Fatalf("expected error: %v", test.err)
				} else {
					t.Fatalf("unexpected error: expected '%v', got '%s'", test.err, err)
				}
			}
			if err == nil {
				if req == nil {
					t.Fatal("unexpected nil request")
				}
				if !isReplayable(req) {
					t.Error("expected replayable request")
				}
				if req.ContentLength != test.len {
					t.Errorf("expected Content-Length: %d, got %d", test.len, req.ContentLength)
				}
				if req.Body != nil {
					r1, _ := ioutil.ReadAll(req.Body)
					replayBody(req)
					r2, _ := ioutil.ReadAll(req.Body)
					if !bytes.Equal(r1, r2) {
						t.Fatal("body is not equal on re-read")
					}
				}
			}
		})
	}
}
