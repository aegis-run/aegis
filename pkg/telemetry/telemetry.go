package telemetry

import (
	"context"
	"log/slog"
	"os"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	"github.com/aegis-run/aegis/internal"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

func aegisResource() (*resource.Resource, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("aegis"),
		attribute.String("id", internal.Identifier),
		attribute.String("version", internal.Version),
		attribute.String("host", host),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
	), nil
}

func noopShutdown() func(context.Context) error {
	return func(context.Context) error { return nil }
}

func NewCounter(meter metric.Meter, name, description string) metric.Int64Counter {
	counter, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		logger.Error("telemetry.NewCounter.failed",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		panic(err)
	}

	return counter
}

func NewHistogram(meter metric.Meter, name, unit, description string) metric.Int64Histogram {
	histogram, err := meter.Int64Histogram(name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		logger.Error("telemetry.NewHistogram.failed",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		panic(err)
	}

	return histogram
}

func MustHistogram(
	meter metric.Meter,
	name string,
	opts ...metric.Int64HistogramOption,
) metric.Int64Histogram {
	hist, err := meter.Int64Histogram(name, opts...)
	if err != nil {
		logger.Error("telemetry.initHistogram.failed",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		panic(err)
	}

	return hist
}

func MustCounter(
	meter metric.Meter,
	name string,
	opts ...metric.Int64CounterOption,
) metric.Int64Counter {
	counter, err := meter.Int64Counter(name, opts...)
	if err != nil {
		logger.Error("telemetry.initCounter.failed",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		panic(err)
	}

	return counter
}
