package telemetry

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/pflag"
)

type Protocol string

const (
	ProtocolHTTP Protocol = "http"
	ProtocolGRPC Protocol = "grpc"
)

// Tracing contains configuration for distributed tracing (Wide Events).
type Tracing struct {
	Enabled  bool              `mapstructure:"enabled"`
	Exporter string            `mapstructure:"exporter"`
	Endpoint string            `mapstructure:"endpoint"`
	Insecure bool              `mapstructure:"insecure"`
	URLPath  string            `mapstructure:"url_path"`
	Headers  map[string]string `mapstructure:"headers"`
	Protocol Protocol          `mapstructure:"protocol"`
}

func (c *Tracing) Validate() error {
	if !c.Enabled || c.Exporter == "none" {
		return nil
	}
	if c.Exporter == string(LogExporterOTLP) && c.Endpoint == "" {
		return errors.New("otlp tracing enabled but endpoint is empty")
	}
	return nil
}

func DefaultTracingConfig() Tracing {
	return Tracing{
		Enabled:  false,
		Exporter: "none",
		Endpoint: "localhost:4317",
		Protocol: ProtocolGRPC,
		Insecure: true,
		Headers:  make(map[string]string),
	}
}

func TracingFlags() *pflag.FlagSet {
	f := pflag.NewFlagSet("tracing", pflag.ContinueOnError)

	cfg := DefaultTracingConfig()
	f.Bool("tracing.enabled", cfg.Enabled, "Enable distributed tracing")
	f.String("tracing.exporter", cfg.Exporter, "Tracer exporter (otlp, none)")
	f.String("tracing.endpoint", cfg.Endpoint, "Tracer OTLP endpoint")
	f.Bool("tracing.insecure", cfg.Insecure, "Tracer exporter insecure (no TLS)")
	f.String("tracing.url_path", cfg.URLPath, "Tracer exporter URL path")
	f.StringToString("tracing.headers", cfg.Headers, "Tracer exporter headers")
	f.String("tracing.protocol", string(cfg.Protocol), "Tracer exporter protocol (http, https)")

	return f
}

// Metrics contains configuration for metrics collection.
type Metrics struct {
	Enabled  bool              `mapstructure:"enabled"`
	Exporter string            `mapstructure:"exporter"` // otlp, prometheus, none
	Endpoint string            `mapstructure:"endpoint"`
	Insecure bool              `mapstructure:"insecure"`
	URLPath  string            `mapstructure:"url_path"` // Default /v1/metrics
	Headers  map[string]string `mapstructure:"headers"`
	Protocol Protocol          `mapstructure:"protocol"`
	Interval time.Duration     `mapstructure:"interval"`
}

func (c *Metrics) Validate() error {
	if !c.Enabled || c.Exporter == "none" {
		return nil
	}
	if c.Exporter == string(LogExporterOTLP) && c.Endpoint == "" {
		return errors.New("otlp metrics enabled but endpoint is empty")
	}
	if c.Interval <= 0 {
		return errors.New("metrics interval must be positive")
	}
	return nil
}

func DefaultMetricsConfig() Metrics {
	return Metrics{
		Enabled:  false,
		Exporter: "none",
		Endpoint: "localhost:4317",
		Protocol: ProtocolGRPC,
		Insecure: true,
		Headers:  make(map[string]string),
		Interval: 1 * time.Minute,
	}
}

func MetricsFlags() *pflag.FlagSet {
	f := pflag.NewFlagSet("metrics", pflag.ContinueOnError)

	cfg := DefaultMetricsConfig()
	f.Bool("metrics.enabled", cfg.Enabled, "Enable metrics collection")
	f.String("metrics.exporter", cfg.Exporter, "Metrics exporter (otlp, none)")
	f.String("metrics.endpoint", cfg.Endpoint, "Metrics OTLP endpoint")
	f.Bool("metrics.insecure", cfg.Insecure, "Metrics exporter insecure (no TLS)")
	f.String("metrics.url_path", cfg.URLPath, "Metrics exporter URL path")
	f.StringToString("metrics.headers", cfg.Headers, "Metrics exporter headers")
	f.String("metrics.protocol", string(cfg.Protocol), "Metrics exporter protocol (http, https)")
	f.Duration("metrics.interval", cfg.Interval, "Metrics collection interval")

	return f
}

// LogFormat controls the log output format.
type LogFormat string

const (
	// LogFormatJSON - JSON (default)
	LogFormatJSON LogFormat = "json"
	// LogFormatText - human-readable text
	LogFormatText LogFormat = "text"
)

type LogExporter string

const (
	LogExporterStdout LogExporter = "stdout"
	LogExporterOTLP   LogExporter = "otlp"
)

// LogExporterConfig holds the configuration for a single log exporter.
type LogExporterConfig struct {
	Type     LogExporter       `mapstructure:"type"`
	Endpoint string            `mapstructure:"endpoint"`
	Insecure bool              `mapstructure:"insecure"`
	URLPath  string            `mapstructure:"url_path"`
	Headers  map[string]string `mapstructure:"headers"`
	Protocol Protocol          `mapstructure:"protocol"`
}

func (c *LogExporterConfig) Validate() error {
	if c.Type == LogExporterOTLP && c.Endpoint == "" {
		return errors.New("otlp log exporter enabled but endpoint is empty")
	}
	return nil
}

// Logs contains configuration for structured logging.
type Logs struct {
	Enabled   bool                `mapstructure:"enabled"`
	Level     slog.Level          `mapstructure:"level"`
	Format    LogFormat           `mapstructure:"format"`
	Exporters []LogExporterConfig `mapstructure:"exporters"`
}

func (c *Logs) Validate() error {
	if !c.Enabled {
		return nil
	}
	var errs []error
	for i, exp := range c.Exporters {
		if err := exp.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("exporter[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

func DefaultLogsConfig() Logs {
	return Logs{
		Enabled: true,
		Level:   slog.LevelInfo,
		Format:  LogFormatJSON,
		Exporters: []LogExporterConfig{
			{Type: LogExporterStdout},
		},
	}
}

func LogsFlags() *pflag.FlagSet {
	f := pflag.NewFlagSet("logs", pflag.ContinueOnError)
	cfg := DefaultLogsConfig()

	f.Bool("logs.enabled", cfg.Enabled, "Enable structured logging")
	f.String("logs.level", cfg.Level.String(), "Log level (debug, info, warn, error)")
	f.String("logs.format", string(cfg.Format), "Log format (text, json)")

	return f
}
