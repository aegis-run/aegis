package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/aegis-run/aegis/pkg/telemetry"
)

const (
	statusError   = "error"
	statusSuccess = "success"
)

type Replica struct {
	pool *pgxpool.Pool
	mode string
}

var _ (DBTX) = (*Replica)(nil)

func (r *Replica) Pool() *pgxpool.Pool { return r.pool }

func (r *Replica) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	ctx, span := telemetry.Start(ctx, "database.postgres.replica.exec", trace.WithAttributes(
		attribute.String("mode", r.mode),
		attribute.String("query", query),
	))
	defer span.End()

	start := time.Now()
	res, err := r.pool.Exec(ctx, query, args...)
	observe(ctx, span, r.mode, "Exec", start, err)

	return res, err
}

func (r *Replica) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	ctx, span := telemetry.Start(ctx, "database.postgres.replica.query", trace.WithAttributes(
		attribute.String("mode", r.mode),
		attribute.String("query", query),
	))
	defer span.End()

	start := time.Now()
	res, err := r.pool.Query(ctx, query, args...)
	observe(ctx, span, r.mode, "Query", start, err)

	return res, err
}

func (r *Replica) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	ctx, span := telemetry.Start(ctx, "database.postgres.replica.query_row", trace.WithAttributes(
		attribute.String("mode", r.mode),
		attribute.String("query", query),
	))
	defer span.End()

	start := time.Now()
	res := r.pool.QueryRow(ctx, query, args...)
	observe(ctx, span, r.mode, "QueryRow", start, nil)

	return res
}

func observe(
	ctx context.Context,
	span trace.Span,
	mode string,
	op string,
	start time.Time,
	err error,
) {
	elapsed := time.Since(start)

	status := statusSuccess
	if err != nil {
		status = statusError
	}

	attrs := metric.WithAttributes(
		attribute.String("replica", mode),
		attribute.String("operation", op),
		attribute.String("status", status),
	)

	opsLatency.Record(ctx, elapsed.Milliseconds(), attrs)
	opsTotal.Add(ctx, 1, attrs)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		errsTotal.Add(ctx, 1, attrs)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}
