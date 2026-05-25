package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aegis-run/aegis/pkg/assert"
)

func TestRuntime_Run(t *testing.T) {
	t.Run("ExecutesTasks", func(t *testing.T) {
		tasksCount := 5

		rt := New()
		ctx, cancel := context.WithCancel(t.Context())

		started := make(chan struct{}, tasksCount)
		for range tasksCount {
			rt.Go(func(ctx context.Context) error {
				started <- struct{}{}
				<-ctx.Done()
				return nil
			})
		}

		done := make(chan error, 1)
		go func() {
			done <- rt.Run(ctx, WithTimeout(100*time.Millisecond))
		}()

		for i := range tasksCount {
			select {
			case <-started:
			case <-t.Context().Done():
				t.Fatalf("test timed out waiting for task %d to start", i+1)
			}
		}

		cancel()

		select {
		case err := <-done:
			assert.Err(t, err, nil)
		case <-t.Context().Done():
			t.Fatal("test timed out waiting for Run to exit")
		}
	})

	t.Run("StopsOnContextCancel", func(t *testing.T) {
		rt := New()
		ctx, cancel := context.WithCancel(t.Context())

		taskStopped := make(chan struct{})
		rt.Go(func(ctx context.Context) error {
			<-ctx.Done()
			close(taskStopped)
			return nil
		})

		done := make(chan error, 1)
		go func() {
			done <- rt.Run(ctx, WithTimeout(100*time.Millisecond))
		}()

		cancel()

		select {
		case <-taskStopped:
		case <-time.After(time.Second):
			t.Fatal("task should stop after context cancel")
		}

		err := <-done
		assert.Err(t, err, nil)
	})

	t.Run("StopsOnTaskError", func(t *testing.T) {
		rt := New()

		taskErr := errors.New("task failed")
		rt.Go(func(_ context.Context) error {
			return taskErr
		})

		taskStopped := make(chan struct{})
		rt.Go(func(ctx context.Context) error {
			<-ctx.Done()
			close(taskStopped)
			return nil
		})

		err := rt.Run(t.Context(), WithTimeout(time.Second))
		assert.Err(t, err, taskErr)

		select {
		case <-taskStopped:
		case <-time.After(time.Second):
			t.Fatal("other task should stop after first task error")
		}
	})

	t.Run("CanOnlyRunOnce", func(t *testing.T) {
		rt := New()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- rt.Run(ctx, WithTimeout(100*time.Millisecond))
		}()

		time.Sleep(20 * time.Millisecond)

		err := rt.Run(ctx, WithTimeout(100*time.Millisecond))
		assert.Err(t, err, ErrAlreadyRunning)

		cancel()

		select {
		case err := <-done:
			assert.Err(t, err, nil)
		case <-t.Context().Done():
			t.Fatal("test timed out waiting for Run to exit")
		}
	})

	t.Run("EmptyRunner", func(t *testing.T) {
		rt := New()
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := rt.Run(ctx, WithTimeout(time.Second))
		assert.Err(t, err, nil)
	})

	t.Run("ContinuesOnCleanupError", func(t *testing.T) {
		rt := New()

		var called []int
		var mu sync.Mutex

		rt.Defer(func() error {
			mu.Lock()
			called = append(called, 1)
			mu.Unlock()
			return nil
		})

		rt.Defer(func() error {
			mu.Lock()
			called = append(called, 2)
			mu.Unlock()
			return errors.New("cleanup 2 failed")
		})

		rt.Defer(func() error {
			mu.Lock()
			called = append(called, 3)
			mu.Unlock()
			return nil
		})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := rt.Run(ctx, WithTimeout(100*time.Millisecond))
		assert.True(t, err != nil)

		mu.Lock()
		assert.Equal(t, []int{3, 2, 1}, called)
		mu.Unlock()
	})

	t.Run("WithTimeout_SetsTimeout", func(t *testing.T) {
		rt := New()

		var cleanupCtx context.Context
		rt.DeferFunc(func(ctx context.Context) error {
			cleanupCtx = ctx
			time.Sleep(200 * time.Millisecond)
			return ctx.Err()
		})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := rt.Run(ctx, WithTimeout(50*time.Millisecond))
		assert.Err(t, err, context.DeadlineExceeded)

		deadline, ok := cleanupCtx.Deadline()
		assert.True(t, ok)
		assert.Equal(t, deadline.IsZero(), false)
	})
}
