package local

import (
	"context"

	"github.com/aegis-run/aegis/internal/engine/check"
)

// localDispatcher routes requests directly to the in-memory AST evaluator engine.
// It acts as the terminal node of the Dispatcher stack.
type localDispatcher struct {
	checker check.Checker
}

// NewDispatcher creates a new Local Dispatcher.
func NewDispatcher(checker check.Checker) *localDispatcher {
	return &localDispatcher{checker: checker}
}

// Check evaluates the recursion depth limit, decrements it, and passes
// the execution to the core evaluation engine.
func (d *localDispatcher) Check(
	ctx context.Context,
	req *check.Request,
	meta check.Meta,
) (*check.Response, error) {
	if meta.Depth() <= 0 {
		return nil, check.ErrMaxDepthExceeded
	}

	return d.checker.Check(ctx, req, meta.DecrementDepth())
}
