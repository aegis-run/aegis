package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

func ConfigureTracing(ctx context.Context, cfg *Tracing) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return noopShutdown(), nil
	}

	logger.Debug("tracing.enabled",
		slog.String("exporter", cfg.Exporter),
		slog.String("protocol", string(cfg.Protocol)),
		slog.String("endpoint", cfg.Endpoint),
	)

	switch cfg.Exporter {
	case "otlp":
		exp, err := otlpTracingExporter(ctx, cfg)
		if err != nil {
			return noopShutdown(), err
		}

		return newTracer(exp), nil
	default:
		return noopShutdown(), fmt.Errorf("%s tracing exporter is unsupported", cfg.Exporter)
	}
}

func newTracer(exporter trace.SpanExporter) func(context.Context) error {
	res, err := aegisResource()
	if err != nil {
		return noopShutdown()
	}

	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(
			trace.NewBatchSpanProcessor(exporter),
		),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown
}

func otlpTracingExporter(ctx context.Context, cfg *Tracing) (trace.SpanExporter, error) {
	switch cfg.Protocol {
	case ProtocolHTTP:
		return otlpTracingHTTPExporter(ctx, cfg)
	case ProtocolGRPC:
		return otlpTracingGRPCExporter(ctx, cfg)
	default:
		return nil, fmt.Errorf("%s protocol is unsupported", cfg.Protocol)
	}
}

func otlpTracingHTTPExporter(ctx context.Context, cfg *Tracing) (trace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	if cfg.URLPath != "" {
		opts = append(opts, otlptracehttp.WithURLPath(cfg.URLPath))
	}

	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptracehttp.New(ctx, opts...)
}

func otlpTracingGRPCExporter(ctx context.Context, cfg *Tracing) (trace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	} else {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	return otlptracegrpc.New(ctx, opts...)
}
