package db

import (
	"context"

	"github.com/aegis-run/aegis/pkg/tuple"
)

type Querier interface {
	Query(
		ctx context.Context,
		filter tuple.TupleFilter,
		page Pagination,
	) (tuples []tuple.Tuple, nextCursor int64, err error)
}
