package tuple

import (
	"errors"
	"fmt"
)

var (
	ErrInstanceRequired = errors.New("instance is required")
	ErrSubjectRequired  = errors.New("subject is required")
	ErrTupleRequired    = errors.New("tuple is required")
	ErrMutationRequired = errors.New("mutation is required")
	ErrFilterRequired   = errors.New("filter is required")
	ErrInvalidTarget    = errors.New("filter must specify a query target (resource or subject)")
)

// ErrUnsupportedMutationOp is returned when an invalid or unspecified operation is encountered.
type ErrUnsupportedMutationOp struct {
	Op string
}

func (e ErrUnsupportedMutationOp) Error() string {
	return fmt.Sprintf("unsupported or unspecified mutation operation: %s", e.Op)
}
