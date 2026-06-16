package dispatch

import (
	"go.opentelemetry.io/otel/metric"

	"github.com/aegis-run/aegis/internal"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

var CacheHits = telemetry.MustCounter(
	internal.Meter,
	"aegis.dispatch.cache.hits.total",
	metric.WithDescription("Total cache hits during dispatch"),
)

var CacheMisses = telemetry.MustCounter(
	internal.Meter,
	"aegis.dispatch.cache.misses.total",
	metric.WithDescription("Total cache misses during dispatch"),
)

var SingleflightShared = telemetry.MustCounter(
	internal.Meter,
	"aegis.dispatch.singleflight.shared.total",
	metric.WithDescription("Total times a dispatch request joined an existing flight"),
)
