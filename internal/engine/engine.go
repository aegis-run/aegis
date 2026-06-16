package engine

import (
	"context"

	"github.com/aegis-run/aegis/internal/dispatch"
	"github.com/aegis-run/aegis/internal/dispatch/local"
	"github.com/aegis-run/aegis/internal/engine/check"
	"github.com/aegis-run/aegis/internal/engine/loader"
)

type Engine struct {
	cfg        *Config
	dispatcher dispatch.Dispatcher
}

func New(
	schema loader.Schema,
	cfg *Config,
	opts ...dispatch.Option,
) *Engine {
	var dispatcher dispatch.Dispatcher

	thunk := dispatch.DispatcherFunc(
		func(ctx context.Context, req *check.Request, meta check.Meta) (*check.Response, error) {
			return dispatcher.Check(ctx, req, meta)
		})

	chk := check.New(schema, thunk, check.Config{
		ExpressionConcurrencyLimit: cfg.ExpressionConcurrencyLimit,
		TraversalConcurrencyLimit:  cfg.TraversalConcurrencyLimit,
		PageSize:                   int32(cfg.PageSize),
	})
	dispatcher = dispatch.With(local.NewDispatcher(chk), opts...)

	return &Engine{
		cfg:        cfg,
		dispatcher: dispatcher,
	}
}

func (e *Engine) Config() *Config {
	return e.cfg
}

func (e *Engine) Check(
	ctx context.Context,
	req *check.Request,
	meta check.Meta,
) (*check.Response, error) {
	return e.dispatcher.Check(ctx, req, meta)
}
