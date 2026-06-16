package check

import (
	"context"

	"github.com/aegis-run/aegis/pkg/async"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

func (c *checker) evalDifference(
	ctx context.Context,
	req *Request,
	meta Meta,
	typeDef *schema.Type,
	expr schema.ExprDifference,
) (_ *Response, err error) {
	ctx, span := telemetry.Start(ctx, "engine.difference")
	defer telemetry.End(span, &err)

	g := async.NewGroup(ctx, async.WithLimit(2))
	defer func() {
		g.Cancel()
		_ = g.Wait()
	}()

	lhs := async.Go(g, func(ctx context.Context) (*Response, error) {
		return c.evalExpr(ctx, req, meta, typeDef, expr.LHS)
	})
	rhs := async.Go(g, func(ctx context.Context) (*Response, error) {
		return c.evalExpr(ctx, req, meta, typeDef, expr.RHS)
	})

	lhsRes, err := lhs.Wait()
	if err != nil {
		return nil, err
	}

	if !lhsRes.IsAllowed() {
		shortCircuits.Add(ctx, 1)
		return Denied(), nil
	}

	rhsRes, err := rhs.Wait()
	if err != nil {
		return nil, err
	}

	if rhsRes.IsAllowed() {
		return Denied(), nil
	}
	return Allowed(), nil
}
