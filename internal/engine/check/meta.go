package check

import (
	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/db"
	"github.com/aegis-run/aegis/pkg/schema"
)

// Meta tracks evaluation state to prevent infinite loops and limit recursion depth.
type Meta interface {
	Depth() int
	DecrementDepth() Meta
	SchemaHash() schema.Hash
	Consistency() consistency.Token
	Querier() db.Querier
}

type meta struct {
	depth   int
	hash    schema.Hash
	token   consistency.Token
	querier db.Querier
}

// NewMeta creates a new metadata instance with the specified initial depth and frozen schema hash.
func NewMeta(depth int, hash schema.Hash, token consistency.Token, querier db.Querier) Meta {
	return &meta{depth: depth, hash: hash, token: token, querier: querier}
}

// Depth returns the current remaining allowed recursion depth.
func (m *meta) Depth() int {
	return m.depth
}

// DecrementDepth returns a new Meta instance with the depth reduced by 1, preserving the hash.
func (m *meta) DecrementDepth() Meta {
	return &meta{depth: m.depth - 1, hash: m.hash, token: m.token, querier: m.querier}
}

// SchemaHash returns the frozen schema hash used for this evaluation tree.
func (m *meta) SchemaHash() schema.Hash {
	return m.hash
}

// Consistency returns the consistency token used for this evaluation tree.
func (m *meta) Consistency() consistency.Token {
	return m.token
}

// Querier returns the datalayer querier used for this evaluation tree.
func (m *meta) Querier() db.Querier {
	return m.querier
}
