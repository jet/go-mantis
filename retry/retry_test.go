package retry

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRetryFirstSuccess(t *testing.T) {
	err := Do(func() error {
		return nil
	},
		Attempts(5),
		Delay(ConstantBackoff(0)),
		WithContext(context.Background()),
		OnRetry(func(try uint, err error) {
			t.Fatal("did not expect OnRetry to trigger")
			return
		}),
		ForErrors(func(err error) bool {
			t.Fatal("did not expect ForErrors to trigger")
			return true // always
		}),
	)
	if err != nil {
		t.Fatal("did not expect Do to return an error")
	}
}

func TestRetrySuccessOnRetry(t *testing.T) {
	var errs []error
	i := 2
	err := Do(func() error {
		i--
		if i == 0 {
			return nil
		}
		return fmt.Errorf("i=%d", i)
	},
		Attempts(5),
		Delay(ConstantBackoff(0)),
		WithContext(context.Background()),
		OnRetry(func(try uint, err error) {
			errs = append(errs, err)
			return
		}),
		ForErrors(func(err error) bool {
			return true // always
		}),
	)
	if err != nil {
		t.Fatal("did not expect Do to return an error")
	}
	if len(errs) != 1 {
		t.Fatal("expected 1 retry error")
	}
}

func TestRetryAllFailed(t *testing.T) {
	var i uint
	var errs []error
	attempts := uint(5)
	err := Do(func() error {
		defer func() { i = i + 1 }()
		return fmt.Errorf("i=%d", i)
	},
		Attempts(attempts),
		Delay(ConstantBackoff(0)),
		WithContext(context.Background()),
		OnRetry(func(try uint, err error) {
			t.Logf("retry %d: %v", try, err)
			errs = append(errs, err)
			return
		}),
		ForErrors(func(err error) bool {
			return true // always
		}),
	)
	if err != nil {
		errs = append(errs, err)
	}
	errlen := 6
	if len(errs) != errlen {
		t.Fatalf("expected %d errors, got %d", errlen, len(errs))
	}
}

func TestRetryContextExpired(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	var i uint
	err := Do(func() error {
		defer func() { i = i + 1 }()
		return fmt.Errorf("try %d", i)
	},
		Attempts(5),
		WithContext(ctx),
		Delay(ConstantBackoff(100*time.Millisecond)),
		ForErrors(func(err error) bool { return true }),
	)
	if i != 1 {
		t.Fatal("Expected only 1 attempt since context will expire before delay completes")
	}
	t.Logf("i=%d, err=%v", i, err)
}

func TestRetryContextDone(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	var i uint
	err := Do(func() error {
		defer func() { i = i + 1 }()
		return fmt.Errorf("try %d", i)
	},
		Attempts(5),
		WithContext(ctx),
		Delay(ConstantBackoff(100*time.Millisecond)),
		ForErrors(func(err error) bool { return true }),
	)
	if i != 0 {
		t.Fatal("Expected no attempt since context was expired before the frist attempt")
	}
	if err != ctx.Err() {
		t.Fatal("Expected returned error to be the ctx.Err()")
	}
	t.Logf("i=%d, err=%v", i, err)
}

func TestRetryCheckFailExecutor(t *testing.T) {
	var i uint
	var errs []error
	attempts := uint(5)
	err := Do(func() error {
		defer func() { i = i + 1 }()
		return fmt.Errorf("try %d", i)
	},
		Attempts(attempts),
		Delay(ConstantBackoff(0)),
		WithContext(context.Background()),
		OnRetry(func(try uint, err error) {
			t.Logf("try %d: %v", try, err)
			errs = append(errs, err)
			return
		}),
		ForErrors(func(err error) bool {
			return false // never retry
		}),
	)
	if err != nil {
		errs = append(errs, err)
	}
	errlen := 1
	if len(errs) != errlen {
		t.Fatalf("expected %d errors, got %d", errlen, len(errs))
	}
}
