package caching

import (
	"context"
	"time"
	"unsafe"

	"github.com/aegis-run/aegis/internal/dispatch"
	"github.com/aegis-run/aegis/internal/engine/check"
	"github.com/aegis-run/aegis/pkg/cache"
	cacheCore "github.com/aegis-run/aegis/pkg/cache/core"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

func NewCache(cfg *cacheCore.Config) (cacheCore.Cache[*check.Response], error) {
	return cache.New[*check.Response](cfg)
}

func WithCaching(c cacheCore.Cache[*check.Response], ttl time.Duration) dispatch.Option {
	return func(delegate dispatch.Dispatcher) dispatch.Dispatcher {
		return &dispatcher{delegate: delegate, cache: c, ttl: ttl}
	}
}

// dispatcher wraps another dispatcher and caches successful check responses.
type dispatcher struct {
	delegate dispatch.Dispatcher
	cache    cacheCore.Cache[*check.Response]
	ttl      time.Duration
}

// Check intercepts the request, checks the cache using a composite key,
// and delegates downward on a cache miss.
func (d *dispatcher) Check(
	ctx context.Context,
	req *check.Request,
	meta check.Meta,
) (_ *check.Response, err error) {
	ctx, span := telemetry.Start(ctx, "dispatch.caching")
	defer telemetry.End(span, &err)

	key := req.Key(meta)

	// 1. Try Cache
	if resp, ok := d.cache.Get(key.String()); ok {
		dispatch.CacheHits.Add(ctx, 1)
		return resp, nil
	}
	
	dispatch.CacheMisses.Add(ctx, 1)

	// 2. Cache Miss: Delegate down the stack (e.g. to singleflight)
	resp, err := d.delegate.Check(ctx, req, meta)
	if err != nil {
		return nil, err // Do not cache errors
	}

	// 3. Cache the result for future identical requests in this tree
	d.cache.Set(key.String(), resp, int64(unsafe.Sizeof(*resp)), d.ttl)
	return resp, nil
}
