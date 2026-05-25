package logger

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"
)

const instrumentationName = "github.com/aegis-run/aegis/pkg/telemetry/logger"

// newOTELHandler returns an slog.Handler that forwards records to the OpenTelemetry
// Logs SDK via the official otelslog bridge.
func newOTELHandler(provider *log.LoggerProvider) slog.Handler {
	return otelslog.NewHandler(
		instrumentationName,
		otelslog.WithLoggerProvider(provider),
	)
}
