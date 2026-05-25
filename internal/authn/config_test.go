package authn

import (
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
)

func TestAuthn_Config(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		tests := []struct {
			name        string
			cfg         Config
			expectedErr any
		}{
			{
				name: "disabled config",
				cfg: Config{
					Enabled: false,
					Method:  "",
				},
				expectedErr: nil,
			},
			{
				name: "enabled missing method",
				cfg: Config{
					Enabled: true,
					Method:  "",
				},
				expectedErr: "authn method is required",
			},
			{
				name: "enabled unknown method",
				cfg: Config{
					Enabled: true,
					Method:  "magic",
				},
				expectedErr: "unsupported authn method: magic",
			},
			{
				name: "valid psk config",
				cfg: Config{
					Enabled: true,
					Method:  MethodPSK,
					PSK:     PSKConfig{Keys: []string{"aegis_123"}},
				},
				expectedErr: nil,
			},
			{
				name: "invalid psk config (no keys)",
				cfg: Config{
					Enabled: true,
					Method:  MethodPSK,
					PSK:     PSKConfig{Keys: []string{}},
				},
				expectedErr: "at least one PSK key must be provided",
			},
			{
				name: "valid oidc config",
				cfg: Config{
					Enabled: true,
					Method:  MethodOIDC,
					OIDC:    OIDCConfig{Issuer: "https://auth.example.com", Audience: "test_audience"},
				},
				expectedErr: nil,
			},
			{
				name: "invalid oidc config (no issuer)",
				cfg: Config{
					Enabled: true,
					Method:  MethodOIDC,
					OIDC:    OIDCConfig{Issuer: ""},
				},
				expectedErr: "OIDC issuer is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.cfg.Validate()
				assert.Err(t, err, tt.expectedErr)
			})
		}
	})
}
