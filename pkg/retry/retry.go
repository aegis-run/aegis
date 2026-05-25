package retry

import (
	"context"
	"errors"
	"time"
)

type Policy struct {
	shouldRetry func(error) bool
	backoff     BackoffFunc
	attempts    int
}

func New(opts ...Option) *Policy {
	pol := &Policy{
		attempts:    3,
		backoff:     LinBackoff(100 * time.Millisecond),
		shouldRetry: func(error) bool { return true },
	}

	for _, opt := range opts {
		opt(pol)
	}

	return pol
}

type Option func(c *Policy)

func Attempts(attempts int) Option {
	return func(c *Policy) { c.attempts = attempts }
}

func Backoff(backoff BackoffFunc) Option {
	return func(c *Policy) { c.backoff = backoff }
}

func ShouldRetry(shouldRetry func(error) bool) Option {
	return func(c *Policy) { c.shouldRetry = shouldRetry }
}

func DoResult[T any](
	ctx context.Context,
	pol *Policy,
	fn func(context.Context) (T, error),
) (t T, err error) {
	if pol == nil {
		pol = New()
	}

	if pol.attempts < 1 {
		return t, errors.New("attempts must be greater than 0")
	}

	for i := 1; i <= pol.attempts; i++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return t, ctxErr
		}

		t, err = fn(ctx)
		if err == nil {
			return t, nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return t, err
		}

		if !pol.shouldRetry(err) {
			return t, err
		}

		if i == pol.attempts {
			break
		}

		delay := pol.backoff(i)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return t, ctx.Err()
		case <-timer.C:
		}
	}

	return t, err
}

func Do(ctx context.Context, p *Policy, fn func(context.Context) error) error {
	_, err := DoResult(ctx, p, func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})

	return err
}
