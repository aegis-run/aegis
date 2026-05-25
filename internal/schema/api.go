package schema

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aegis-run/aegis/internal/datalayer"
	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
	schemav1 "github.com/aegis-run/aegis/proto/aegis/schema/v1"
)

type API struct {
	schemav1.UnimplementedSchemaServer

	store datalayer.Schema
}

func NewAPI(store datalayer.Schema) *API {
	return &API{store: store}
}

func (api *API) Write(
	ctx context.Context,
	req *schemav1.WriteRequest,
) (res *schemav1.WriteResponse, err error) {
	ctx, span := telemetry.Start(ctx, "schema.api.write")
	defer telemetry.End(span, &err)

	payload, err := schema.Encode(req.GetSchema())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid schema payload: %v", err)
	}

	version, err := api.store.Write(ctx, payload)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write schema version: %v", err)
	}

	telemetry.Attr(ctx, attribute.String("schema.hash", version.Hash.Hex()))

	return &schemav1.WriteResponse{
		Hash: &schemav1.SchemaHash{Digest: version.Hash.Digest()},
	}, nil
}

func (api *API) Read(
	ctx context.Context,
	req *schemav1.ReadRequest,
) (res *schemav1.ReadResponse, err error) {
	ctx, span := telemetry.Start(ctx, "schema.api.read")
	defer telemetry.End(span, &err)

	version, err := api.readVersion(ctx, req)
	if err != nil {
		if errors.Is(err, dlerr.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "schema version not found")
		}
		return nil, status.Errorf(
			codes.Internal,
			"failed to read schema version: %v",
			err,
		)
	}

	telemetry.Attr(ctx, attribute.String("schema.hash", version.Hash.Hex()))

	sch, err := schema.Decode(version.Data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to decode stored schema: %v",
			err,
		)
	}

	return &schemav1.ReadResponse{
		Schema: sch,
		Hash:   &schemav1.SchemaHash{Digest: version.Hash.Digest()},
	}, nil
}

func (api *API) readVersion(
	ctx context.Context,
	req *schemav1.ReadRequest,
) (schema.Version, error) {
	if req.GetHash() == nil {
		return api.store.ReadLatest(ctx)
	}

	hash, err := schema.ParseHashDigest(req.GetHash().GetDigest())
	if err != nil {
		return schema.Version{}, fmt.Errorf("invalid schema hash: %w", err)
	}

	telemetry.Attr(ctx, attribute.String("schema.requested_hash", hash.Hex()))

	return api.store.ReadByHash(ctx, hash)
}
