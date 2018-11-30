package retry

import (
	"context"
	"time"
)

const (
	// DefaultAttempts is the default number of retry attempts if left unconfigured
	DefaultAttempts = 10
	// DefaultBackoffTime is the default wait time for retry attempts
	DefaultBackoffTime = 100 * time.Millisecond
)

// Option configures retry options
type Option func(cfg *config)

// Delay configures the backoff strategy for retries
func Delay(backoff BackoffFunc) Option {
	return func(cfg *config) {
		cfg.Backoff = backoff
	}
}

// Attempts option configure the maximum number of retries beyond the initial attempt
// If attempts = 0, only the initial attempt will be run - this effectively disables retries
func Attempts(attempts uint) Option {
	return func(cfg *config) {
		cfg.Attempts = attempts
	}
}

// ForErrors option configures which errors will be retried
func ForErrors(fn ErrorTestFunc) Option {
	return func(cfg *config) {
		cfg.ErrorTest = fn
	}
}

// OnRetry option configures the function which is called after every failed attempt
func OnRetry(fn OnRetryFunc) Option {
	return func(cfg *config) {
		cfg.OnRetry = fn
	}
}

// WithContext option configures the retry context
// This is used for early cancellation. On each round of try, the context object is check to see if it has expired.
// - If the context expires before the first try, then the ctx.Err() is returned
// - If the context expires at any other time, then the last error is returned
func WithContext(ctx context.Context) Option {
	return func(cfg *config) {
		cfg.Context = ctx
	}
}

// RetryableFunc is the signature of a function which can be retried by `retry.Do`
type RetryableFunc func() error

// OnRetryFunc is a function signature that is called before every retry with the current try number (starting from 1) and the last error from the previous attempt.
type OnRetryFunc func(try uint, err error)

// ErrorTestFunc is the signature of a function that returns true if the next retry should be attempted, after inspection of the error
type ErrorTestFunc func(err error) bool

type config struct {
	// Context for cancellation
	Context context.Context
	// Attempts specifies the maximum number of times that DoWithRetry will retry the function, in addition to the first attempt.
	// - Attempts = 0 means that the DoRetryFunc will be executed exactly once.
	// - Attempts > 0 means that it will be executed at most `1 + MaxRetries` times.
	Attempts uint
	// Backoff configures the back-off policy. Between retries, it will call .Delay(try uint) on
	// the Backoff object, which, depending on implementation, should return the time.Duration to wait between retries on that attempt.
	Backoff BackoffFunc
	// ErrorTestFunc inspects errors returned from DoRetryFunc and returns true if
	// the error was non-fatal; can be tried again (after back-off)
	// If this is nil, all errors will trigger a retry
	ErrorTest ErrorTestFunc
	// OnRetry is called before every retry with the last error, and the current try (starting from 1)
	OnRetry OnRetryFunc
}
