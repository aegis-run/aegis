package ristretto

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"

	"github.com/aegis-run/aegis/pkg/cache/core"
)

type Ristretto[V any] struct {
	c *ristretto.Cache[string, V]
}

func New[V any](cfg *core.Config) (*Ristretto[V], error) {
	c, err := ristretto.NewCache(&ristretto.Config[string, V]{
		NumCounters: cfg.Memory.NumCounters,
		MaxCost:     cfg.Memory.MaxCost,
		BufferItems: cfg.Memory.BufferItems,
	})
	if err != nil {
		return nil, err
	}
	return &Ristretto[V]{c: c}, nil
}

func (r *Ristretto[V]) Get(key string) (V, bool) {
	return r.c.Get(key)
}

func (r *Ristretto[V]) Set(key string, value V, cost int64, ttl time.Duration) bool {
	if ttl > 0 {
		return r.c.SetWithTTL(key, value, cost, ttl)
	}

	return r.c.Set(key, value, cost)
}

func (r *Ristretto[V]) Wait() {
	r.c.Wait()
}

func (r *Ristretto[V]) Close() {
	r.c.Close()
}
