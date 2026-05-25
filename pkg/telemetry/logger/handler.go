package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	active := make([]slog.Handler, 0, len(handlers))
	for _, h := range handlers {
		if h != nil {
			active = append(active, h)
		}
	}
	if len(active) == 1 {
		return active[0]
	}
	return slog.NewMultiHandler(active...)
}

// traceHandler wraps an inner slog.Handler and injects OpenTelemetry
// trace_id and span_id attributes from the context so that log
// records can be correlated with distributed traces in any backend.
//
// Attribute keys "trace_id" and "span_id" follow common backend conventions
// (like Datadog, Grafana, etc.) while the OTel bridge (otelslog) handles
// the native OTel fields (TraceID, SpanID) in the LogRecord itself.
type traceHandler struct {
	inner slog.Handler
}

func newTraceHandler(inner slog.Handler) slog.Handler {
	return &traceHandler{inner: inner}
}

func (t *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.inner.Enabled(ctx, level)
}

func (t *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return t.inner.Handle(ctx, r)
}

func (t *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{inner: t.inner.WithAttrs(attrs)}
}

func (t *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{inner: t.inner.WithGroup(name)}
}
