package schema

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aegis-run/aegis/internal/datalayer"
	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
	schemav1 "github.com/aegis-run/aegis/proto/aegis/schema/v1"
)

type API struct {
	schemav1.UnimplementedSchemaServer

	schema      datalayer.Schema
	consistency datalayer.Consistency
}

func NewAPI(dl *datalayer.DataLayer) *API {
	return &API{
		schema:      dl.Schema,
		consistency: dl.Consistency,
	}
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

	version, token, err := api.schema.Write(ctx, payload)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write schema version: %v", err)
	}

	telemetry.Attr(ctx, attribute.String("schema.hash", version.Hash.Hex()))

	return &schemav1.WriteResponse{
		Hash:      &schemav1.SchemaHash{Digest: version.Hash.Digest()},
		WrittenAt: consistency.Encode(token),
	}, nil
}

func (api *API) Read(
	ctx context.Context,
	req *schemav1.ReadRequest,
) (res *schemav1.ReadResponse, err error) {
	ctx, span := telemetry.Start(ctx, "schema.api.read")
	defer telemetry.End(span, &err)

	// Resolve the consistency requirement to determine RO vs RW routing.
	// The token itself is not passed into the schema datalayer — it only controls
	// which replica is used. Schema rows are not xid-visibility-gated.
	requirement := consistency.DecodeRequirement(req.GetConsistency())
	if _, err = api.consistency.Resolve(ctx, requirement); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid consistency requirement: %v", err)
	}

	version, err := api.readVersion(ctx, req)
	if err != nil {
		if dlerr.IsNotFound(err) {
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
		Schema:    sch,
		Hash:      &schemav1.SchemaHash{Digest: version.Hash.Digest()},
		WrittenAt: consistency.Encode(version.WrittenAt),
	}, nil
}

func (api *API) readVersion(
	ctx context.Context,
	req *schemav1.ReadRequest,
) (schema.Version, error) {
	if req.GetHash() == nil {
		return api.schema.ReadLatest(ctx)
	}

	hash, err := schema.ParseHashDigest(req.GetHash().GetDigest())
	if err != nil {
		return schema.Version{}, fmt.Errorf("invalid schema hash: %w", err)
	}

	telemetry.Attr(ctx, attribute.String("schema.requested_hash", hash.Hex()))

	return api.schema.ReadByHash(ctx, hash)
}
