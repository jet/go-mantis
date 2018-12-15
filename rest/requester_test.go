package rest

import (
	"fmt"
	"net/http"
	"testing"
)

func TestResponseTesterStatusCodesNoDelegate(t *testing.T) {
	statusCodes := map[int]bool{
		// Server Errors
		http.StatusBadGateway: false,
		http.StatusConflict:   false,

		// Failed 2xx
		http.StatusNoContent: false,
		// Success 2xx
		http.StatusOK: true,
	}
	tests := []struct {
		Request FakeResponse
		Success bool
		Err     error
	}{
		// request error
		{
			Request: FakeResponse{Err: fmt.Errorf("request error")},
			Success: false,
			Err:     fmt.Errorf("request error"),
		},
		// Explicit response for 502
		{
			Request: FakeResponse{StatusCode: http.StatusBadGateway},
			Success: false,
			Err:     nil,
		},
		// Explicit response for 409
		{
			Request: FakeResponse{StatusCode: http.StatusConflict},
			Success: false,
			Err:     nil,
		},
		// Explicit response for 204
		{
			Request: FakeResponse{StatusCode: http.StatusNoContent},
			Success: false,
			Err:     nil,
		},
		// Explicit response for 200
		{
			Request: FakeResponse{StatusCode: http.StatusOK},
			Success: true,
			Err:     nil,
		},
		// Default response for 2xx codes
		{
			Request: FakeResponse{StatusCode: http.StatusCreated},
			Success: false,
			Err:     fmt.Errorf(httpErrPrefix + ": 201 Created"),
		},
		// Default response for 5xx
		{
			Request: FakeResponse{StatusCode: http.StatusGatewayTimeout},
			Success: false,
			Err:     fmt.Errorf(httpErrPrefix + ": 504 Gateway Timeout"),
		},
		// Default response for 4xx
		{
			Request: FakeResponse{StatusCode: http.StatusNotFound},
			Success: false,
			Err:     fmt.Errorf(httpErrPrefix + ": 404 Not Found"),
		},
	}
	tester := ResponseTesterStatusCodes(nil, statusCodes)
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			success, err := tester(test.Request.Do(nil))
			t.Logf("success=%t,err=%v", success, err)
			if test.Success != success {
				t.Errorf("expected success=%t but got success=%t", test.Success, success)
			}
			if err != nil && test.Err == nil {
				t.Errorf("unexpected error %v", err)
			}
			if err == nil && test.Err != nil {
				t.Errorf("expected error %v not returned", test.Err)
			}
			if err != nil && test.Err != nil {
				if err.Error() != test.Err.Error() {
					t.Errorf("expected error %v but got %v", test.Err, err)
				}
			}
		})
	}
}

func TestResponseTesterStatusCodes(t *testing.T) {
	statusCodes := map[int]bool{
		// Server Errors
		http.StatusBadGateway: false,
		http.StatusConflict:   false,

		// Failed 2xx
		http.StatusNoContent: false,
		// Success 2xx
		http.StatusOK: true,
	}
	tests := []struct {
		Request FakeResponse
		Success bool
		Err     error
	}{
		// request error
		{
			Request: FakeResponse{Err: fmt.Errorf("request error")},
			Success: false,
			Err:     fmt.Errorf("request error"),
		},
		// Explicit response for 502
		{
			Request: FakeResponse{StatusCode: http.StatusBadGateway},
			Success: false,
			Err:     nil,
		},
		// Explicit response for 409
		{
			Request: FakeResponse{StatusCode: http.StatusConflict},
			Success: false,
			Err:     nil,
		},
		// Explicit response for 204
		{
			Request: FakeResponse{StatusCode: http.StatusNoContent},
			Success: false,
			Err:     nil,
		},
		// Explicit response for 200
		{
			Request: FakeResponse{StatusCode: http.StatusOK},
			Success: true,
			Err:     nil,
		},
		// Default response for 2xx codes
		{
			Request: FakeResponse{StatusCode: http.StatusCreated},
			Success: true,
			Err:     nil,
		},
		// Default response for 5xx
		{
			Request: FakeResponse{StatusCode: http.StatusGatewayTimeout},
			Success: false,
			Err:     nil,
		},
		// Default response for 4xx
		{
			Request: FakeResponse{StatusCode: http.StatusNotFound},
			Success: false,
			Err:     nil,
		},
	}
	tester := ResponseTesterStatusCodes(DefaultResponseTester, statusCodes)
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			success, err := tester(test.Request.Do(nil))
			t.Logf("success=%t,err=%v", success, err)
			if test.Success != success {
				t.Errorf("expected success=%t but got success=%t", test.Success, success)
			}
			if err != nil && test.Err == nil {
				t.Errorf("unexpected error %v", err)
			}
			if err == nil && test.Err != nil {
				t.Errorf("expected error %v not returned", test.Err)
			}
		})
	}

}
