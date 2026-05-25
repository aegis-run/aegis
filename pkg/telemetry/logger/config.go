package logger

import (
	"log/slog"

	"go.opentelemetry.io/otel/sdk/log"
)

// Format controls the encoding used when writing to stdout.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Exporter names a sink type.
type Exporter string

const (
	ExporterStdout Exporter = "stdout"
	ExporterOTLP   Exporter = "otlp"
)

// HandlerEntry describes one sink in a multi-exporter setup.
type HandlerEntry struct {
	Exporter       Exporter
	LoggerProvider *log.LoggerProvider
}

// Config holds all logger configuration.
type Config struct {
	// Enabled gates all log emission. When false every log call is a no-op.
	Enabled bool

	// Level is the minimum severity that is emitted.
	Level slog.Level

	// Format selects text vs JSON encoding for stdout handlers.
	Format Format

	// Handlers lists the active sinks.
	Handlers []HandlerEntry
}

// DefaultConfig returns a safe out-of-the-box configuration that logs JSON to
// stdout at INFO level.
func DefaultConfig() *Config {
	return &Config{
		Enabled:  true,
		Level:    slog.LevelInfo,
		Format:   FormatJSON,
		Handlers: []HandlerEntry{{Exporter: ExporterStdout}},
	}
}
