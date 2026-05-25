package migrate

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aegis-run/aegis/pkg/db/core"
)

type Config struct {
	Engine          core.Engine `mapstructure:"engine"`
	URI             string      `mapstructure:"uri"`
	MigrationsTable string      `mapstructure:"migrations_table"`
}

func DefaultConfig() Config {
	defaults := core.DefaultConfig()

	return Config{
		Engine:          defaults.Engine,
		URI:             defaults.PrimaryURI,
		MigrationsTable: defaults.MigrationsTable,
	}
}

func (c Config) Validate() error {
	var errs []error

	switch c.Engine {
	case core.MEMORY, core.POSTGRES:
		// valid
	case "":
		errs = append(errs, errors.New("database engine is required"))
	default:
		errs = append(errs, fmt.Errorf("unsupported database engine: %s", c.Engine))
	}

	if c.Engine == core.POSTGRES && c.URI == "" {
		errs = append(errs, errors.New("database URI is required for postgres engine"))
	}

	if c.MigrationsTable == "" {
		errs = append(errs, errors.New("migrations table is required"))
	}

	return errors.Join(errs...)
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("database", pflag.ExitOnError)

	defaults := DefaultConfig()
	fs.String("db.engine", string(defaults.Engine), "database engine (memory or postgres)")
	fs.String("db.uri", defaults.URI, "primary database connection URI")
	fs.String("db.migrations_table", defaults.MigrationsTable, "migrations table name")

	return fs
}

func Load(v *viper.Viper) (Config, error) {
	cfg := DefaultConfig()

	raw := struct {
		Database Config `mapstructure:"db"`
	}{
		Database: cfg,
	}

	if err := v.Unmarshal(&raw); err != nil {
		return cfg, err
	}

	cfg = raw.Database

	return cfg, cfg.Validate()
}
