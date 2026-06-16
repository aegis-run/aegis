package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aegis-run/aegis/pkg/assert"
	"github.com/aegis-run/aegis/pkg/db/postgres"
)

// mockTx is a helper type that implements the pgx.Tx interface for unit testing TracedTx.
type mockTx struct {
	pgx.Tx
	execCalled     bool
	queryCalled    bool
	queryRowCalled bool
	commitCalled   bool
	rollbackCalled bool
	commitErr      error
	rollbackErr    error
}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	m.execCalled = true
	return pgconn.CommandTag{}, nil
}

func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.queryCalled = true
	return nil, nil
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.queryRowCalled = true
	return nil
}

func (m *mockTx) Commit(ctx context.Context) error {
	m.commitCalled = true
	return m.commitErr
}

func (m *mockTx) Rollback(ctx context.Context) error {
	m.rollbackCalled = true
	return m.rollbackErr
}

func TestTracedTxDelegationAndTelemetry(t *testing.T) {
	mock := &mockTx{}
	tx := postgres.WrapTxForTest(mock, "primary")

	ctx := context.Background()

	// Test Exec delegation
	_, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "alice")
	assert.Err(t, err, nil)
	assert.Equal(t, mock.execCalled, true)

	// Test Query delegation
	_, err = tx.Query(ctx, "SELECT name FROM users")
	assert.Err(t, err, nil)
	assert.Equal(t, mock.queryCalled, true)

	// Test QueryRow delegation
	_ = tx.QueryRow(ctx, "SELECT name FROM users LIMIT 1")
	assert.Equal(t, mock.queryRowCalled, true)
}

func TestTracedTxCommitError(t *testing.T) {
	mockErr := errors.New("db commit error")
	mock := &mockTx{commitErr: mockErr}
	tx := postgres.WrapTxForTest(mock, "primary")

	err := tx.Commit(context.Background())
	assert.Err(t, err)

	var errCommit *postgres.ErrTxCommit
	assert.True(t, errors.As(err, &errCommit))
	assert.True(t, errors.Is(err, mockErr))
	assert.Equal(t, mock.commitCalled, true)
}

func TestTracedTxRollbackError(t *testing.T) {
	mockErr := errors.New("db rollback error")
	mock := &mockTx{rollbackErr: mockErr}
	tx := postgres.WrapTxForTest(mock, "primary")

	err := tx.Rollback(context.Background())
	assert.Err(t, err)

	var errRollback *postgres.ErrTxRollback
	assert.True(t, errors.As(err, &errRollback))
	assert.True(t, errors.Is(err, mockErr))
	assert.Equal(t, mock.rollbackCalled, true)
}

func TestTxBeginError(t *testing.T) {
	// A nil/empty replica should trigger begin error
	r := &postgres.Replica{}

	err := postgres.Tx(context.Background(), r, func(tx postgres.DBTX) error {
		return nil
	})
	assert.Err(t, err)

	var errBegin *postgres.ErrTxBegin
	assert.True(t, errors.As(err, &errBegin))
}
