package error_test

import (
	"errors"
	"testing"

	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
	"github.com/aegis-run/aegis/pkg/assert"
)

func TestDatalayerErrors(t *testing.T) {
	// Test ErrNotFound
	errNotFound := &dlerr.ErrNotFound{
		Resource: "user",
		ID:       "alice",
		Err:      errors.New("db error"),
	}

	assert.True(t, dlerr.IsNotFound(errNotFound))
	assert.Equal(t, dlerr.IsAlreadyExists(errNotFound), false)
	assert.Equal(t, errNotFound.DatalayerCode(), "NOT_FOUND")

	var notFoundStruct *dlerr.ErrNotFound
	assert.True(t, errors.As(errNotFound, &notFoundStruct))
	assert.Equal(t, notFoundStruct.Resource, "user")
	assert.Equal(t, notFoundStruct.ID, "alice")

	// Test ErrAlreadyExists
	errAlreadyExists := &dlerr.ErrAlreadyExists{
		Resource: "workspace",
		ID:       "acme",
		Err:      errors.New("db error"),
	}

	assert.True(t, dlerr.IsAlreadyExists(errAlreadyExists))
	assert.Equal(t, dlerr.IsNotFound(errAlreadyExists), false)
	assert.Equal(t, errAlreadyExists.DatalayerCode(), "ALREADY_EXISTS")

	var alreadyExistsStruct *dlerr.ErrAlreadyExists
	assert.True(t, errors.As(errAlreadyExists, &alreadyExistsStruct))
	assert.Equal(t, alreadyExistsStruct.Resource, "workspace")
	assert.Equal(t, alreadyExistsStruct.ID, "acme")
}
