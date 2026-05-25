package datalayer

import (
	"fmt"

	dlpostgres "github.com/aegis-run/aegis/internal/datalayer/impl/postgres"
	db "github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/db/postgres"
)

type DataLayer struct {
	Schema Schema
}

func New(database db.DB) (*DataLayer, error) {
	switch database.Engine() {
	case string(db.POSTGRES):
		db, ok := database.(postgres.DB)
		if !ok {
			return nil, fmt.Errorf("database is not a postgres DB")
		}
		return &DataLayer{
			Schema: dlpostgres.NewSchema(db),
		}, nil
	case string(db.MEMORY):
		return nil, fmt.Errorf("in-memory datalayer not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported database engine: %s", database.Engine())
	}
}
