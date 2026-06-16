package memory

import (
	"fmt"
	"strconv"

	"github.com/aegis-run/aegis/pkg/consistency"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

// ConsistencyToken is the concrete implementation of the Token interface for the
// in-memory datastore. It wraps a simple monotonic uint64 sequence counter.
type ConsistencyToken struct {
	Value uint64
}

var _ consistency.Token = (*ConsistencyToken)(nil)

// Compare checks the ordering of two MemoryTokens.
// Returns an error if the other token is not a MemoryToken.
func (t ConsistencyToken) Compare(other consistency.Token) (int, error) {
	o, ok := other.(ConsistencyToken)
	if !ok {
		return 0, fmt.Errorf("incompatible token types: expected MemoryToken")
	}

	if t.Value < o.Value {
		return -1, nil
	}
	if t.Value > o.Value {
		return 1, nil
	}
	return 0, nil
}

// String serializes the ConsistencyToken to a base-10 string.
func (t ConsistencyToken) String() string {
	return strconv.FormatUint(t.Value, 10)
}

// DecodeConsistencyToken decodes a protobuf ConsistencyToken into a Memory ConsistencyToken.
func DecodeConsistencyToken(pb *v1.ConsistencyToken) (ConsistencyToken, error) {
	if pb == nil || pb.Token == "" {
		return ConsistencyToken{}, fmt.Errorf("empty consistency token")
	}

	val, err := strconv.ParseUint(pb.Token, 10, 64)
	if err != nil {
		return ConsistencyToken{}, fmt.Errorf("invalid memory token format: %w", err)
	}

	return ConsistencyToken{Value: val}, nil
}
