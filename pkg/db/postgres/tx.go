package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

// ErrTxBegin represents a failure to start a database transaction.
type ErrTxBegin struct {
	Err error
}

func (e *ErrTxBegin) Error() string {
	return fmt.Sprintf("failed to begin transaction: %v", e.Err)
}

func (e *ErrTxBegin) Unwrap() error {
	return e.Err
}

// ErrTxCommit represents a failure to commit a database transaction.
type ErrTxCommit struct {
	Err error
}

func (e *ErrTxCommit) Error() string {
	return fmt.Sprintf("failed to commit transaction: %v", e.Err)
}

func (e *ErrTxCommit) Unwrap() error {
	return e.Err
}

// ErrTxRollback represents a failure to rollback a database transaction.
type ErrTxRollback struct {
	Err error
}

func (e *ErrTxRollback) Error() string {
	return fmt.Sprintf("failed to rollback transaction: %v", e.Err)
}

func (e *ErrTxRollback) Unwrap() error {
	return e.Err
}

// TracedTx wraps a standard pgx.Tx to provide transparent telemetry (spans and metrics)
// for all query operations executed within a database transaction.
type TracedTx struct {
	tx   pgx.Tx
	mode string
}

// Ensure TracedTx implements DBTX.
var _ DBTX = (*TracedTx)(nil)

// Exec executes a SQL command within the transaction with tracing and metrics.
func (t *TracedTx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	ctx, span := telemetry.Start(ctx, "database.postgres.tx.exec", trace.WithAttributes(
		attribute.String("mode", t.mode),
		attribute.String("query", query),
	))
	defer span.End()

	start := time.Now()
	res, err := t.tx.Exec(ctx, query, args...)
	observe(ctx, span, t.mode, "Exec", start, err)

	return res, err
}

// Query executes a query that returns rows within the transaction with tracing and metrics.
func (t *TracedTx) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	ctx, span := telemetry.Start(ctx, "database.postgres.tx.query", trace.WithAttributes(
		attribute.String("mode", t.mode),
		attribute.String("query", query),
	))
	defer span.End()

	start := time.Now()
	res, err := t.tx.Query(ctx, query, args...)
	observe(ctx, span, t.mode, "Query", start, err)

	return res, err
}

// QueryRow executes a query that returns at most one row within the transaction with
// tracing and metrics.
func (t *TracedTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	ctx, span := telemetry.Start(ctx, "database.postgres.tx.query_row", trace.WithAttributes(
		attribute.String("mode", t.mode),
		attribute.String("query", query),
	))
	defer span.End()

	start := time.Now()
	res := t.tx.QueryRow(ctx, query, args...)
	observe(ctx, span, t.mode, "QueryRow", start, nil)

	return res
}

// Commit commits the transaction and traces the operation.
func (t *TracedTx) Commit(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "database.postgres.tx.commit", trace.WithAttributes(
		attribute.String("mode", t.mode),
	))
	defer span.End()

	start := time.Now()
	err := t.tx.Commit(ctx)
	observe(ctx, span, t.mode, "Commit", start, err)

	if err != nil {
		return &ErrTxCommit{Err: err}
	}
	return nil
}

// Rollback rolls back the transaction and traces the operation.
func (t *TracedTx) Rollback(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "database.postgres.tx.rollback", trace.WithAttributes(
		attribute.String("mode", t.mode),
	))
	defer span.End()

	start := time.Now()
	err := t.tx.Rollback(ctx)
	observe(ctx, span, t.mode, "Rollback", start, err)

	if err != nil {
		return &ErrTxRollback{Err: err}
	}
	return nil
}

type TxFunc func(DBTX) error
type TxWithResultFunc[T any] func(DBTX) (T, error)

// Tx executes fn within a database transaction on the provided Replica.
// It automatically manages transaction begin, commit, rollback, and proper resource cleanup.
// If fn returns an error or panics, the transaction is automatically rolled back.
func Tx(ctx context.Context, r *Replica, fn TxFunc) error {
	_, err := TxWithResult(ctx, r, func(tx DBTX) (any, error) {
		return nil, fn(tx)
	})
	return err
}

// TxWithResult executes fn within a database transaction on the provided Replica,
// returning a result. It automatically manages transaction begin, commit, rollback, and
// proper resource cleanup. If fn returns an error or panics, the transaction is
// automatically rolled back.
//
//nolint:cyclop
func TxWithResult[T any](ctx context.Context, r *Replica, fn TxWithResultFunc[T]) (t T, err error) {
	if r == nil || r.pool == nil {
		return t, &ErrTxBegin{Err: errors.New("database pool is nil")}
	}

	ctx, span := telemetry.Start(ctx, "database.postgres.tx_runner", trace.WithAttributes(
		attribute.String("mode", r.mode),
	))
	defer span.End()

	rawTx, err := r.pool.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return t, &ErrTxBegin{Err: err}
	}

	tx := &TracedTx{
		tx:   rawTx,
		mode: r.mode,
	}

	defer func() {
		if p := recover(); p != nil {
			rollbackCtx, rollbackCancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
			rollbackErr := tx.Rollback(rollbackCtx)
			rollbackCancel()

			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				if ctx.Err() != nil {
					logger.DebugContext(ctx, "database.postgres.tx.panic_rollback_failed_after_cancel",
						"error", rollbackErr,
					)
				} else {
					logger.ErrorContext(ctx, "database.postgres.tx.panic_rollback_failed",
						"error", rollbackErr,
					)
				}
			}
			panic(p) // re-panic after rolling back
		}
	}()

	t, err = fn(tx)
	if err != nil {
		rollbackCtx, rollbackCancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		rollbackErr := tx.Rollback(rollbackCtx)
		rollbackCancel()

		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			if ctx.Err() != nil {
				logger.DebugContext(ctx, "database.postgres.tx.rollback_failed_after_cancel",
					"error", rollbackErr,
				)
			} else {
				logger.ErrorContext(ctx, "database.postgres.tx.rollback_failed",
					"error", rollbackErr,
				)
			}
		}

		return t, err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return t, commitErr
	}

	return t, nil
}

// WrapTxForTest wraps a standard pgx.Tx for testing purposes.
func WrapTxForTest(tx pgx.Tx, mode string) *TracedTx {
	return &TracedTx{
		tx:   tx,
		mode: mode,
	}
}
