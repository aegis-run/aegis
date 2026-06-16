package check

import (
	"context"
	"fmt"

	"github.com/aegis-run/aegis/pkg/schema"
)

func (c *checker) evalExpr(
	ctx context.Context,
	req *Request,
	meta Meta,
	typeDef *schema.Type,
	expr schema.Expr,
) (*Response, error) {
	switch e := expr.(type) {
	case schema.ExprUnion:
		return c.evalUnion(ctx, req, meta, typeDef, e)

	case schema.ExprIntersection:
		return c.evalIntersection(ctx, req, meta, typeDef, e)

	case schema.ExprDifference:
		return c.evalDifference(ctx, req, meta, typeDef, e)

	case schema.ExprSelfRef:
		return c.evalSelfRef(ctx, req, meta, e)

	case schema.ExprTraversal:
		return c.evalTraversal(ctx, req, meta, e)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}
