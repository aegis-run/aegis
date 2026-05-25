// Package runtime provides application lifecycle management with graceful shutdown.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

// TaskFunc is a long-running task that respects context cancellation.
type TaskFunc func(ctx context.Context) error

// ShutdownFunc is called during graceful shutdown.
type ShutdownFunc func(ctx context.Context) error

// CloseFunc is a shutdown function that doesn't need context.
type CloseFunc func() error

// state represents the runtime lifecycle state.
type state uint32

const (
	stateIdle state = iota
	stateRunning
	stateShuttingDown
	stateStopped
)

// Runtime manages application lifecycle with graceful shutdown support.
type Runtime struct {
	mu sync.Mutex
	wg sync.WaitGroup

	cleanups []ShutdownFunc
	state    state

	errCh  chan error
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Runtime instance.
func New() *Runtime {
	ctx, cancel := context.WithCancel(context.Background())

	return &Runtime{
		cleanups: make([]ShutdownFunc, 0),
		errCh:    make(chan error, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// fail triggers shutdown (first error wins)
func (rt *Runtime) fail(err error) {
	select {
	case rt.errCh <- err:
	default:
	}
	rt.cancel()
}

// Recover handles panics in goroutines and logs the stack trace.
func (rt *Runtime) Recover() {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic: %v", r)
		logger.Error("runtime.panic",
			"panic", r,
			"stack", string(debug.Stack()),
		)
		rt.fail(err)
	}
}

// Go starts a background task managed by the runtime.
// Tasks are tracked and waited for during shutdown.
// Panics are recovered and logged. Errors are sent to the error channel.
func (rt *Runtime) Go(fn TaskFunc) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state == stateShuttingDown {
		return
	}

	rt.wg.Go(func() {
		if err := fn(rt.ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("runtime.task_failed",
				"error", err,
			)
			rt.fail(err)
		}
	})
}

// Defer registers cleanup functions to run during shutdown.
// Functions run in reverse order of registration.
func (rt *Runtime) Defer(fns ...CloseFunc) {
	if len(fns) == 0 {
		return
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state == stateShuttingDown {
		return
	}

	for _, fn := range fns {
		rt.cleanups = append(rt.cleanups, func(context.Context) error {
			return fn()
		})
	}
}

// DeferFunc registers context-aware cleanup functions to run during shutdown.
// Functions run in reverse order of registration.
func (rt *Runtime) DeferFunc(fns ...ShutdownFunc) {
	if len(fns) == 0 {
		return
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state == stateShuttingDown {
		return
	}

	rt.cleanups = append(rt.cleanups, fns...)
}
