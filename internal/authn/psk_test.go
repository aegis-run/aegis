package authn

import (
	"strings"
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
)

func TestAuthn_PSK(t *testing.T) {
	t.Run("GenerateKey", func(t *testing.T) {
		key, err := GenerateKey()
		assert.Err(t, err, nil)
		assert.True(t, strings.HasPrefix(string(key), prefix))
	})

	t.Run("ParseKey", func(t *testing.T) {
		validKey, err := GenerateKey()
		assert.Err(t, err, nil)

		tests := []struct {
			name        string
			raw         string
			expectedErr any // Can be nil, error, or string
		}{
			{
				name:        "valid key",
				raw:         string(validKey),
				expectedErr: nil,
			},
			{
				name:        "missing prefix",
				raw:         "some_random_string",
				expectedErr: ErrKeyMissingPrefix,
			},
			{
				name:        "invalid base64",
				raw:         prefix + "invalid_base64_!@#",
				expectedErr: "is not valid base64", // Substring match
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseKey(tt.raw)
				assert.Err(t, err, tt.expectedErr)
			})
		}
	})

	t.Run("String Masking", func(t *testing.T) {
		tests := []struct {
			key      Key
			expected string
		}{
			{key: Key("aegis_1234567890"), expected: "aegis_1234***"},
			{key: Key("aegis_123"), expected: "aegis_***"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.key.String(), tt.expected)
		}
	})
}

func TestAuthn_PSKAuthenticator(t *testing.T) {
	validKey1, _ := GenerateKey()
	validKey2, _ := GenerateKey()

	t.Run("Initialization", func(t *testing.T) {
		_, err := pskAuthenticator([]string{string(validKey1), string(validKey2)})
		assert.Err(t, err, nil)

		_, err = pskAuthenticator([]string{"invalid_key"})
		assert.Err(t, err, ErrKeyMissingPrefix)
	})

	t.Run("Authenticate", func(t *testing.T) {
		auth, err := pskAuthenticator([]string{string(validKey1)})
		assert.Err(t, err, nil)

		tests := []struct {
			name        string
			credential  string
			expectedErr any
		}{
			{
				name:        "valid credential",
				credential:  string(validKey1),
				expectedErr: nil,
			},
			{
				name:        "invalid credential",
				credential:  string(validKey2), // Not loaded in this authenticator
				expectedErr: ErrInvalidCredential,
			},
			{
				name:        "garbage credential",
				credential:  "random_garbage",
				expectedErr: ErrInvalidCredential,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, err := auth.Authenticate(t.Context(), tt.credential)
				assert.Err(t, err, tt.expectedErr)

				if tt.expectedErr == nil {
					assert.True(t, id != nil)
					assert.Equal(t, id.ID, validKey1.String())
				}
			})
		}
	})
}
