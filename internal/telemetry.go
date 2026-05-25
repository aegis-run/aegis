package internal

import "go.opentelemetry.io/otel"

var (
	Trace = otel.Tracer("aegis")
	Meter = otel.Meter("aegis")
)
