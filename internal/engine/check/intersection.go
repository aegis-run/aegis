package check

import (
	"context"
	"errors"

	"github.com/aegis-run/aegis/pkg/async"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

func (c *checker) evalIntersection(
	ctx context.Context,
	req *Request,
	meta Meta,
	typeDef *schema.Type,
	expr schema.ExprIntersection,
) (_ *Response, err error) {
	ctx, span := telemetry.Start(ctx, "engine.intersection")
	defer telemetry.End(span, &err)

	g := async.NewGroup(ctx, async.WithLimit(c.cfg.ExpressionConcurrencyLimit))
	defer g.Cancel()

	fanoutSize.Record(ctx, int64(len(expr.Terms)))

	checks := async.StreamBatch(g, expr.Terms,
		func(ctx context.Context, t schema.Expr) (*Response, error) {
			return c.evalExpr(ctx, req, meta, typeDef, t)
		},
	)

	var errs error
	for res, err := range checks {
		if async.IsCanceled(err) {
			continue
		}
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if !res.IsAllowed() {
			g.Cancel()
			shortCircuits.Add(ctx, 1)
			return Denied(), nil
		}
	}

	if err := g.Wait(); err != nil && !async.IsCanceled(err) {
		return nil, err
	}

	if errs != nil {
		return nil, errs
	}

	return Allowed(), nil
}
