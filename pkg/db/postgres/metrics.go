package postgres

import (
	"github.com/aegis-run/aegis/internal"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var opsLatency = telemetry.MustHistogram(internal.Meter, "aegis.database.operations.latency",
	metric.WithUnit("ms"),
	metric.WithDescription("Histogram of database operation latencies in milliseconds."),
	metric.WithExplicitBucketBoundaries(1, 2, 5, 10, 20, 50, 100, 200, 300, 400, 500, 750, 1000, 2000, 3000, 5000, 10000),
)

var opsTotal = telemetry.MustCounter(internal.Meter, "aegis.database.operations.total",
	metric.WithDescription("Total number of database operations processed."),
)

var errsTotal = telemetry.MustCounter(internal.Meter, "aegis.database.errors.total",
	metric.WithDescription("Total number of database errors encountered."),
)
