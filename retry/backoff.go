package retry

import (
	"math/rand"
	"time"
)

// BackoffFunc Computes the delay between retries on the given current retry number
// It may return 0 if try == 0, but try should always be called with try > 0
type BackoffFunc func(try uint) time.Duration

// ConstantBackoff always returns the same duration
func ConstantBackoff(delay time.Duration) BackoffFunc {
	return func(try uint) time.Duration {
		return delay
	}
}

// ExpontentialRandomBackoff implements Exponential Random Backoff
// See: https://en.wikipedia.org/wiki/Exponential_backoff
//
//        Takes a delay (w) and a maxExp (E)
//        Given: Try (n)
//        Then:  Delay (d) = w * RAND[0 to (2^MIN(n,E) - 1)]
func ExpontentialRandomBackoff(delay time.Duration, maxExp uint) BackoffFunc {
	return func(n uint) time.Duration {
		if n == 0 {
			return time.Duration(0)
		}
		slots := 2 << (maxExp - 1)
		if n < maxExp {
			slots = (2 << (n - 1))
		}
		return delay * time.Duration(rand.Intn(slots))
	}
}
