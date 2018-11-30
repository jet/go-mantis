// Package retry is a library for retrying operations and backing off between tries
package retry // import "github.com/jet/go-mantis/retry"

import (
	"context"
	"time"
)

// Do executes the function provided and retries based on the options provided.
// The Default Options are:
//
// - `Attempts(DefaultAttempts)`
// - `Delay(ConstantBackoff(DefaultBackoffTime))`
//
// If all tries failed, the last erorr the function encountered is returned.
//
func Do(fn RetryableFunc, opts ...Option) error {
	cfg := &config{
		Attempts: DefaultAttempts,
		Backoff:  ConstantBackoff(DefaultBackoffTime),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}
	if cfg.Backoff == nil {
		cfg.Backoff = ConstantBackoff(0)
	}
	if cfg.ErrorTest == nil {
		cfg.ErrorTest = func(err error) bool {
			return true // always retry
		}
	}
	if cfg.OnRetry == nil {
		cfg.OnRetry = func(try uint, err error) {
		}
	}
	ctx := cfg.Context
	var lastErr error
	var tries uint
	for i := uint(0); i <= cfg.Attempts; i++ {
		if i > 0 {
			cfg.OnRetry(tries, lastErr)
		}
		select {
		case <-ctx.Done():
			if lastErr == nil {
				lastErr = ctx.Err()
			}
			return lastErr
		default:
			err := fn()
			tries++
			if err != nil {
				lastErr = err
				if !cfg.ErrorTest(err) {
					return lastErr
				}
				select {
				case <-ctx.Done():
					return lastErr
				case <-time.After(cfg.Backoff(tries)):
					continue
				}
			}
			return nil
		}
	}
	return lastErr
}
