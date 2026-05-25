package servers

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertPath string `mapstructure:"cert_path"`
	KeyPath  string `mapstructure:"key_path"`
}

func (c *TLSConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.CertPath == "" {
		return errors.New("tls enabled but cert_path is empty")
	}
	if c.KeyPath == "" {
		return errors.New("tls enabled but key_path is empty")
	}
	return nil
}

type GRPCConfig struct {
	Port string    `mapstructure:"port"`
	TLS  TLSConfig `mapstructure:"tls"`
}

func (c *GRPCConfig) Validate() error {
	if c.Port == "" {
		return errors.New("grpc port is empty")
	}
	return c.TLS.Validate()
}

type HTTPConfig struct {
	Enabled        bool      `mapstructure:"enabled"`
	Port           string    `mapstructure:"port"`
	GRPCTargetHost string    `mapstructure:"grpc_target_host"`
	TLS            TLSConfig `mapstructure:"tls"`
}

func (c *HTTPConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Port == "" {
		return errors.New("http port is empty")
	}
	if c.GRPCTargetHost == "" {
		return errors.New("http grpc_target_host is empty")
	}
	return c.TLS.Validate()
}

type PprofConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
}

func (c *PprofConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Port == "" {
		return errors.New("pprof port is empty")
	}
	return nil
}

type Config struct {
	Host         string `mapstructure:"host"`
	NameOverride string `mapstructure:"name_override"`

	GRPC GRPCConfig `mapstructure:"grpc"`
	HTTP HTTPConfig `mapstructure:"http"`
}

func (c *Config) Validate() error {
	var errs []error
	if err := c.GRPC.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("grpc: %w", err))
	}
	if err := c.HTTP.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("http: %w", err))
	}
	return errors.Join(errs...)
}

func DefaultConfig() Config {
	return Config{
		Host:         "",
		NameOverride: "",
		GRPC: GRPCConfig{
			Port: "43615",
			TLS:  TLSConfig{Enabled: false},
		},
		HTTP: HTTPConfig{
			Enabled:        true,
			Port:           "43614",
			GRPCTargetHost: "127.0.0.1",
			TLS:            TLSConfig{Enabled: false},
		},
	}
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("server", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.String("server.host", cfg.Host, "host/interface to bind the server.")
	fs.String("server.name_override", cfg.NameOverride, "")

	fs.String("server.grpc_port", cfg.GRPC.Port, "")
	fs.Bool("server.grpc_tls_enabled", cfg.GRPC.TLS.Enabled, "")
	fs.String("server.grpc_tls_cert_path", cfg.GRPC.TLS.CertPath, "")
	fs.String("server.grpc_tls_key_path", cfg.GRPC.TLS.KeyPath, "")

	fs.Bool("server.http_enabled", cfg.HTTP.Enabled, "")
	fs.String("server.http_port", cfg.HTTP.Port, "")
	fs.String("server.http_grpc_target_host", cfg.HTTP.GRPCTargetHost, "")
	fs.Bool("server.http_tls_enabled", cfg.HTTP.TLS.Enabled, "")
	fs.String("server.http_tls_cert_path", cfg.HTTP.TLS.CertPath, "")
	fs.String("server.http_tls_key_path", cfg.HTTP.TLS.KeyPath, "")

	return fs
}

func DefaultPprofConfig() PprofConfig {
	return PprofConfig{
		Enabled: false,
		Port:    "6060",
	}
}

func PprofFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("profiler", pflag.ExitOnError)

	cfg := DefaultPprofConfig()
	fs.Bool("profiler.enabled", cfg.Enabled, "")
	fs.String("profiler.port", cfg.Port, "")

	return fs
}
