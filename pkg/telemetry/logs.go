package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc/credentials"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

// ConfigureLogs wires up all configured log exporters and applies them to the
// global logger. It returns a shutdown function that flushes and stops the
// OTel log provider.
func ConfigureLogs(ctx context.Context, cfg *Logs) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return noopShutdown(), nil
	}

	logger.Debug("logs.enabled",
		slog.String("level", cfg.Level.String()),
		slog.String("format", string(cfg.Format)),
		slog.Int("exporters", len(cfg.Exporters)),
	)

	handlers, shutdown, err := buildLogHandlers(ctx, cfg)
	if err != nil {
		return noopShutdown(), err
	}

	logger.Configure(&logger.Config{
		Enabled:  cfg.Enabled,
		Level:    cfg.Level,
		Format:   logger.Format(cfg.Format),
		Handlers: handlers,
	})

	return shutdown, nil
}

func buildLogHandlers(
	ctx context.Context,
	cfg *Logs,
) ([]logger.HandlerEntry, func(context.Context) error, error) {
	var (
		handlers   []logger.HandlerEntry
		processors []log.Processor
	)

	for _, exp := range cfg.Exporters {
		logger.Debug("logs.exporter",
			slog.String("type", string(exp.Type)),
			slog.String("protocol", string(exp.Protocol)),
			slog.String("endpoint", exp.Endpoint),
		)

		switch exp.Type {
		case LogExporterStdout:
			handlers = append(handlers, logger.HandlerEntry{
				Exporter: logger.ExporterStdout,
			})

		case LogExporterOTLP:
			sdkExp, err := otlpLogsExporter(ctx, &exp)
			if err != nil {
				return nil, noopShutdown(), fmt.Errorf("logs: build otlp exporter: %w", err)
			}
			processors = append(processors, log.NewBatchProcessor(sdkExp))

		default:
			return nil, noopShutdown(), fmt.Errorf("logs: unsupported exporter %q", exp.Type)
		}
	}

	if len(processors) == 0 {
		return handlers, noopShutdown(), nil
	}

	lp, err := newLogProvider(processors)
	if err != nil {
		return nil, noopShutdown(), err
	}

	handlers = append(handlers, logger.HandlerEntry{
		Exporter:       logger.ExporterOTLP,
		LoggerProvider: lp,
	})

	return handlers, lp.Shutdown, nil
}

func newLogProvider(processors []log.Processor) (*log.LoggerProvider, error) {
	res, err := aegisResource()
	if err != nil {
		return nil, fmt.Errorf("logs: build resource: %w", err)
	}

	opts := []log.LoggerProviderOption{log.WithResource(res)}
	for _, p := range processors {
		opts = append(opts, log.WithProcessor(p))
	}

	return log.NewLoggerProvider(opts...), nil
}

func otlpLogsExporter(ctx context.Context, cfg *LogExporterConfig) (log.Exporter, error) {
	switch cfg.Protocol {
	case ProtocolHTTP:
		return otlpLogsHTTPExporter(ctx, cfg)
	case ProtocolGRPC:
		return otlpLogsGRPCExporter(ctx, cfg)
	default:
		return nil, fmt.Errorf("%s protocol is unsupported", cfg.Protocol)
	}
}

func otlpLogsHTTPExporter(ctx context.Context, cfg *LogExporterConfig) (log.Exporter, error) {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.Endpoint),
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(cfg.Headers))
	}
	if cfg.URLPath != "" {
		opts = append(opts, otlploghttp.WithURLPath(cfg.URLPath))
	}
	if cfg.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	return otlploghttp.New(ctx, opts...)
}

func otlpLogsGRPCExporter(ctx context.Context, cfg *LogExporterConfig) (log.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.Endpoint),
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
	}
	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	} else {
		opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}
	return otlploggrpc.New(ctx, opts...)
}
