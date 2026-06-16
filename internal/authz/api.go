package authz

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aegis-run/aegis/internal/datalayer"
	"github.com/aegis-run/aegis/internal/engine"
	"github.com/aegis-run/aegis/internal/engine/check"
	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/telemetry"
	authzv1 "github.com/aegis-run/aegis/proto/aegis/authz/v1"
)

type API struct {
	authzv1.UnimplementedAuthzServer

	engine *engine.Engine

	schema      datalayer.Schema
	consistency datalayer.Consistency
}

func NewAPI(engine *engine.Engine, dl *datalayer.DataLayer) *API {
	return &API{engine: engine, schema: dl.Schema, consistency: dl.Consistency}
}

func (api *API) Authorize(
	ctx context.Context,
	req *authzv1.AuthorizeRequest,
) (res *authzv1.AuthorizeResponse, err error) {
	ctx, span := telemetry.Start(ctx, "authz.api.authorize")
	defer telemetry.End(span, &err)

	requirement := consistency.DecodeRequirement(req.GetConsistency())
	token, err := api.consistency.Resolve(ctx, requirement)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid consistency requirement: %v", err)
	}

	querier, cleanup, err := api.consistency.Querier(ctx, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	defer cleanup()

	sch, err := api.schema.ReadLatest(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read schema: %v", err)
	}

	resp, err := api.engine.Check(ctx,
		check.DecodeRequest(req),
		check.NewMeta(api.engine.Config().MaxDepth, sch.Hash, token, querier),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "evaluation failed: %v", err)
	}

	return resp.Encode(consistency.Encode(token)), nil
}
