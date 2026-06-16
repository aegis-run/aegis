package types

import "github.com/jackc/pgx/v5/pgtype"

// XID8 is a Go newtype representing a 64-bit epoch-aware Postgres transaction ID.
// Mapped directly to PostgreSQL's xid8 database type in sqlc.yaml overrides.
type XID8 = pgtype.Uint64
