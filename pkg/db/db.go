package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/db/memory"
	"github.com/aegis-run/aegis/pkg/db/migrate"
	"github.com/aegis-run/aegis/pkg/db/postgres"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

func New(ctx context.Context, cfg *core.Config) (core.DB, error) {
	switch cfg.Engine {
	case core.MEMORY:
		return memory.New(ctx)

	case core.POSTGRES:
		db, err := postgres.New(ctx, cfg)
		if err != nil {
			return nil, err
		}

		if cfg.AutoMigrate {
			logger.InfoContext(ctx, "database.postgres.auto_migrate")
			migrate.SetLogger(slog.Default())

			migrateCfg := migrate.Config{
				Engine:          cfg.Engine,
				URI:             cfg.PrimaryURI,
				MigrationsTable: cfg.MigrationsTable,
			}

			if err := migrate.Migrate(ctx, &migrateCfg); err != nil {
				if closeErr := db.Close(); closeErr != nil {
					return nil, fmt.Errorf(
						"auto-migrate failed and db close failed: %w",
						errors.Join(err, closeErr),
					)
				}
				return nil, err
			}
		}
		return db, err

	default:
		return nil, fmt.Errorf("unsupported engine: %s", cfg.Engine)
	}
}
