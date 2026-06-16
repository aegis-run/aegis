package postgres

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
	"github.com/aegis-run/aegis/pkg/consistency"
	db "github.com/aegis-run/aegis/pkg/db/postgres"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

type Schema struct {
	db db.DB
}

func NewSchema(database db.DB) *Schema {
	return &Schema{db: database}
}

type writeResult struct {
	version schema.Version
	token   consistency.Token
}

func (s *Schema) Write(
	ctx context.Context,
	payload []byte,
) (ver schema.Version, token consistency.Token, err error) {
	hash := schema.Sum(payload)

	ctx, span := telemetry.Start(ctx, "datalayer.schema.write", trace.WithAttributes(
		attribute.String("schema.hash", hash.Hex()),
	))
	defer telemetry.End(span, &err)

	res, err := db.TxWithResult(ctx, s.db.RW(), func(tx db.DBTX) (writeResult, error) {
		row, err := db.Query.InsertSchemaVersion(ctx, tx, db.InsertSchemaVersionParams{
			Hash: hash,
			Data: payload,
		})
		if err == nil {
			logger.InfoContext(ctx, "datalayer.schema.written",
				"hash", hash.Hex(),
			)
			writtenAt := db.ConsistencyToken{Value: row.WrittenAt}
			return writeResult{
				version: schema.Version{
					Hash:      hash,
					Data:      append([]byte(nil), payload...),
					WrittenAt: writtenAt,
					CreatedAt: row.CreatedAt,
				},
				token: writtenAt,
			}, nil
		}

		translatedErr := translateError(err, "schema", hash.Hex())
		if dlerr.IsNotFound(translatedErr) {
			// Schema with this hash already exists — read it back.
			existing, readErr := s.ReadByHash(ctx, hash)
			if readErr != nil {
				return writeResult{}, readErr
			}
			return writeResult{version: existing, token: existing.WrittenAt}, nil
		}

		return writeResult{}, fmt.Errorf("failed to insert schema version: %w", err)
	})
	if err != nil {
		return schema.Version{}, nil, err
	}
	return res.version, res.token, nil
}

func (s *Schema) ReadLatest(ctx context.Context) (ver schema.Version, err error) {
	ctx, span := telemetry.Start(ctx, "datalayer.schema.read_latest")
	defer telemetry.End(span, &err)

	row, err := db.Query.GetLatestSchemaVersion(ctx, s.db.RW())
	if err != nil {
		translatedErr := translateError(err, "schema", "latest")
		if dlerr.IsNotFound(translatedErr) {
			return schema.Version{}, translatedErr
		}

		return schema.Version{}, fmt.Errorf("failed to get latest schema version: %w", err)
	}

	return schema.Version{
		Hash:      row.Hash,
		Data:      append([]byte(nil), row.Data...),
		WrittenAt: db.ConsistencyToken{Value: row.WrittenAt},
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *Schema) ReadByHash(ctx context.Context, hash schema.Hash) (ver schema.Version, err error) {
	ctx, span := telemetry.Start(ctx, "datalayer.schema.read", trace.WithAttributes(
		attribute.String("schema.hash", hash.Hex()),
	))
	defer telemetry.End(span, &err)

	row, err := db.Query.GetSchemaVersionByHash(ctx, s.db.RW(), hash)
	if err != nil {
		translatedErr := translateError(err, "schema", hash.Hex())
		if dlerr.IsNotFound(translatedErr) {
			return schema.Version{}, translatedErr
		}
		return schema.Version{}, fmt.Errorf("failed to get schema version by hash: %w", err)
	}

	return schema.Version{
		Hash:      row.Hash,
		Data:      append([]byte(nil), row.Data...),
		WrittenAt: db.ConsistencyToken{Value: row.WrittenAt},
		CreatedAt: row.CreatedAt,
	}, nil
}
