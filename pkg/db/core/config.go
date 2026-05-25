package core

import (
	"errors"
	"time"

	"github.com/spf13/pflag"
)

type Engine string

const (
	MEMORY   Engine = "memory"
	POSTGRES Engine = "postgres"
)

type Config struct {
	Engine                      Engine        `mapstructure:"engine"`
	AutoMigrate                 bool          `mapstructure:"auto_migrate"`
	MigrationsTable             string        `mapstructure:"migrations_table"`
	PrimaryURI                  string        `mapstructure:"uri"`
	ReadonlyURI                 string        `mapstructure:"readonly_uri"`
	MaxConnections              int32         `mapstructure:"max_connections"`
	MinConnections              int32         `mapstructure:"min_connections"`
	MinIdleConnections          int32         `mapstructure:"min_idle_connections"`
	MaxConnectionLifetime       time.Duration `mapstructure:"max_connection_lifetime"`
	MaxConnectionIdleTime       time.Duration `mapstructure:"max_connection_idle_time"`
	HealthCheckPeriod           time.Duration `mapstructure:"health_check_period"`
	MaxConnectionLifetimeJitter time.Duration `mapstructure:"max_connection_lifetime_jitter"`
	ConnectTimeout              time.Duration `mapstructure:"connect_timeout"`
	MaxRetries                  int           `mapstructure:"max_retries"`
}

func DefaultConfig() Config {
	return Config{
		Engine:                      MEMORY,
		MigrationsTable:             "migrations",
		MaxConnections:              20,
		MinConnections:              1,
		MinIdleConnections:          1,
		MaxConnectionLifetime:       300 * time.Second,
		MaxConnectionIdleTime:       60 * time.Second,
		MaxConnectionLifetimeJitter: 1 * time.Minute,
		HealthCheckPeriod:           30 * time.Second,
		ConnectTimeout:              5 * time.Second,
		MaxRetries:                  3,
	}
}

//nolint:cyclop
func (c Config) Validate() error {
	if c.Engine != MEMORY && c.PrimaryURI == "" {
		return errors.New("primary database URI is required")
	}

	if c.MaxConnections <= 0 {
		return errors.New("max_connections must be greater than 0")
	}

	if c.MinConnections < 0 {
		return errors.New("min_connections cannot be negative")
	}

	if c.MinIdleConnections < 0 {
		return errors.New("min_idle_connections cannot be negative")
	}

	if c.MinConnections > c.MaxConnections {
		return errors.New("min_connections cannot exceed max_connections")
	}

	if c.MinIdleConnections > c.MaxConnections {
		return errors.New("min_idle_connections cannot exceed max_connections")
	}

	if c.MinIdleConnections > c.MinConnections {
		return errors.New("min_idle_connections cannot exceed min_connections")
	}

	if c.MaxConnectionLifetime < 0 {
		return errors.New("max_connection_lifetime cannot be negative")
	}

	if c.MaxConnectionIdleTime < 0 {
		return errors.New("max_connection_idle_time cannot be negative")
	}

	if c.MaxConnectionLifetimeJitter < 0 {
		return errors.New("max_connection_lifetime_jitter cannot be negative")
	}

	if c.HealthCheckPeriod < 0 {
		return errors.New("health_check_period cannot be negative")
	}

	if c.ConnectTimeout < 0 {
		return errors.New("connect_timeout cannot be negative")
	}

	if c.MaxRetries < 0 {
		return errors.New("max_retries cannot be negative")
	}

	return nil
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("datastore", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.String("datastore.engine", string(cfg.Engine), "Database engine (memory or postgres)")
	fs.Bool("datastore.auto_migrate", cfg.AutoMigrate, "Enable auto-migration")
	fs.String("datastore.migrations_table", cfg.MigrationsTable, "Migrations table name")
	fs.String("datastore.uri", cfg.PrimaryURI, "Primary database URI")
	fs.String("datastore.readonly_uri", cfg.ReadonlyURI, "Read-only database URI")
	fs.Int32("datastore.max_connections", cfg.MaxConnections, "Max database connections")
	fs.Int32("datastore.min_connections", cfg.MinConnections, "Min database connections")
	fs.Int32("datastore.min_idle_connections", cfg.MinIdleConnections, "Min idle database connections")
	fs.Duration("datastore.max_connection_lifetime", cfg.MaxConnectionLifetime, "Max database connection lifetime")
	fs.Duration("datastore.max_connection_idle_time", cfg.MaxConnectionIdleTime, "Max database connection idle time")
	fs.Duration("datastore.health_check_period", cfg.HealthCheckPeriod, "Database health check period")
	fs.Duration("datastore.max_connection_lifetime_jitter", cfg.MaxConnectionLifetimeJitter, "Max database connection lifetime jitter")
	fs.Duration("datastore.connect_timeout", cfg.ConnectTimeout, "Database connection timeout")
	fs.Int("datastore.max_retries", cfg.MaxRetries, "Max database operation retries")

	return fs
}
