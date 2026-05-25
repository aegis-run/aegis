package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/retry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

type database struct {
	ro, rw     *Replica
	maxRetries int
}

func New(ctx context.Context, cfg *db.Config) (*database, error) {
	logger.DebugContext(ctx, "database.postgres.starting")

	write, read, err := initPools(ctx, cfg)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			return
		}

		closePools(write, read)
	}()

	if err = ping(ctx, write, "primary", cfg.MaxRetries); err != nil {
		return nil, err
	}

	if read != write {
		if err = ping(ctx, read, "read_replica", cfg.MaxRetries); err != nil {
			return nil, err
		}
	}

	logger.InfoContext(ctx, "database.postgres.connected")
	return &database{
		rw:         &Replica{pool: write, mode: "primary"},
		ro:         &Replica{pool: read, mode: "readonly"},
		maxRetries: cfg.MaxRetries,
	}, nil
}

var _ DB = (*database)(nil)

func (db *database) RO() *Replica { return db.ro }
func (db *database) RW() *Replica { return db.rw }

func (*database) Engine() string { return "postgres" }

func (db *database) Close() error {
	if db == nil {
		return nil
	}

	closePools(db.rw.pool, db.ro.pool)
	return nil
}

func (db *database) IsReady(ctx context.Context) (bool, error) {
	if db == nil {
		return false, errors.New("database is nil")
	}

	if db.rw == nil || db.rw.pool == nil {
		return false, errors.New("write pool is nil")
	}

	if db.ro == nil || db.ro.pool == nil {
		return false, errors.New("read pool is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.rw.pool.Ping(ctx); err != nil {
		return false, err
	}

	if db.ro.pool != db.rw.pool {
		if err := db.ro.pool.Ping(ctx); err != nil {
			return false, err
		}
	}

	return true, nil
}

func applyConfig(uri string, cfg *db.Config) (*pgxpool.Config, error) {
	poolCfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URI: %w", err)
	}

	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTrimSQLInSpanName())

	poolCfg.MaxConns = cfg.MaxConnections
	poolCfg.MinConns = cfg.MinConnections
	poolCfg.MinIdleConns = cfg.MinIdleConnections

	poolCfg.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	poolCfg.MaxConnLifetime = cfg.MaxConnectionLifetime

	if cfg.MaxConnectionLifetimeJitter > 0 {
		poolCfg.MaxConnLifetimeJitter = cfg.MaxConnectionLifetimeJitter
	} else if cfg.MaxConnectionLifetime > 0 {
		poolCfg.MaxConnLifetimeJitter = time.Duration(0.2 * float64(cfg.MaxConnectionLifetime))
	}

	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	return poolCfg, nil
}

func initPools(ctx context.Context, cfg *db.Config) (write, read *pgxpool.Pool, err error) {
	writeCfg, err := applyConfig(cfg.PrimaryURI, cfg)
	if err != nil {
		return nil, nil, err
	}

	write, err = initPool(ctx, writeCfg)
	if err != nil {
		return nil, nil, err
	}

	if cfg.ReadonlyURI == "" || cfg.ReadonlyURI == cfg.PrimaryURI {
		return write, write, nil
	}

	readCfg, err := applyConfig(cfg.ReadonlyURI, cfg)
	if err != nil {
		return nil, nil, err
	}

	read, err = initPool(ctx, readCfg)
	if err != nil {
		return nil, nil, err
	}

	return write, read, nil
}

func closePools(write, read *pgxpool.Pool) {
	if read != nil && read != write {
		read.Close()
	}

	if write != nil {
		write.Close()
	}
}

func initPool(ctx context.Context, poolCfg *pgxpool.Config) (*pgxpool.Pool, error) {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(initCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

func ping(ctx context.Context, pool *pgxpool.Pool, target string, maxRetries int) error {
	logger.DebugContext(ctx, "database.postgres.ping",
		slog.String("target", target),
	)

	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	policy := retry.New(
		retry.Attempts(maxRetries),
		retry.Backoff(retry.DefaultExpBackoff()),
	)
	err := retry.Do(pingCtx, policy, func(ctx context.Context) error {
		attemptCtx, attemptCancel := context.WithTimeout(ctx, 2*time.Second)
		defer attemptCancel()

		if err := pool.Ping(attemptCtx); err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to verify database connection: %w", err)
	}

	logger.InfoContext(ctx, "database.postgres.connected",
		slog.String("target", target),
	)
	return nil
}
