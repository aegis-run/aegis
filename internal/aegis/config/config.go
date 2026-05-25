package config

import (
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aegis-run/aegis/internal/authn"
	"github.com/aegis-run/aegis/internal/servers"
	db "github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

type Config struct {
	Server    servers.Config      `mapstructure:"server"`
	Profiler  servers.PprofConfig `mapstructure:"profiler"`
	Datastore db.Config           `mapstructure:"datastore"`
	Tracing   telemetry.Tracing   `mapstructure:"tracing"`
	Metrics   telemetry.Metrics   `mapstructure:"metrics"`
	Logs      telemetry.Logs      `mapstructure:"logs"`
	Authn     authn.Config        `mapstructure:"authn"`
}

// Default returns the default configuration for the Aegis service.
func Default() Config {
	return Config{
		Server:    servers.DefaultConfig(),
		Profiler:  servers.DefaultPprofConfig(),
		Datastore: db.DefaultConfig(),
		Tracing:   telemetry.DefaultTracingConfig(),
		Metrics:   telemetry.DefaultMetricsConfig(),
		Logs:      telemetry.DefaultLogsConfig(),
		Authn:     authn.DefaultConfig(),
	}
}

// Flags returns the flag set for the Aegis service.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ExitOnError)

	f.AddFlagSet(servers.Flags())
	f.AddFlagSet(servers.PprofFlags())
	f.AddFlagSet(db.Flags())
	f.AddFlagSet(telemetry.TracingFlags())
	f.AddFlagSet(telemetry.MetricsFlags())
	f.AddFlagSet(telemetry.LogsFlags())
	f.AddFlagSet(authn.Flags())

	return f
}

func Load(v *viper.Viper) (Config, error) {
	cfg := Default()

	err := v.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	))
	if err != nil {
		return cfg, err
	}

	return cfg, cfg.Validate()
}

func (cfg *Config) Validate() error {
	var errs []error

	if err := cfg.Server.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("server: %w", err))
	}

	if err := cfg.Tracing.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("tracing: %w", err))
	}

	if err := cfg.Metrics.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("metrics: %w", err))
	}

	if err := cfg.Logs.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("logs: %w", err))
	}

	if err := cfg.Authn.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("authn: %w", err))
	}

	return errors.Join(errs...)
}
