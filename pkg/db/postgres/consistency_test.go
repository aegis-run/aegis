package postgres_test

import (
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/db/memory"
	"github.com/aegis-run/aegis/pkg/db/postgres"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestConsistencyToken(t *testing.T) {
	t1 := postgres.ConsistencyToken{
		Value: pgtype.Uint64{Uint64: 100, Valid: true},
	}
	t2 := postgres.ConsistencyToken{
		Value: pgtype.Uint64{Uint64: 200, Valid: true},
	}
	t3 := postgres.ConsistencyToken{
		Value: pgtype.Uint64{Uint64: 100, Valid: true},
	}

	// Test Compare
	res, err := t1.Compare(t2)
	assert.Err(t, err, nil)
	assert.Equal(t, res, -1)

	res, err = t2.Compare(t1)
	assert.Err(t, err, nil)
	assert.Equal(t, res, 1)

	res, err = t1.Compare(t3)
	assert.Err(t, err, nil)
	assert.Equal(t, res, 0)

	// Test String Serialization
	assert.Equal(t, t1.String(), "100")
	assert.Equal(t, t2.String(), "200")

	// Test Incompatible Types Compare
	memToken := memory.ConsistencyToken{Value: 100}
	_, err = t1.Compare(memToken)
	assert.Err(t, err)

	// Test Protobuf Codec (Encode)
	pb := consistency.Encode(t1)
	assert.Equal(t, pb.Token, "100")

	// Test Protobuf Codec (Decode)
	decoded, err := postgres.DecodeConsistencyToken(pb)
	assert.Err(t, err, nil)
	assert.Equal(t, decoded.Value.Uint64, uint64(100))

	// Test Decode Error Boundaries
	_, err = postgres.DecodeConsistencyToken(nil)
	assert.Err(t, err)

	_, err = postgres.DecodeConsistencyToken(&v1.ConsistencyToken{Token: ""})
	assert.Err(t, err)

	_, err = postgres.DecodeConsistencyToken(&v1.ConsistencyToken{Token: "not-an-int"})
	assert.Err(t, err)
}

func TestFromToken(t *testing.T) {
	t1 := postgres.ConsistencyToken{
		Value: pgtype.Uint64{Uint64: 100, Valid: true},
	}

	// Test Nil Token
	token, err := postgres.FromToken(nil)
	assert.Err(t, err, nil)
	assert.Equal(t, token.Value.Uint64, uint64(0))

	// Test Value Type
	token, err = postgres.FromToken(t1)
	assert.Err(t, err, nil)
	assert.Equal(t, token.Value.Uint64, uint64(100))

	// Test Pointer Type
	token, err = postgres.FromToken(&t1)
	assert.Err(t, err, nil)
	assert.Equal(t, token.Value.Uint64, uint64(100))

	// Test Nil Pointer
	var nilPtr *postgres.ConsistencyToken
	_, err = postgres.FromToken(nilPtr)
	assert.Err(t, err)

	// Test Incompatible Type
	mockToken := memory.ConsistencyToken{Value: 100}
	_, err = postgres.FromToken(mockToken)
	assert.Err(t, err)
}
