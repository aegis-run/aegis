package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	dlerr "github.com/aegis-run/aegis/internal/datalayer/error"
)

// translateError maps pgx/postgres driver errors into standardized,
// database-agnostic datalayer domain-specific typed errors.
func translateError(err error, resource string, id string) error {
	if err == nil {
		return nil
	}

	// Map native pgx.ErrNoRows to the generic dlerr.ErrNotFound
	if errors.Is(err, pgx.ErrNoRows) {
		return &dlerr.ErrNotFound{
			Resource: resource,
			ID:       id,
			Err:      err,
		}
	}

	// Map PostgreSQL constraint violation codes to the generic dlerr.ErrAlreadyExists
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505": // unique_violation
			return &dlerr.ErrAlreadyExists{
				Resource: resource,
				ID:       id,
				Err:      err,
			}
		}
	}

	return err
}
