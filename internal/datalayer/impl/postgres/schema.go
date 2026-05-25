package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
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

func (s *Schema) Write(ctx context.Context, payload []byte) (ver schema.Version, err error) {
	hash := schema.Sum(payload)

	ctx, span := telemetry.Start(ctx, "datalayer.schema.write", trace.WithAttributes(
		attribute.String("schema.hash", hash.Hex()),
	))
	defer telemetry.End(span, &err)

	row, err := db.Query.InsertSchemaVersion(ctx, s.db.RW(), db.InsertSchemaVersionParams{
		Hash: hash,
		Data: payload,
	})
	if err == nil {
		logger.InfoContext(ctx, "datalayer.schema.written",
			"hash", hash.Hex(),
		)
		return schema.Version{
			Hash:      hash,
			Data:      append([]byte(nil), payload...),
			CreatedAt: row.CreatedAt,
		}, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return s.ReadByHash(ctx, hash)
	}

	return schema.Version{}, fmt.Errorf("failed to insert schema version: %w", err)
}

func (s *Schema) ReadLatest(ctx context.Context) (ver schema.Version, err error) {
	ctx, span := telemetry.Start(ctx, "datalayer.schema.read_latest")
	defer telemetry.End(span, &err)

	row, err := db.Query.GetLatestSchemaVersion(ctx, s.db.RW())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return schema.Version{}, dlerr.ErrNotFound
		}

		return schema.Version{}, fmt.Errorf("failed to get latest schema version: %w", err)
	}

	return schema.Version{
		Hash:      row.Hash,
		Data:      append([]byte(nil), row.Data...),
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
		if errors.Is(err, sql.ErrNoRows) {
			return schema.Version{}, dlerr.ErrNotFound
		}
		return schema.Version{}, fmt.Errorf("failed to get schema version by hash: %w", err)
	}

	return schema.Version{
		Hash:      row.Hash,
		Data:      append([]byte(nil), row.Data...),
		CreatedAt: row.CreatedAt,
	}, nil
}
