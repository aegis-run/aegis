// Package logger provides wide-event structured logging with OpenTelemetry
// trace correlation support.
package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
)

type logState struct {
	logger  *slog.Logger
	enabled bool
}

const (
	// LevelFatal is the architectural severity for terminal failures.
	LevelFatal slog.Level = 12
)

var (
	mu    sync.Mutex
	state atomic.Pointer[logState]
)

func init() {
	Configure(DefaultConfig())
}

func Logger() *slog.Logger {
	return state.Load().logger
}

// Configure applies cfg to the global logger. It is safe to call concurrently.
func Configure(cfg *Config) {
	mu.Lock()
	defer mu.Unlock()

	opts := &slog.HandlerOptions{Level: cfg.Level}

	var handlers []slog.Handler
	for _, entry := range cfg.Handlers {
		switch entry.Exporter {
		case ExporterStdout:
			if cfg.Format == FormatText {
				handlers = append(handlers, slog.NewTextHandler(os.Stdout, opts))
			} else {
				handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))
			}
		case ExporterOTLP:
			if entry.LoggerProvider != nil {
				handlers = append(handlers, newOTELHandler(entry.LoggerProvider))
			}
		}
	}

	// Fallback to stdout JSON if no handlers specified.
	if len(handlers) == 0 {
		handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))
	}

	// Always wrap in traceHandler for correlation.
	l := slog.New(newTraceHandler(newMultiHandler(handlers...)))
	slog.SetDefault(l)

	state.Store(&logState{
		logger:  l,
		enabled: cfg.Enabled,
	})
}

// AddBaseAttrs adds base attributes to every subsequent log record.
func AddBaseAttrs(attrs ...slog.Attr) {
	mu.Lock()
	defer mu.Unlock()

	curr := state.Load()
	newL := slog.New(curr.logger.Handler().WithAttrs(attrs))
	slog.SetDefault(newL)

	state.Store(&logState{
		logger:  newL,
		enabled: curr.enabled,
	})
}

// SetHandler explicitly overrides the underlying slog.Handler.
func SetHandler(h slog.Handler) {
	mu.Lock()
	defer mu.Unlock()

	curr := state.Load()
	newL := slog.New(h)
	slog.SetDefault(newL)

	state.Store(&logState{
		logger:  newL,
		enabled: curr.enabled,
	})
}

// Debug emits a debug-level message.
func Debug(msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}

	s.logger.Debug(msg, args...)
}

// Info emits an info-level message.
func Info(msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.Info(msg, args...)
}

// Warn emits a warning-level message.
func Warn(msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.Warn(msg, args...)
}

// Error emits an error-level message.
func Error(msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.Error(msg, args...)
}

// Fatal emits a fatal-level message and exits the program with status 1.
func Fatal(msg string, args ...any) {
	s := state.Load()
	if s.enabled {
		s.logger.Log(context.Background(), LevelFatal, msg, args...)
	}
	os.Exit(1)
}

// FatalContext emits a fatal-level message, passing ctx for trace correlation, and exits.
func FatalContext(ctx context.Context, msg string, args ...any) {
	s := state.Load()
	if s.enabled {
		s.logger.Log(ctx, LevelFatal, msg, args...)
	}
	os.Exit(1)
}

// Log logs a message with the given level and arguments.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	state.Load().logger.Log(ctx, level, msg, args...)
}

// DebugContext emits a debug-level message, passing ctx for trace correlation.
func DebugContext(ctx context.Context, msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.DebugContext(ctx, msg, args...)
}

// InfoContext emits an info-level message, passing ctx for trace correlation.
func InfoContext(ctx context.Context, msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.InfoContext(ctx, msg, args...)
}

// WarnContext emits a warning-level message, passing ctx for trace correlation.
func WarnContext(ctx context.Context, msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.WarnContext(ctx, msg, args...)
}

// ErrorContext emits an error-level message, passing ctx for trace correlation.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	s := state.Load()
	if !s.enabled {
		return
	}
	s.logger.ErrorContext(ctx, msg, args...)
}
