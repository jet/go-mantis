package rest

import (
	"bytes"
	"fmt"
	"testing"
)

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		response FakeResponse
		want     error
		limit    int64
	}{
		{
			response: FakeResponse{StatusCode: 404, Body: nil},
			want:     fmt.Errorf(httpErrPrefix + ": 404 Not Found"),
			limit:    ErrorHTTPResponseBodyBytes,
		},
		{
			response: FakeResponse{StatusCode: 404, BodyReader: errReader{Err: fmt.Errorf("read error")}},
			want:     fmt.Errorf(httpErrPrefix + ": 404 Not Found"),
			limit:    ErrorHTTPResponseBodyBytes,
		},
		{
			response: FakeResponse{StatusCode: 404, Body: []byte("error response")},
			want:     fmt.Errorf(httpErrPrefix + ": 404 Not Found; error response"),
			limit:    ErrorHTTPResponseBodyBytes,
		},
		{
			response: FakeResponse{StatusCode: 404, Body: []byte("error response")},
			want:     fmt.Errorf(httpErrPrefix + ": 404 Not Found; error response"),
			limit:    0,
		},
		{
			response: FakeResponse{StatusCode: 404, Body: []byte("error response")},
			want:     fmt.Errorf(httpErrPrefix + ": 404 Not Found; error"),
			limit:    5,
		},
		{
			response: FakeResponse{StatusCode: 404, Body: []byte("error response")},
			want:     fmt.Errorf(httpErrPrefix + ": 404 Not Found; error response"),
			limit:    -1,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			resp := test.response.Make()
			var err error
			if test.limit >= 0 {
				err = NewErrorHTTPResponseLimitBody(resp, test.limit)
			} else {
				err = NewErrorHTTPResponse(resp)
			}
			if err.Error() != test.want.Error() {
				t.Errorf("expected: [%v], got [%v]", test.want, err)
			}
			herr, ok := err.(*ErrorHTTPResponse)
			if !ok {
				t.Fatalf("expected error to by of type *ErrorHTTPResponse, but was %T instead", err)
			}
			if herr.Status() != test.response.StatusCode {
				t.Errorf("expected: err.Status() %d, got %d", test.response.StatusCode, herr.Status())
			}
			if test.response.Body != nil {
				body := test.response.Body
				if test.limit > 0 && len(test.response.Body) > int(test.limit) {
					body = test.response.Body[0:test.limit]
				}
				if !bytes.Equal(body, herr.Body()) {
					t.Errorf("error body does not equal response body")
				}
			}
		})
	}
}
