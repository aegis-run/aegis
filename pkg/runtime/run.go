package runtime

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

var (
	ErrAlreadyRunning = errors.New("runtime already running")
	ErrCannotReuse    = errors.New("runtime cannot be reused")
)

// RunConfig holds runtime configuration options.
type RunConfig struct {
	Signals []os.Signal
	Timeout time.Duration
}

var defaultCfg = RunConfig{
	Timeout: 30 * time.Second,
	Signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
}

// RunOption configures RunConfig.
type RunOption func(*RunConfig)

// WithTimeout sets the shutdown timeout.
func WithTimeout(d time.Duration) RunOption {
	return func(c *RunConfig) {
		c.Timeout = d
	}
}

// WithSignals sets the OS signals to listen for.
func WithSignals(sigs ...os.Signal) RunOption {
	return func(c *RunConfig) {
		c.Signals = sigs
	}
}

// Run starts the runtime and blocks until shutdown.
// It listens for signals, context cancellation, or task errors.
// Returns an error if shutdown failed or a task errored.
func (rt *Runtime) Run(ctx context.Context, opts ...RunOption) error {
	cfg := defaultCfg
	for _, o := range opts {
		o(&cfg)
	}

	if err := rt.start(cfg); err != nil {
		return err
	}

	err := rt.wait(ctx, cfg)

	return rt.shutdown(err, cfg)
}

func (rt *Runtime) start(cfg RunConfig) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	switch rt.state {
	case stateIdle:
		rt.state = stateRunning
	case stateRunning:
		return ErrAlreadyRunning
	case stateShuttingDown, stateStopped:
		return ErrCannotReuse
	default:
		return ErrCannotReuse
	}

	logger.Debug("runtime.started",
		"timeout", cfg.Timeout,
	)
	return nil
}

func (rt *Runtime) wait(ctx context.Context, cfg RunConfig) (err error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, cfg.Signals...)
	defer signal.Stop(sigCh)

	reason := ""
	select {
	case sig := <-sigCh:
		reason = "signal: " + sig.String()
		logger.Info("runtime.shutdown_signal",
			"signal", sig.String(),
		)
	case <-ctx.Done():
		reason = "context canceled"
		logger.Info("runner.context_canceled",
			"err", ctx.Err(),
		)
	case err = <-rt.errCh:
		reason = "task error"
	}

	if err != nil {
		logger.Error("runtime.shutdown_with_error",
			"error", err,
		)
	} else {
		logger.Debug("runtime.shutting_down",
			"reason", reason,
		)
	}

	return err
}

//nolint:contextcheck // shutdown context must not inherit from canceled runtime context
func (rt *Runtime) shutdown(cause error, cfg RunConfig) error {
	rt.mu.Lock()
	rt.state = stateShuttingDown
	cleanups := slices.Clone(rt.cleanups)
	rt.mu.Unlock()

	rt.cancel()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.WithoutCancel(rt.ctx), cfg.Timeout)
	defer cancelShutdown()

	var errs []error
	if cause != nil {
		errs = append(errs, cause)
	}

	errs = rt.cleanup(shutdownCtx, cleanups, errs)
	errs = rt.waitForTasks(shutdownCtx, errs)

	rt.state = stateStopped

	if len(errs) > 0 {
		finalErr := errors.Join(errs...)
		logger.Error("runtime.shutdown_completed_with_errors",
			"error", finalErr,
		)
		return finalErr
	}

	logger.Debug("runtime.shutdown_complete")
	return nil
}

func (rt *Runtime) cleanup(ctx context.Context, cleanups []ShutdownFunc, errs []error) []error {
	for _, cl := range slices.Backward(cleanups) {
		if cl == nil {
			continue
		}

		if err := cl(ctx); err != nil {
			logger.Error("runtime.cleanup_failed",
				"error", err,
			)
			errs = append(errs, err)
		}
	}
	return errs
}

func (rt *Runtime) waitForTasks(ctx context.Context, errs []error) []error {
	done := make(chan struct{})
	go func() {
		rt.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("runtime.all_tasks_exited")
	case <-ctx.Done():
		logger.Error("runtime.shutdown_timeout")
		errs = append(errs, context.DeadlineExceeded)
	}
	return errs
}
