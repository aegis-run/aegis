package async

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Group struct {
	ctx    context.Context
	cancel context.CancelFunc
	g      *errgroup.Group
	limit  int
	name   string
	attrs  []Attr
}

type Attr struct {
	Key   string
	Value any
}

type Option func(*options)

type options struct {
	limit int
}

func WithLimit(limit int) Option {
	return func(o *options) {
		o.limit = limit
	}
}

func NewGroup(ctx context.Context, opts ...Option) *Group {
	cfg := options{}
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	if cfg.limit > 0 {
		g.SetLimit(cfg.limit)
	}

	return &Group{
		ctx:    gctx,
		cancel: cancel,
		g:      g,
		limit:  cfg.limit,
	}
}

func (g *Group) Context() context.Context {
	return g.ctx
}

func (g *Group) Go(f func(context.Context) error) {
	g.g.Go(func() error {
		return f(g.ctx)
	})
}

func (g *Group) Cancel() {
	g.cancel()
}

func (g *Group) Wait() error {
	return g.g.Wait()
}

func (g *Group) Limit() int {
	return g.limit
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Attrs() []Attr {
	return g.attrs
}
