package data

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"github.com/aegis-run/aegis/internal/datalayer"
// 	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
// 	"github.com/aegis-run/aegis/pkg/assert"
// 	"github.com/aegis-run/aegis/pkg/consistency"
// 	"github.com/aegis-run/aegis/pkg/db"
// 	"github.com/aegis-run/aegis/pkg/tuple"
// 	datav1 "github.com/aegis-run/aegis/proto/aegis/data/v1"
// 	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
// )

// type mockMutator struct {
// 	mutateFn func(ctx context.Context, mutations []tuple.TupleMutation) (consistency.Token, error)
// }

// func (m *mockMutator) Mutate(ctx context.Context, mutations []tuple.TupleMutation) (consistency.Token, error) {
// 	return m.mutateFn(ctx, mutations)
// }

// type mockConsistency struct {
// 	resolveFn func(ctx context.Context, req consistency.Requirement) (consistency.Token, error)
// 	querierFn func(ctx context.Context, token consistency.Token) (db.Querier, func(), error)
// }

// func (c *mockConsistency) Resolve(ctx context.Context, req consistency.Requirement) (consistency.Token, error) {
// 	if c.resolveFn == nil {
// 		return nil, nil
// 	}
// 	return c.resolveFn(ctx, req)
// }

// func (c *mockConsistency) Querier(ctx context.Context, token consistency.Token) (db.Querier, func(), error) {
// 	return c.querierFn(ctx, token)
// }

// type mockQuerier struct {
// 	queryFn func(ctx context.Context, filter tuple.TupleFilter, page db.Pagination) ([]tuple.Tuple, int64, error)
// }

// func (q *mockQuerier) Query(ctx context.Context, filter tuple.TupleFilter, page db.Pagination) ([]tuple.Tuple, int64, error) {
// 	return q.queryFn(ctx, filter, page)
// }

// type mockToken struct {
// 	val string
// }

// func (t mockToken) Compare(other consistency.Token) (int, error) {
// 	o := other.(mockToken)
// 	if t.val < o.val {
// 		return -1, nil
// 	}
// 	if t.val > o.val {
// 		return 1, nil
// 	}
// 	return 0, nil
// }

// func (t mockToken) String() string {
// 	return t.val
// }

// func TestDataAPI_Mutate(t *testing.T) {
// 	t.Run("empty mutations returns empty response", func(t *testing.T) {
// 		mut := &mockMutator{}
// 		consist := &mockConsistency{}
// 		dl := &datalayer.DataLayer{
// 			Mutator:     mut,
// 			Consistency: consist,
// 		}
// 		api := NewAPI(dl)
// 		res, err := api.Mutate(context.Background(), &datav1.MutateRequest{})
// 		assert.Err(t, err, nil)
// 		assert.True(t, res != nil)
// 	})

// 	t.Run("successful mutate", func(t *testing.T) {
// 		mut := &mockMutator{
// 			mutateFn: func(ctx context.Context, mutations []tuple.TupleMutation) (consistency.Token, error) {
// 				assert.Equal(t, len(mutations), 1)
// 				assert.Equal(t, mutations[0].Op, tuple.OpWrite)
// 				assert.Equal(t, mutations[0].Tuple.Resource.Type, "organization")
// 				assert.Equal(t, mutations[0].Tuple.Resource.ID, "acme")
// 				return mockToken{val: "100"}, nil
// 			},
// 		}
// 		consist := &mockConsistency{}
// 		dl := &datalayer.DataLayer{
// 			Mutator:     mut,
// 			Consistency: consist,
// 		}

// 		api := NewAPI(dl)
// 		req := &datav1.MutateRequest{
// 			Mutations: []*datav1.TupleMutation{
// 				{
// 					Operation: datav1.TupleMutation_OPERATION_WRITE,
// 					Tuple: &v1.Tuple{
// 						Resource: &v1.Instance{Type: "organization", Id: "acme"},
// 						Relation: "owner",
// 						Subject: &v1.Subject{
// 							Instance: &v1.Instance{Type: "user", Id: "alice"},
// 						},
// 					},
// 				},
// 			},
// 		}

// 		res, err := api.Mutate(context.Background(), req)
// 		assert.Err(t, err, nil)
// 		assert.True(t, res != nil)
// 	})

// 	t.Run("mutate conflict translates to AlreadyExists", func(t *testing.T) {
// 		mut := &mockMutator{
// 			mutateFn: func(ctx context.Context, mutations []tuple.TupleMutation) (consistency.Token, error) {
// 				return nil, &dlerr.ErrAlreadyExists{
// 					Resource: "tuple",
// 					Err:      errors.New("conflict"),
// 				}
// 			},
// 		}
// 		consist := &mockConsistency{}
// 		dl := &datalayer.DataLayer{
// 			Mutator:     mut,
// 			Consistency: consist,
// 		}

// 		api := NewAPI(dl)
// 		req := &datav1.MutateRequest{
// 			Mutations: []*datav1.TupleMutation{
// 				{
// 					Operation: datav1.TupleMutation_OPERATION_WRITE,
// 					Tuple: &v1.Tuple{
// 						Resource: &v1.Instance{Type: "organization", Id: "acme"},
// 						Relation: "owner",
// 						Subject: &v1.Subject{
// 							Instance: &v1.Instance{Type: "user", Id: "alice"},
// 						},
// 					},
// 				},
// 			},
// 		}

// 		_, err := api.Mutate(context.Background(), req)
// 		assert.True(t, err != nil)
// 	})
// }

// func TestDataAPI_Query(t *testing.T) {
// 	t.Run("nil request returns error", func(t *testing.T) {
// 		mut := &mockMutator{}
// 		consist := &mockConsistency{}
// 		dl := &datalayer.DataLayer{
// 			Mutator:     mut,
// 			Consistency: consist,
// 		}
// 		api := NewAPI(dl)
// 		_, err := api.Query(context.Background(), nil)
// 		assert.True(t, err != nil)
// 	})

// 	t.Run("successful query without consistency strategy", func(t *testing.T) {
// 		mut := &mockMutator{}
// 		q := &mockQuerier{
// 			queryFn: func(
// 				ctx context.Context,
// 				filter tuple.TupleFilter,
// 				page db.Pagination,
// 			) ([]tuple.Tuple, int64, error) {
// 				assert.Equal(t, page.Limit, int32(50))
// 				assert.Equal(t, page.Cursor, int64(0))

// 				return []tuple.Tuple{
// 					{
// 						Resource: tuple.Instance{Type: "organization", ID: "acme"},
// 						Relation: "owner",
// 						Subject: tuple.Subject{
// 							Instance: tuple.Instance{Type: "user", ID: "alice"},
// 						},
// 					},
// 				}, 12345, nil
// 			},
// 		}
// 		consist := &mockConsistency{
// 			resolveFn: func(ctx context.Context, req consistency.Requirement) (consistency.Token, error) {
// 				return nil, nil
// 			},
// 			querierFn: func(ctx context.Context, token consistency.Token) (db.Querier, func(), error) {
// 				assert.True(t, token == nil)
// 				return q, func() {}, nil
// 			},
// 		}
// 		dl := &datalayer.DataLayer{
// 			Mutator:     mut,
// 			Consistency: consist,
// 		}

// 		api := NewAPI(dl)
// 		req := &datav1.QueryRequest{
// 			Filter: &datav1.TupleFilter{
// 				QueryTarget: &datav1.TupleFilter_Resource{
// 					Resource: &datav1.TupleFilter_InstanceFilter{Type: "organization", Id: "acme"},
// 				},
// 				Relation: "owner",
// 			},
// 		}

// 		res, err := api.Query(context.Background(), req)
// 		assert.Err(t, err, nil)
// 		assert.Equal(t, len(res.GetData()), 1)
// 		assert.Equal(t, res.GetData()[0].Resource.Type, "organization")
// 		assert.Equal(t, res.GetPage().GetCount(), uint32(1))
// 	})

// 	t.Run("fully consistent query uses resolved token", func(t *testing.T) {
// 		mut := &mockMutator{}
// 		q := &mockQuerier{
// 			queryFn: func(
// 				ctx context.Context,
// 				filter tuple.TupleFilter,
// 				page db.Pagination,
// 			) ([]tuple.Tuple, int64, error) {
// 				return nil, 0, nil
// 			},
// 		}
// 		consist := &mockConsistency{
// 			resolveFn: func(ctx context.Context, req consistency.Requirement) (consistency.Token, error) {
// 				if req.Strategy == consistency.StrategyFullyConsistent {
// 					return mockToken{val: "300"}, nil
// 				}
// 				return mockToken{val: req.Token}, nil
// 			},
// 			querierFn: func(ctx context.Context, token consistency.Token) (db.Querier, func(), error) {
// 				assert.Equal(t, token.String(), "300")
// 				return q, func() {}, nil
// 			},
// 		}
// 		dl := &datalayer.DataLayer{
// 			Mutator:     mut,
// 			Consistency: consist,
// 		}

// 		api := NewAPI(dl)
// 		req := &datav1.QueryRequest{
// 			Filter: &datav1.TupleFilter{
// 				QueryTarget: &datav1.TupleFilter_Resource{
// 					Resource: &datav1.TupleFilter_InstanceFilter{Type: "organization", Id: "acme"},
// 				},
// 			},
// 			Consistency: &v1.Consistency{
// 				Strategy: &v1.Consistency_FullyConsistent{
// 					FullyConsistent: true,
// 				},
// 			},
// 		}

// 		res, err := api.Query(context.Background(), req)
// 		assert.Err(t, err, nil)
// 		assert.True(t, res != nil)
// 	})
// }
