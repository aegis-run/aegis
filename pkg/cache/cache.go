package cache

import (
	"fmt"

	"github.com/aegis-run/aegis/pkg/cache/core"
	"github.com/aegis-run/aegis/pkg/cache/ristretto"
)

func New[V any](cfg *core.Config) (core.Cache[V], error) {
	switch cfg.Engine {
	case core.MEMORY:
		return ristretto.New[V](cfg)

	default:
		return nil, fmt.Errorf("unsupported engine: %s", cfg.Engine)
	}
}
