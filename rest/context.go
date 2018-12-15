package rest

import (
	"context"
	"net/http"
)

type contextKey string

// contextKeyTryCount used to extract the try counter from the request context
const (
	contextKeyTryCount    = contextKey("request_try_count")
	contextKeyRequestMeta = contextKey("request_metadata")
)

// TryCount extracts the retry count from the request context
//
//     try := TryCount(req)
func TryCount(req *http.Request) uint {
	val := req.Context().Value(contextKeyTryCount)
	if val == nil {
		return 0
	}
	try, ok := val.(uint)
	if ok {
		return try
	}
	return 0
}

// WithTryCount sets try count in the request context
//
//     req = WithTryCount(req, uint(1))
func WithTryCount(req *http.Request, try uint) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), contextKeyTryCount, try))
}
