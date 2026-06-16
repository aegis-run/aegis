package check

import (
	"context"
	"errors"
	"fmt"

	"github.com/aegis-run/aegis/pkg/async"
	"github.com/aegis-run/aegis/pkg/db"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/tuple"
)

func (c *checker) evalTraversal(
	ctx context.Context,
	req *Request,
	meta Meta,
	expr schema.ExprTraversal,
) (_ *Response, err error) {
	ctx, span := telemetry.Start(ctx, "engine.traversal")
	defer telemetry.End(span, &err)

	var cursor int64
	var errs error

	for {
		batch, nextCursor, err := meta.Querier().Query(ctx,
			tuple.ResourceFilter(req.Resource, expr.Relation),
			db.Pagination{Cursor: cursor, Limit: c.cfg.PageSize},
		)
		if err != nil {
			return nil, fmt.Errorf("query traversal relation %s: %w", expr.Relation, err)
		}

		tuplesFetched.Add(ctx, int64(len(batch)))
		fanoutSize.Record(ctx, int64(len(batch)))

		g := async.NewGroup(ctx, async.WithLimit(c.cfg.TraversalConcurrencyLimit))

		checks := async.StreamBatch(g, batch,
			func(ctx context.Context, t tuple.Tuple) (*Response, error) {
				return c.delegate.Check(ctx,
					NewRequest(t.Subject.Instance, expr.Permission, req.Actor),
					meta.DecrementDepth(),
				)
			},
		)

		for res, err := range checks {
			if async.IsCanceled(err) {
				continue
			}
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			if res.IsAllowed() {
				g.Cancel()
				shortCircuits.Add(ctx, 1)
				return res, nil
			}
		}

		if err := g.Wait(); err != nil && !async.IsCanceled(err) {
			return nil, err
		}

		if nextCursor == 0 || ctx.Err() != nil {
			break
		}
		cursor = nextCursor
	}

	if errs != nil {
		return nil, errs
	}

	return Denied(), nil
}
