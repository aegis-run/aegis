package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type contextKey struct{}

var mainSpanKey = contextKey{}

// InjectMainSpan designates the provided span as the canonical wide event.
func InjectMainSpan(ctx context.Context, span trace.Span) context.Context {
	span.SetAttributes(attribute.Bool("main", true))
	return context.WithValue(ctx, mainSpanKey, span)
}

// span pulls the canonical main span if present, falling back to the active span.
func span(ctx context.Context) trace.Span {
	if s, ok := ctx.Value(mainSpanKey).(trace.Span); ok {
		return s
	}

	return trace.SpanFromContext(ctx)
}
