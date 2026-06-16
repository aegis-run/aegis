package postgres

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/aegis-run/aegis/pkg/consistency"
	core "github.com/aegis-run/aegis/pkg/db"
	db "github.com/aegis-run/aegis/pkg/db/postgres"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
	"github.com/aegis-run/aegis/pkg/tuple"
)

// Tuples is the PostgreSQL concrete implementation of the datalayer Tuples interface.
type Tuples struct {
	db db.DB
}

// NewTuples creates a new Postgres Tuples datastore manager.
func NewTuples(database db.DB) *Tuples {
	return &Tuples{db: database}
}

// executeMutation runs a single write or delete operation in the transaction.
func (t *Tuples) executeMutation(
	ctx context.Context,
	tx db.DBTX,
	mut tuple.TupleMutation,
) error {
	switch mut.Op {
	case tuple.OpWrite:
		_, err := db.Query.InsertTuple(ctx, tx, db.InsertTupleParams{
			ResourceType:      mut.Tuple.Resource.Type,
			ResourceID:        mut.Tuple.Resource.ID,
			Relation:          mut.Tuple.Relation,
			SubjectType:       mut.Tuple.Subject.Instance.Type,
			SubjectID:         mut.Tuple.Subject.Instance.ID,
			SubjectPermission: mut.Tuple.Subject.Permission,
		})
		if err != nil {
			return fmt.Errorf("failed to insert tuple: %w",
				translateError(err, mut.Tuple.Resource.Type, mut.Tuple.Resource.ID),
			)
		}
	case tuple.OpDelete:
		err := db.Query.DeleteTuple(ctx, tx, db.DeleteTupleParams{
			ResourceType:      mut.Tuple.Resource.Type,
			ResourceID:        mut.Tuple.Resource.ID,
			Relation:          mut.Tuple.Relation,
			SubjectType:       mut.Tuple.Subject.Instance.Type,
			SubjectID:         mut.Tuple.Subject.Instance.ID,
			SubjectPermission: mut.Tuple.Subject.Permission,
		})
		if err != nil {
			return fmt.Errorf("failed to delete tuple: %w",
				translateError(err, mut.Tuple.Resource.Type, mut.Tuple.Resource.ID),
			)
		}
	}
	return nil
}

// Mutate executes a set of relationship mutations atomically in a transaction.
func (t *Tuples) Mutate(
	ctx context.Context,
	mutations []tuple.TupleMutation,
) (token consistency.Token, err error) {
	ctx, span := telemetry.Start(ctx, "datalayer.tuples.mutate", trace.WithAttributes(
		attribute.Int("mutations.count", len(mutations)),
	))
	defer telemetry.End(span, &err)

	if len(mutations) == 0 {
		return db.ConsistencyToken{}, nil
	}

	return db.TxWithResult(ctx, t.db.RW(), func(tx db.DBTX) (consistency.Token, error) {
		for _, mut := range mutations {
			if err = t.executeMutation(ctx, tx, mut); err != nil {
				return nil, err
			}
		}

		xid, err := db.Query.GetCurrentXactID(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current transaction ID: %w", err)
		}

		logger.InfoContext(ctx, "datalayer.tuples.mutated",
			"xid", xid.Uint64,
		)

		return db.ConsistencyToken{Value: xid}, nil
	})
}

// queryRunner implements datalayer.Querier
type queryRunner struct {
	tx db.DBTX
}

func (q *queryRunner) Query(
	ctx context.Context,
	filter tuple.TupleFilter,
	page core.Pagination,
) (tuples []tuple.Tuple, nextCursor int64, err error) {
	ctx, span := telemetry.Start(ctx, "datalayer.tuples.query", trace.WithAttributes(
		attribute.Int("query.target", int(filter.Target)),
		attribute.String("query.relation", filter.Relation),
		// attribute.String("db.mode", q.mode),
	))
	defer telemetry.End(span, &err)

	rows, err := db.Query.FindTuples(ctx, q.tx, db.FindTuplesParams{
		ResourceType:      filter.ResourceType,
		ResourceID:        filter.ResourceID,
		Relation:          filter.Relation,
		SubjectType:       filter.SubjectType,
		SubjectID:         filter.SubjectID,
		SubjectPermission: filter.SubjectPermission,
		LastPk:            page.Cursor,
		LimitVal:          page.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tuples: %w", err)
	}

	telemetry.Attr(ctx, attribute.Int("query.results.count", len(rows)))

	if len(rows) > 0 {
		nextCursor = rows[len(rows)-1].Pk
	}

	return mapRowsToTuples(rows), nextCursor, nil
}

// Query retrieves paginated relationship tuples matching a given query filter.
func (t *Tuples) Query(
	ctx context.Context,
	filter tuple.TupleFilter,
	token consistency.Token,
	page core.Pagination,
) (tuples []tuple.Tuple, nextCursor int64, newToken consistency.Token, err error) {
	ctx, span := telemetry.Start(ctx, "datalayer.tuples.query", trace.WithAttributes(
		attribute.Int("query.target", int(filter.Target)),
		attribute.String("query.relation", filter.Relation),
	))
	defer telemetry.End(span, &err)

	if token != nil {
		pgToken, err := db.FromToken(token)
		if err != nil {
			return nil, 0, nil, err
		}

		tuples, nextCursor, newToken, visible, err := t.checkAndQueryReplica(
			ctx, filter, pgToken, page,
		)
		if err != nil {
			return nil, 0, nil, err
		}
		if visible {
			return tuples, nextCursor, newToken, nil
		}
	}

	target := t.db.RO()
	if token != nil {
		target = t.db.RW()
	}

	return t.query(ctx, target, filter, page)
}

type replicaQueryResult struct {
	tuples     []tuple.Tuple
	nextCursor int64
	token      consistency.Token
	visible    bool
}

// checkAndQueryReplica starts a read-only transaction on the replica to check visibility
// and query the database if visible. Returns (tuples, nextCursor, token, visible, error).
func (t *Tuples) checkAndQueryReplica(
	ctx context.Context,
	filter tuple.TupleFilter,
	token db.ConsistencyToken,
	page core.Pagination,
) ([]tuple.Tuple, int64, consistency.Token, bool, error) {
	res, err := db.TxWithResult(ctx, t.db.RO(), func(tx db.DBTX) (replicaQueryResult, error) {
		visible, err := db.Query.IsXactVisible(ctx, tx, token.Value)
		if err != nil {
			return replicaQueryResult{}, fmt.Errorf(
				"failed to check transaction visibility: %w",
				err,
			)
		}

		if !visible {
			return replicaQueryResult{visible: false}, nil
		}

		rows, err := db.Query.FindTuples(ctx, tx, db.FindTuplesParams{
			ResourceType:      filter.ResourceType,
			ResourceID:        filter.ResourceID,
			Relation:          filter.Relation,
			SubjectType:       filter.SubjectType,
			SubjectID:         filter.SubjectID,
			SubjectPermission: filter.SubjectPermission,
			LastPk:            page.Cursor,
			LimitVal:          page.Limit,
		})
		if err != nil {
			return replicaQueryResult{}, fmt.Errorf("failed to query tuples: %w", err)
		}

		xmax, err := db.Query.GetSnapshotXmax(ctx, tx)
		if err != nil {
			return replicaQueryResult{}, fmt.Errorf("failed to get snapshot xmax: %w", err)
		}

		var nextCursor int64
		if len(rows) > 0 {
			nextCursor = rows[len(rows)-1].Pk
		}

		return replicaQueryResult{
			tuples:     mapRowsToTuples(rows),
			nextCursor: nextCursor,
			token:      db.ConsistencyToken{Value: xmax},
			visible:    true,
		}, nil
	})
	if err != nil {
		return nil, 0, nil, false, err
	}
	return res.tuples, res.nextCursor, res.token, res.visible, nil
}

type queryResult struct {
	tuples     []tuple.Tuple
	nextCursor int64
	token      consistency.Token
}

// query executes the read transaction against the specified target database.
func (t *Tuples) query(
	ctx context.Context,
	target *db.Replica,
	filter tuple.TupleFilter,
	page core.Pagination,
) ([]tuple.Tuple, int64, consistency.Token, error) {
	res, err := db.TxWithResult(ctx, target, func(tx db.DBTX) (queryResult, error) {
		rows, err := db.Query.FindTuples(ctx, tx, db.FindTuplesParams{
			ResourceType:      filter.ResourceType,
			ResourceID:        filter.ResourceID,
			Relation:          filter.Relation,
			SubjectType:       filter.SubjectType,
			SubjectID:         filter.SubjectID,
			SubjectPermission: filter.SubjectPermission,
			LastPk:            page.Cursor,
			LimitVal:          page.Limit,
		})
		if err != nil {
			return queryResult{}, fmt.Errorf("failed to query tuples: %w", err)
		}

		xmax, err := db.Query.GetSnapshotXmax(ctx, tx)
		if err != nil {
			return queryResult{}, fmt.Errorf("failed to get snapshot xmax: %w", err)
		}

		var nextCursor int64
		if len(rows) > 0 {
			nextCursor = rows[len(rows)-1].Pk
		}

		return queryResult{
			tuples:     mapRowsToTuples(rows),
			nextCursor: nextCursor,
			token:      db.ConsistencyToken{Value: xmax},
		}, nil
	})
	if err != nil {
		return nil, 0, nil, err
	}

	return res.tuples, res.nextCursor, res.token, nil
}

// mapRowsToTuples converts PostgreSQL internal rows to domain tuple structures.
func mapRowsToTuples(rows []db.Tuple) []tuple.Tuple {
	tuples := make([]tuple.Tuple, len(rows))
	for i, r := range rows {
		tuples[i] = tuple.Tuple{
			Resource: tuple.Instance{Type: r.ResourceType, ID: r.ResourceID},
			Relation: r.Relation,
			Subject: tuple.Subject{
				Instance:   tuple.Instance{Type: r.SubjectType, ID: r.SubjectID},
				Permission: r.SubjectPermission,
			},
		}
	}
	return tuples
}
