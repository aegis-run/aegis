package memory

import (
	"context"
	"sync"

	"github.com/hashicorp/go-memdb"

	db "github.com/aegis-run/aegis/pkg/db/core"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

type database struct {
	sync.RWMutex
	db *memdb.MemDB
}

func New(ctx context.Context) (*database, error) {
	logger.DebugContext(ctx, "database.memory.starting")

	if err := Schema.Validate(); err != nil {
		return nil, err
	}

	db, err := memdb.NewMemDB(Schema)
	if err != nil {
		return nil, err
	}

	return &database{db: db}, nil
}

var _ db.DB = (*database)(nil)

func (*database) Engine() string { return "memory" }

func (db *database) Close() error {
	db.Lock()
	defer db.Unlock()
	db.db = nil

	return nil
}

func (*database) IsReady(context.Context) (bool, error) { return true, nil }
