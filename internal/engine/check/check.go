package check

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/aegis-run/aegis/internal/engine/loader"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
)

// Checker defines the core engine logic for resolving authorization checks locally.
type Checker interface {
	Check(ctx context.Context, req *Request, meta Meta) (*Response, error)
}

type checker struct {
	schema loader.Schema
	cfg    Config

	delegate Checker
}

func New(schema loader.Schema, delegate Checker, cfg Config) Checker {
	return &checker{
		schema:   schema,
		cfg:      cfg.withDefaults(),
		delegate: delegate,
	}
}

func (c *checker) Check(ctx context.Context, req *Request, meta Meta) (_ *Response, err error) {
	start := time.Now()

	ctx, span := telemetry.Start(ctx, "engine.check", trace.WithAttributes(
		attribute.String("resource.type", req.Resource.Type),
		attribute.String("resource.id", req.Resource.ID),
		attribute.String("permission", req.Permission),
		attribute.String("actor.type", req.Actor.Type),
		attribute.String("actor.id", req.Actor.ID),
	))
	defer telemetry.End(span, &err)

	defer func() {
		duration := time.Since(start).Milliseconds()
		checkDuration.Record(ctx, duration)

		status := "allowed"
		if err != nil {
			status = "error"
		}

		checkRequests.Add(ctx, 1, metric.WithAttributes(attribute.String("status", status)))
	}()

	typeDef, err := c.schema.GetType(ctx, req.Resource.Type, meta.SchemaHash())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTypeNotFound, req.Resource.Type)
	}

	if perm := typeDef.Permission(req.Permission); perm != nil {
		return c.evalExpr(ctx, req, meta, typeDef, perm.Expr)
	}

	if rel := typeDef.Relation(req.Permission); rel != nil {
		return c.evalSelfRef(ctx, req, meta, schema.ExprSelfRef{Relation: req.Permission})
	}

	return nil, fmt.Errorf("%w: %s on type %s", ErrPermissionNotFound,
		req.Permission, req.Resource.Type,
	)
}
