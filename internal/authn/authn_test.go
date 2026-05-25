package authn

import (
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
)

func TestAuthn_New(t *testing.T) {
	validKey, _ := GenerateKey()

	tests := []struct {
		name        string
		cfg         *Config
		expectedErr any  // string, error, or nil
		expectNil   bool // whether we expect the returned Authenticator to be nil
	}{
		{
			name: "disabled returns error",
			cfg: &Config{
				Enabled: false,
			},
			expectedErr: "authn: disabled",
			expectNil:   true,
		},
		{
			name: "unknown method returns error",
			cfg: &Config{
				Enabled: true,
				Method:  "magic_method",
			},
			expectedErr: "unsupported authn method: magic_method",
			expectNil:   true,
		},
		{
			name: "oidc method fails with invalid config (no issuer)",
			cfg: &Config{
				Enabled: true,
				Method:  MethodOIDC,
			},
			expectedErr: "OIDC issuer is required",
			expectNil:   true,
		},
		{
			name: "psk method fails with invalid keys",
			cfg: &Config{
				Enabled: true,
				Method:  MethodPSK,
				PSK: PSKConfig{
					Keys: []string{"invalid_key_without_prefix"},
				},
			},
			expectedErr: ErrKeyMissingPrefix,
			expectNil:   true,
		},
		{
			name: "psk method succeeds",
			cfg: &Config{
				Enabled: true,
				Method:  MethodPSK,
				PSK: PSKConfig{
					Keys: []string{string(validKey)},
				},
			},
			expectedErr: nil,
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator, err := New(t.Context(), tt.cfg)
			assert.Err(t, err, tt.expectedErr)

			if tt.expectNil {
				assert.True(t, err != nil)
			} else {
				assert.True(t, authenticator != nil)
			}
		})
	}
}

func TestAuthn_Context(t *testing.T) {
	ctx := t.Context()

	id := IdentityFromContext(ctx)
	assert.True(t, id == nil)

	expectedID := &Identity{ID: "test_user"}
	ctx = ContextWithIdentity(ctx, expectedID)

	id = IdentityFromContext(ctx)
	assert.True(t, id != nil)
	assert.Equal(t, id.ID, expectedID.ID)
}
