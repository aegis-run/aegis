package datalayer

import (
	"context"

	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/db"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/tuple"
)

type Schema interface {
	Write(ctx context.Context, payload []byte) (schema.Version, consistency.Token, error)
	ReadLatest(ctx context.Context) (schema.Version, error)
	ReadByHash(ctx context.Context, hash schema.Hash) (schema.Version, error)
}

// type Data interface {
// 	Mutate(ctx context.Context, mutations []tuple.TupleMutation) (consistency.Token, error)
// 	Query(
// 		ctx context.Context,
// 		filter tuple.TupleFilter,
// 		token consistency.Token,
// 		page db.Pagination,
// 	) (tuples []tuple.Tuple, nextCursor int64, newToken consistency.Token, err error)
// }

type Mutator interface {
	Mutate(ctx context.Context, mutations []tuple.TupleMutation) (consistency.Token, error)
}

type Consistency interface {
	// Resolve maps a consistency requirement to a database token (as before).
	Resolve(ctx context.Context, req consistency.Requirement) (consistency.Token, error)

	// Querier returns a Querier bound to the resolved token.
	// It handles RO/RW routing, replica lag checks, and transaction management internally.
	Querier(ctx context.Context, token consistency.Token) (q db.Querier, cleanup func(), err error)
}
