package authn

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

const prefix = "aegis_"
const entropyBytes = 32

var (
	ErrKeyMissingPrefix = errors.New("key must start with " + prefix)
)

type Key string

func GenerateKey() (Key, error) {
	buf := make([]byte, entropyBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return Key(prefix + base64.StdEncoding.EncodeToString(buf)), nil
}

func ParseKey(raw string) (Key, error) {
	if !strings.HasPrefix(raw, prefix) {
		return "", ErrKeyMissingPrefix
	}

	encoded := raw[len(prefix):]
	if _, err := base64.StdEncoding.DecodeString(encoded); err != nil {
		return "", fmt.Errorf("key body is not valid base64: %w", err)
	}

	return Key(raw), nil
}

func (k Key) String() string {
	raw := string(k)
	if len(raw) <= len(prefix)+8 {
		return prefix + "***"
	}
	return raw[:len(prefix)+4] + "***"
}

var ErrInvalidCredential = errors.New("psk: invalid credential")

type PSKAuthenticator struct {
	secret []byte
	hashes map[string]*Identity
}

func (a *PSKAuthenticator) Authenticate(ctx context.Context, credential string) (*Identity, error) {
	hashed := a.hash(credential)
	id, ok := a.hashes[hashed]
	if !ok {
		logger.WarnContext(ctx, "authn.psk.invalid_credential")
		return nil, ErrInvalidCredential
	}

	logger.InfoContext(ctx, "authn.psk.success", "identity", id.ID)
	return id, nil
}

func pskAuthenticator(keys []string) (*PSKAuthenticator, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	a := &PSKAuthenticator{
		secret: secret,
		hashes: make(map[string]*Identity, len(keys)),
	}

	for _, k := range keys {
		key, err := ParseKey(k)
		if err != nil {
			return nil, err
		}

		hashed := a.hash(string(key))
		a.hashes[hashed] = &Identity{
			ID: key.String(),
		}
	}

	return a, nil
}

func (a *PSKAuthenticator) hash(key string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}
