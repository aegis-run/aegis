package authn

import (
	"context"
	"errors"
)

type Identity struct {
	ID string
}

type Authenticator interface {
	Authenticate(ctx context.Context, credential string) (*Identity, error)
}

func New(_ context.Context, cfg *Config) (Authenticator, error) {
	if !cfg.Enabled {
		return nil, errors.New("authn: disabled")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	switch cfg.Method {
	case MethodPSK:
		return pskAuthenticator(cfg.PSK.Keys)
	case MethodOIDC:
		return nil, errors.New("authn: oidc not supported")
	default:
		return nil, errors.New("authn: unknown method")
	}
}

type contextKey struct{}

func ContextWithIdentity(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func IdentityFromContext(ctx context.Context) *Identity {
	id, ok := ctx.Value(contextKey{}).(*Identity)
	if !ok {
		return nil
	}
	return id
}
