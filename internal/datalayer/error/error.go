package error

import (
	"errors"
	"fmt"
)

// DatalayerError defines the behavior for all typed datalayer errors.
type DatalayerError interface {
	error
	DatalayerCode() string
}

// ErrNotFound represents a resource not found database error.
type ErrNotFound struct {
	Resource string
	ID       string
	Err      error
}

// Ensure ErrNotFound implements DatalayerError.
//
//nolint:errcheck
var _ DatalayerError = (*ErrNotFound)(nil)

func (e *ErrNotFound) Error() string {
	if e.Resource != "" {
		if e.ID != "" {
			return fmt.Sprintf("%s '%s' not found: %v", e.Resource, e.ID, e.Err)
		}
		return fmt.Sprintf("%s not found: %v", e.Resource, e.Err)
	}
	return fmt.Sprintf("resource not found: %v", e.Err)
}

func (e *ErrNotFound) Unwrap() error {
	return e.Err
}

func (e *ErrNotFound) DatalayerCode() string {
	return "NOT_FOUND"
}

// ErrAlreadyExists represents a duplicate key / constraint violation database error.
type ErrAlreadyExists struct {
	Resource string
	ID       string
	Err      error
}

// Ensure ErrAlreadyExists implements DatalayerError.
//
//nolint:errcheck
var _ DatalayerError = (*ErrAlreadyExists)(nil)

func (e *ErrAlreadyExists) Error() string {
	if e.Resource == "" {
		return fmt.Sprintf("resource already exists: %v", e.Err)
	}

	if e.ID != "" {
		return fmt.Sprintf("%s '%s' already exists: %v", e.Resource, e.ID, e.Err)
	}

	return fmt.Sprintf("%s already exists: %v", e.Resource, e.Err)
}

func (e *ErrAlreadyExists) Unwrap() error {
	return e.Err
}

func (e *ErrAlreadyExists) DatalayerCode() string {
	return "ALREADY_EXISTS"
}

// IsNotFound checks if the error is of type *ErrNotFound.
func IsNotFound(err error) bool {
	var e *ErrNotFound
	return errors.As(err, &e)
}

// IsAlreadyExists checks if the error is of type *ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	var e *ErrAlreadyExists
	return errors.As(err, &e)
}
