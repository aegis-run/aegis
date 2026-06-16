package migrate

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	db "github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/db/postgres"
	"github.com/aegis-run/aegis/pkg/db/postgres/migrations"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

var gooseLogger *slog.Logger

// SetLogger sets the logger used by migrations. If nil, slog.Default() is used.
func SetLogger(l *slog.Logger) {
	gooseLogger = l
}

func Migrate(ctx context.Context, cfg *db.Config) error {
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeDB(db)

	p, err := newProvider(db, cfg)
	if err != nil {
		return err
	}

	_, err = p.Up(ctx)
	return err
}

func Reset(ctx context.Context, cfg *db.Config) error {
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeDB(db)

	p, err := newProvider(db, cfg)
	if err != nil {
		return err
	}

	_, err = p.DownTo(ctx, 0)
	return err
}

func Status(ctx context.Context, cfg *db.Config) error {
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeDB(db)

	p, err := newProvider(db, cfg)
	if err != nil {
		return err
	}

	_, err = p.Status(ctx)
	return err
}

func newProvider(db postgres.DB, cfg *db.Config) (*goose.Provider, error) {
	pool := stdlib.OpenDBFromPool(db.RW().Pool())

	l := gooseLogger
	if l == nil {
		l = slog.Default()
	}

	return goose.NewProvider(
		goose.DialectPostgres,
		pool,
		migrations.Migrations,
		goose.WithTableName(cfg.MigrationsTable),
		goose.WithSlog(l),
		goose.WithVerbose(true),
	)
}

func closeDB(db db.DB) {
	if err := db.Close(); err != nil {
		logger.Error("database.postgres.close_failed",
			"error", err)
	}
}
