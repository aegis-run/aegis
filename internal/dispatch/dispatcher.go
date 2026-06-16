package dispatch

import (
	"context"

	"github.com/aegis-run/aegis/internal/engine/check"
)

// Dispatcher defines the interface for evaluating sub-problems during AST traversal.
// It is implemented by various decorators (caching, singleflight, remote routing)
// and ultimately terminates at the local dispatcher which executes the engine logic.
type Dispatcher interface {
	Check(ctx context.Context, req *check.Request, meta check.Meta) (*check.Response, error)
}

type Option func(Dispatcher) Dispatcher

type DispatcherFunc func(
	ctx context.Context,
	req *check.Request,
	meta check.Meta,
) (*check.Response, error)

func With(base Dispatcher, opts ...Option) Dispatcher {
	d := base
	for _, opt := range opts {
		d = opt(d)
	}
	return d
}

func (f DispatcherFunc) Check(
	ctx context.Context,
	req *check.Request,
	meta check.Meta,
) (*check.Response, error) {
	return f(ctx, req, meta)
}
