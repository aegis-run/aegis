package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/aegis-run/aegis/pkg/consistency"
	core "github.com/aegis-run/aegis/pkg/db"
	db "github.com/aegis-run/aegis/pkg/db/postgres"
)

// Consistency provides database consistency token resolution for Postgres.
type Consistency struct {
	db db.DB

	quantizedToken atomic.Value
	cancel         context.CancelFunc
}

// NewConsistency creates a new Consistency instance with the given database connection
// and starts a background goroutine to update the quantized token every window duration.
func NewConsistency(ctx context.Context, database db.DB, window time.Duration) *Consistency {
	ctx, cancel := context.WithCancel(ctx)
	c := &Consistency{
		db:     database,
		cancel: cancel,
	}

	// Fetch initial token synchronously so StrategyMinimizeLatency is immediately ready.
	initialToken, err := c.fetchCurrentXactID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch initial quantized token", "err", err)
		// Fallback to a zero-token if we somehow fail on startup
		initialToken = db.ConsistencyToken{}
	}
	c.quantizedToken.Store(initialToken)

	go c.runQuantizer(ctx, window)

	return c
}

// Close stops the background quantization goroutine.
func (c *Consistency) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Consistency) runQuantizer(ctx context.Context, window time.Duration) {
	ticker := time.NewTicker(window)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			token, err := c.fetchCurrentXactID(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "failed to refresh quantized token", "err", err)
				continue
			}
			c.quantizedToken.Store(token)
		}
	}
}

func (c *Consistency) fetchCurrentXactID(ctx context.Context) (consistency.Token, error) {
	return db.TxWithResult(ctx, c.db.RW(), func(tx db.DBTX) (consistency.Token, error) {
		xid, err := db.Query.GetCurrentXactID(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current transaction ID: %w", err)
		}
		return db.ConsistencyToken{Value: xid}, nil
	})
}

// Resolve maps a database-agnostic consistency requirement to a database token.
func (c *Consistency) Resolve(
	ctx context.Context,
	req consistency.Requirement,
) (consistency.Token, error) {
	switch req.Strategy {
	case consistency.StrategyFresherThan:
		if req.Token == "" {
			return nil, fmt.Errorf("empty consistency token")
		}
		val, err := strconv.ParseUint(req.Token, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid postgres token format: %w", err)
		}
		return db.ConsistencyToken{
			Value: pgtype.Uint64{Uint64: val, Valid: true},
		}, nil

	case consistency.StrategyFullyConsistent:
		return c.fetchCurrentXactID(ctx)

	case consistency.StrategyMinimizeLatency:
		val := c.quantizedToken.Load()
		if val == nil {
			return nil, fmt.Errorf("quantized token is not initialized")
		}

		return val.(consistency.Token), nil
	default:
		//nolint:nilnil
		return nil, nil
	}
}

func (c *Consistency) Querier(
	ctx context.Context,
	token consistency.Token,
) (core.Querier, func(), error) {
	target := c.db.RO()

	if token != nil && c.db.RO().Pool() != c.db.RW().Pool() {
		pgToken, err := db.FromToken(token)
		if err != nil {
			return nil, nil, err
		}

		visible, err := c.isVisibleOnReplica(ctx, pgToken)
		if err != nil {
			return nil, nil, err
		}

		if !visible {
			target = c.db.RW()
		}
	}

	return &queryRunner{tx: target}, func() {}, nil
}

func (c *Consistency) isVisibleOnReplica(
	ctx context.Context,
	token db.ConsistencyToken,
) (bool, error) {
	return db.TxWithResult(ctx, c.db.RO(), func(tx db.DBTX) (bool, error) {
		visible, err := db.Query.IsXactVisible(ctx, tx, token.Value)
		if err != nil {
			return false, fmt.Errorf("failed to check transaction visibility: %w", err)
		}
		return visible, nil
	})
}
