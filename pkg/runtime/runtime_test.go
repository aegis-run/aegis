package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aegis-run/aegis/pkg/assert"
)

func TestNew(t *testing.T) {
	rt := New()
	assert.True(t, rt != nil)
	assert.Equal(t, len(rt.cleanups), 0)
	assert.Equal(t, rt.state, stateIdle)
	assert.True(t, rt.ctx != nil)
	assert.True(t, rt.cancel != nil)
}

func TestRuntime_Go(t *testing.T) {
	t.Run("StartsImmediately", func(t *testing.T) {
		rt := New()
		ctx, cancel := context.WithCancel(t.Context())

		started := make(chan struct{})
		rt.Go(func(ctx context.Context) error {
			started <- struct{}{}
			<-ctx.Done()
			return nil
		})

		done := make(chan error, 1)
		go func() {
			done <- rt.Run(ctx, WithTimeout(100*time.Millisecond))
		}()

		select {
		case <-started:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("task did not start immediately")
		}

		cancel()

		select {
		case err := <-done:
			assert.Err(t, err, nil)
		case <-t.Context().Done():
			t.Fatal("test timed out waiting for Run to exit")
		}
	})

	t.Run("CapturesError", func(t *testing.T) {
		rt := New()

		testErr := errors.New("boom")
		rt.Go(func(context.Context) error {
			return testErr
		})

		err := rt.Run(t.Context(), WithTimeout(100*time.Millisecond))
		assert.Err(t, err, testErr)
	})
}

func TestRuntime_Defer(t *testing.T) {
	t.Run("ExecutesInReverseOrder", func(t *testing.T) {
		rt := New()

		var order []int

		rt.Defer(func() error {
			order = append(order, 1)
			return nil
		})
		rt.Defer(func() error {
			order = append(order, 2)
			return nil
		})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := rt.Run(ctx)
		assert.Err(t, err, nil)
		assert.Equal(t, order, []int{2, 1})
	})

	t.Run("HandlesCleanupErrors", func(t *testing.T) {
		rt := New()
		cleanupErr := errors.New("cleanup failed")

		rt.Defer(func() error {
			return cleanupErr
		})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := rt.Run(ctx)
		assert.Err(t, err, cleanupErr)
	})
}
