package datalayer

import (
	"context"
	"fmt"

	dlpostgres "github.com/aegis-run/aegis/internal/datalayer/impl/postgres"
	db "github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/db/postgres"
)

type DataLayer struct {
	Schema      Schema
	Mutator     Mutator
	Consistency Consistency
}

func New(ctx context.Context, database db.DB, cfg db.Config) (*DataLayer, error) {
	switch database.Engine() {
	case string(db.POSTGRES):
		db, ok := database.(postgres.DB)
		if !ok {
			return nil, fmt.Errorf("database is not a postgres DB")
		}

		return &DataLayer{
			Schema:      dlpostgres.NewSchema(db),
			Mutator:     dlpostgres.NewTuples(db),
			Consistency: dlpostgres.NewConsistency(ctx, db, cfg.QuantizationWindow),
		}, nil
	case string(db.MEMORY):
		return nil, fmt.Errorf("in-memory datalayer not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported database engine: %s", database.Engine())
	}
}
