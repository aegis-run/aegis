package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
	"github.com/aegis-run/aegis/pkg/assert"
)

func TestTranslateError(t *testing.T) {
	// Test nil
	assert.Equal(t, translateError(nil, "", ""), nil)

	// Test ErrNoRows -> ErrNotFound
	errNotFound := translateError(pgx.ErrNoRows, "user", "alice")
	assert.True(t, dlerr.IsNotFound(errNotFound))
	assert.Equal(t, dlerr.IsAlreadyExists(errNotFound), false)

	var notFoundStruct *dlerr.ErrNotFound
	assert.True(t, errors.As(errNotFound, &notFoundStruct))
	assert.Equal(t, notFoundStruct.Resource, "user")
	assert.Equal(t, notFoundStruct.ID, "alice")
	assert.Equal(t, notFoundStruct.DatalayerCode(), "NOT_FOUND")

	// Test PgError Unique Constraint Violation -> ErrAlreadyExists
	pgUniqueErr := &pgconn.PgError{
		Code:    "23505",
		Message: "duplicate key value violates unique constraint",
	}
	errAlreadyExists := translateError(pgUniqueErr, "workspace", "acme")
	assert.True(t, dlerr.IsAlreadyExists(errAlreadyExists))
	assert.Equal(t, dlerr.IsNotFound(errAlreadyExists), false)

	var alreadyExistsStruct *dlerr.ErrAlreadyExists
	assert.True(t, errors.As(errAlreadyExists, &alreadyExistsStruct))
	assert.Equal(t, alreadyExistsStruct.Resource, "workspace")
	assert.Equal(t, alreadyExistsStruct.ID, "acme")
	assert.Equal(t, alreadyExistsStruct.DatalayerCode(), "ALREADY_EXISTS")

	// Test Generic/Unhandled Error -> Unchanged
	genericErr := errors.New("something went wrong")
	translatedErr := translateError(genericErr, "foo", "bar")
	assert.Equal(t, translatedErr, genericErr)
}
