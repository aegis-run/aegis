package cli

import (
	"context"
	"log/slog"
)

// SlogHandler implements slog.Handler to route structured traces seamlessly into the CLI
// Output renderer.
type SlogHandler struct {
	out *Output
}

// NewSlogHandler creates a natively mapped handler wrapping the provided CLI interface.
func NewSlogHandler(out *Output) *SlogHandler {
	return &SlogHandler{out: out}
}

// Enabled reports whether the handler handles natively structured records at the given
// architectural severity.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	// Info level and above are always permitted to the CLI.
	if level >= slog.LevelInfo {
		return true
	}

	// Debug level is only permitted if verbose mode is explicitly active.
	return h.out.IsVerbose()
}

// Handle rigidly formats and colorizes the slog.Record mapping natively to the UI.
func (h *SlogHandler) Handle(_ context.Context, r slog.Record) error {
	// Format the primary message explicitly based on severity level
	switch {
	case r.Level >= 12: // LevelFatal
		h.out.Error("%s", r.Message)
	case r.Level >= slog.LevelError:
		h.out.Error("%s", r.Message)
	case r.Level >= slog.LevelWarn:
		h.out.Warn("%s", r.Message)
	case r.Level >= slog.LevelInfo:
		h.out.Info("%s", r.Message)
	case r.Level >= slog.LevelDebug:
		h.out.Verbose("%s", r.Message)
	default:
		h.out.Raw("%s", r.Message)
	}

	// If there are attached attributes dynamically attached to the trace, gracefully render
	// them natively beneath
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			h.out.Subtle("%s=%v", a.Key, a.Value.Any())
			return true
		})
	}

	return nil
}

// WithAttrs returns a safely copied Handler with structural attributes.
func (h *SlogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// For CLI UI implementations we dynamically ignore persisting base attributes natively
	// to prevent console clutter.
	return h
}

// WithGroup seamlessly returns a statically structured group without crashing string bounds.
func (h *SlogHandler) WithGroup(_ string) slog.Handler {
	return h
}
