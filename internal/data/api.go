package data

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aegis-run/aegis/internal/datalayer"
	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/db"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/tuple"
	datav1 "github.com/aegis-run/aegis/proto/aegis/data/v1"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

type API struct {
	datav1.UnimplementedDataServer

	mutator     datalayer.Mutator
	consistency datalayer.Consistency
}

func NewAPI(dl *datalayer.DataLayer) *API {
	return &API{
		mutator:     dl.Mutator,
		consistency: dl.Consistency,
	}
}

func (api *API) Mutate(
	ctx context.Context,
	req *datav1.MutateRequest,
) (res *datav1.MutateResponse, err error) {
	ctx, span := telemetry.Start(ctx, "data.api.mutate")
	defer telemetry.End(span, &err)

	if req == nil || len(req.GetMutations()) == 0 {
		return &datav1.MutateResponse{}, nil
	}

	mutations := make([]tuple.TupleMutation, len(req.GetMutations()))
	for i, m := range req.GetMutations() {
		mut, err := tuple.DecodeMutation(m)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to decode mutation: %v", err)
		}
		mutations[i] = mut
	}

	token, err := api.mutator.Mutate(ctx, mutations)
	if err != nil {
		if dlerr.IsAlreadyExists(err) {
			return nil, status.Errorf(codes.AlreadyExists, "relationship already exists: %v", err)
		}
		if dlerr.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "resource or target not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to mutate relationship tuples: %v", err)
	}

	return &datav1.MutateResponse{
		ConsistencyToken: consistency.Encode(token),
	}, nil
}

func (api *API) Query(
	ctx context.Context,
	req *datav1.QueryRequest,
) (res *datav1.QueryResponse, err error) {
	ctx, span := telemetry.Start(ctx, "data.api.query")
	defer telemetry.End(span, &err)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	filter, err := tuple.DecodeFilter(req.GetFilter())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decode query filter: %v", err)
	}

	var token consistency.Token
	if req.GetConsistency() != nil {
		requirement := consistency.DecodeRequirement(req.GetConsistency())
		token, err = api.consistency.Resolve(ctx, requirement)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid consistency requirement: %v", err)
		}
	}

	page, err := db.DecodePagination(req.GetPage())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decode pagination: %v", err)
	}

	q, cleanup, err := api.consistency.Querier(ctx, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create querier: %v", err)
	}
	defer cleanup()

	tuples, nextCursor, err := q.Query(ctx, filter, page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query relationship tuples: %v", err)
	}

	data := make([]*v1.Tuple, len(tuples))
	for i, t := range tuples {
		data[i] = t.Encode()
	}

	// Make sure telemetry gets the result count as an attribute
	telemetry.Attr(ctx, attribute.Int("query.results.count", len(tuples)))

	return &datav1.QueryResponse{
		Data:             data,
		Page:             page.Next(len(tuples), nextCursor).Encode(),
		ConsistencyToken: consistency.Encode(token),
	}, nil
}
