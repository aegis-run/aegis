package authn

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

type Method string

const (
	MethodPSK  Method = "psk"
	MethodOIDC Method = "oidc"
)

type Config struct {
	Enabled bool       `mapstructure:"enabled"`
	Method  Method     `mapstructure:"method"`
	PSK     PSKConfig  `mapstructure:"psk"`
	OIDC    OIDCConfig `mapstructure:"oidc"`
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	switch c.Method {
	case MethodPSK:
		return c.PSK.Validate()
	case MethodOIDC:
		return c.OIDC.Validate()
	case "":
		return errors.New("authn method is required when enabled")
	default:
		return fmt.Errorf("unsupported authn method: %s", c.Method)
	}
}

type PSKConfig struct {
	Keys []string `mapstructure:"keys"`
}

func (p *PSKConfig) Validate() error {
	if len(p.Keys) == 0 {
		return errors.New("at least one PSK key must be provided")
	}

	return nil
}

type OIDCConfig struct {
	Issuer              string        `mapstructure:"issuer"`
	Audience            string        `mapstructure:"audience"`
	IdentityClaim       string        `mapstructure:"identity"`
	JWKSRefreshInterval time.Duration `mapstructure:"jwks_refresh_interval"`
}

func (o *OIDCConfig) Validate() error {
	if o.Issuer == "" {
		return errors.New("OIDC issuer is required")
	}
	if o.Audience == "" {
		return errors.New("OIDC Audience is required")
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		Enabled: false,
		PSK:     PSKConfig{},
		OIDC: OIDCConfig{
			IdentityClaim:       "sub",
			JWKSRefreshInterval: time.Hour,
		},
	}
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("authn", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.Bool("authn.enabled", cfg.Enabled, "Enable authentication")
	fs.String("authn.method", string(cfg.Method), "Authentication method (psk or oidc)")

	fs.StringSlice("authn.psk_keys", cfg.PSK.Keys, "List of PSK keys")

	fs.String("authn.oidc_issuer", cfg.OIDC.Issuer, "OIDC issuer URL")
	fs.String("authn.oidc_audience", cfg.OIDC.Audience, "OIDC audience")
	fs.String("authn.oidc_identity_claim", cfg.OIDC.IdentityClaim, "OIDC identity claim")
	fs.Duration("authn.oidc_jwks_refresh_interval",
		cfg.OIDC.JWKSRefreshInterval,
		"OIDC JWKS refresh interval",
	)

	return fs
}
