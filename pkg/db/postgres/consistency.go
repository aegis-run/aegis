package postgres

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/db/postgres/types"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

// ConsistencyToken is the concrete implementation of the Token interface for Postgres.
// It wraps the XID8 newtype to guarantee causal consistency.
type ConsistencyToken struct {
	Value types.XID8
}

var _ consistency.Token = (*ConsistencyToken)(nil)

// Compare checks the ordering of two ConsistencyTokens.
// Returns an error if the other token is not a ConsistencyToken.
func (t ConsistencyToken) Compare(other consistency.Token) (int, error) {
	o, ok := other.(ConsistencyToken)
	if !ok {
		return 0, fmt.Errorf("incompatible token types: expected ConsistencyToken")
	}

	if t.Value.Uint64 < o.Value.Uint64 {
		return -1, nil
	}
	if t.Value.Uint64 > o.Value.Uint64 {
		return 1, nil
	}
	return 0, nil
}

// String serializes the ConsistencyToken to a base-10 string.
func (t ConsistencyToken) String() string {
	return strconv.FormatUint(t.Value.Uint64, 10)
}

// DecodeConsistencyToken decodes a protobuf ConsistencyToken into a Postgres ConsistencyToken.
func DecodeConsistencyToken(pb *v1.ConsistencyToken) (ConsistencyToken, error) {
	if pb == nil || pb.Token == "" {
		return ConsistencyToken{}, fmt.Errorf("empty consistency token")
	}

	val, err := strconv.ParseUint(pb.Token, 10, 64)
	if err != nil {
		return ConsistencyToken{}, fmt.Errorf("invalid postgres token format: %w", err)
	}

	return ConsistencyToken{
		Value: pgtype.Uint64{Uint64: val, Valid: true},
	}, nil
}

// FromToken converts a generic consistency.Token interface into a concrete Postgres
// ConsistencyToken. It returns an error if the token is of an incompatible type. If the
// input token is nil, it returns a zero-value ConsistencyToken and no error.
func FromToken(token consistency.Token) (ConsistencyToken, error) {
	if token == nil {
		return ConsistencyToken{}, nil
	}

	switch t := token.(type) {
	case ConsistencyToken:
		return t, nil
	case *ConsistencyToken:
		if t == nil {
			return ConsistencyToken{}, fmt.Errorf("nil consistency token pointer")
		}
		return *t, nil
	default:
		return ConsistencyToken{}, fmt.Errorf(
			"incompatible token type: expected postgres.ConsistencyToken, got %T",
			token,
		)
	}
}
