package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/aegis-run/aegis/internal"
)

// Start wraps the global tracer to start a new span.
func Start(
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return internal.Trace.Start(ctx, name, opts...)
}

func End(span trace.Span, err *error) {
	if err != nil && *err != nil {
		// code := errkit.GetCode(*err)
		// if !code.IsZero() {
		// 	span.SetAttributes(attribute.String("mithras.error.code", code.String()))
		// }

		span.RecordError(*err)
		span.SetStatus(codes.Error, (*err).Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}

// Attr adds key-value attributes to the canonical wide event.
func Attr(ctx context.Context, attrs ...attribute.KeyValue) {
	span(ctx).SetAttributes(attrs...)
}

// Event adds a discrete event (log point) to the canonical wide event.
func Event(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span(ctx).AddEvent(name, trace.WithAttributes(attrs...))
}

// Ok marks the canonical wide event as successful.
func Ok(ctx context.Context) {
	span(ctx).SetStatus(codes.Ok, "")
}

// Err records an error on the canonical wide event and extracts errkit codes.
func Err(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	sp := span(ctx)

	// code := errkit.GetCode(err)
	// if !code.IsZero() {
	// 	sp.SetAttributes(attribute.String("mithras.error.code", code.String()))
	// }

	sp.RecordError(err)
	sp.SetStatus(codes.Error, err.Error())

	return err
}
