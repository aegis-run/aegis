package datalayer

import (
	"context"

	"github.com/aegis-run/aegis/pkg/schema"
)

type Schema interface {
	Write(ctx context.Context, payload []byte) (schema.Version, error)
	ReadLatest(ctx context.Context) (schema.Version, error)
	ReadByHash(ctx context.Context, hash schema.Hash) (schema.Version, error)
}
