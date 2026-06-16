package singleflight

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/aegis-run/aegis/internal/dispatch"
	"github.com/aegis-run/aegis/internal/engine/check"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

func WithSingleflight() dispatch.Option {
	return func(delegate dispatch.Dispatcher) dispatch.Dispatcher {
		return &dispatcher{delegate: delegate}
	}
}

// dispatcher wraps another dispatcher and ensures that identical concurrent
// check requests are only evaluated once, sharing the result with all waiting callers.
type dispatcher struct {
	delegate dispatch.Dispatcher
	sg       singleflight.Group
}

// Check blocks identical concurrent requests and merges their execution paths.
func (d *dispatcher) Check(
	ctx context.Context,
	req *check.Request,
	meta check.Meta,
) (_ *check.Response, err error) {
	ctx, span := telemetry.Start(ctx, "dispatch.singleflight")
	defer telemetry.End(span, &err)

	ch := d.sg.DoChan(req.Key(meta).String(), func() (any, error) {
		return d.delegate.Check(context.WithoutCancel(ctx), req, meta)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Shared {
			dispatch.SingleflightShared.Add(ctx, 1)
		}

		if res.Err != nil {
			return nil, res.Err
		}

		if r, ok := res.Val.(*check.Response); ok {
			return r, nil
		}

		return nil, fmt.Errorf("unexpected response type: %T", res.Val)
	}
}
