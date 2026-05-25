package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aegis-run/aegis/pkg/db/core"
	pgMigrate "github.com/aegis-run/aegis/pkg/db/postgres/migrate"
)

func SetLogger(l *slog.Logger) {
	pgMigrate.SetLogger(l)
}

func Migrate(ctx context.Context, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	coreCfg := toCoreConfig(cfg)

	switch cfg.Engine {
	case core.MEMORY:
		return nil

	case core.POSTGRES:
		return pgMigrate.Migrate(ctx, &coreCfg)

	default:
		return fmt.Errorf("unsupported engine: %s", cfg.Engine)
	}
}

func Reset(ctx context.Context, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	coreCfg := toCoreConfig(cfg)

	switch cfg.Engine {
	case core.MEMORY:
		return nil

	case core.POSTGRES:
		return pgMigrate.Reset(ctx, &coreCfg)

	default:
		return fmt.Errorf("unsupported engine: %s", cfg.Engine)
	}
}

func Status(ctx context.Context, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	coreCfg := toCoreConfig(cfg)

	switch cfg.Engine {
	case core.MEMORY:
		return nil

	case core.POSTGRES:
		return pgMigrate.Status(ctx, &coreCfg)

	default:
		return fmt.Errorf("unsupported engine: %s", cfg.Engine)
	}
}

func toCoreConfig(cfg *Config) core.Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Engine = cfg.Engine
	coreCfg.PrimaryURI = cfg.URI
	coreCfg.MigrationsTable = cfg.MigrationsTable

	return coreCfg
}
