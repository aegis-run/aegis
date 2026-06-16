package check

import (
	"go.opentelemetry.io/otel/metric"

	"github.com/aegis-run/aegis/internal"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

var checkDuration = telemetry.MustHistogram(
	internal.Meter,
	"aegis.engine.check.duration",
	metric.WithUnit("ms"),
	metric.WithDescription("Histogram of check resolution latency in milliseconds"),
	metric.WithExplicitBucketBoundaries(
		1, 2, 5, 10, 20, 50, 100, 200, 300,
		400, 500, 750, 1000, 2000, 3000, 5000, 10000,
	),
)

var checkRequests = telemetry.MustCounter(
	internal.Meter,
	"aegis.engine.check.requests.total",
	metric.WithDescription("Total check requests"),
)

var shortCircuits = telemetry.MustCounter(
	internal.Meter,
	"aegis.engine.short_circuit.total",
	metric.WithDescription("Total times an errgroup short-circuited"),
)

var tuplesFetched = telemetry.MustCounter(
	internal.Meter,
	"aegis.engine.tuples_fetched.total",
	metric.WithDescription("Total number of database tuples fetched during a check"),
)

var fanoutSize = telemetry.MustHistogram(
	internal.Meter,
	"aegis.engine.fanout.size",
	metric.WithDescription("Number of sub-problems (fanout) evaluated per group"),
)
