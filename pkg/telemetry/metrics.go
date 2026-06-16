package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	orn "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

func ConfigureMetrics(ctx context.Context, cfg *Metrics) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return noopShutdown(), nil
	}

	logger.Debug("metrics.enabled",
		slog.String("exporter", cfg.Exporter),
		slog.String("protocol", string(cfg.Protocol)),
		slog.String("endpoint", cfg.Endpoint),
		slog.Duration("interval", cfg.Interval),
	)

	switch cfg.Exporter {
	case "otlp":
		exp, err := otlpMetricsExporter(ctx, cfg)
		if err != nil {
			return noopShutdown(), err
		}

		reader := metric.NewPeriodicReader(exp, metric.WithInterval(cfg.Interval))

		return newMeter(reader), nil

	case "prometheus":
		reader, err := prometheus.New()
		if err != nil {
			return noopShutdown(), err
		}
		return newMeter(reader), nil

	default:
		return noopShutdown(), fmt.Errorf("%s metrics exporter is unsupported", cfg.Exporter)
	}
}

func newMeter(reader metric.Reader) func(context.Context) error {
	res, err := aegisResource()
	if err != nil {
		return noopShutdown()
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(res),
	)

	if err = orn.Start(
		orn.WithMinimumReadMemStatsInterval(time.Second),
		orn.WithMeterProvider(mp),
	); err != nil {
		return noopShutdown()
	}

	if err = host.Start(host.WithMeterProvider(mp)); err != nil {
		return noopShutdown()
	}

	otel.SetMeterProvider(mp)

	return mp.Shutdown
}

func otlpMetricsExporter(ctx context.Context, cfg *Metrics) (metric.Exporter, error) {
	switch cfg.Protocol {
	case ProtocolHTTP:
		return otlpMetricsHTTPExporter(ctx, cfg)
	case ProtocolGRPC:
		return otlpMetricsGRPCExporter(ctx, cfg)
	default:
		return nil, fmt.Errorf("%s protocol is unsupported", cfg.Protocol)
	}
}

func otlpMetricsHTTPExporter(ctx context.Context, cfg *Metrics) (metric.Exporter, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Headers))
	}

	if cfg.URLPath != "" {
		opts = append(opts, otlpmetrichttp.WithURLPath(cfg.URLPath))
	}

	if cfg.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exp, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return exp, nil
}

func otlpMetricsGRPCExporter(ctx context.Context, cfg *Metrics) (metric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.Headers))
	}

	if cfg.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	} else {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exp, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return exp, nil
}
