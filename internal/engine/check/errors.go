package check

import "errors"

var (
	// ErrMaxDepthExceeded is returned when the AST evaluation recursion limit is reached,
	// preventing infinite loops and stack overflows.
	ErrMaxDepthExceeded = errors.New("maximum evaluation depth exceeded")
	// ErrTypeNotFound is returned when a type is not found in the schema.
	ErrTypeNotFound = errors.New("type not found in schema")
	// ErrPermissionNotFound is returned when a permission is not found in the schema.
	ErrPermissionNotFound = errors.New("permission not found in schema")
)
