package memory

import (
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
	"github.com/aegis-run/aegis/pkg/consistency"
	"github.com/aegis-run/aegis/pkg/db/postgres"
	"github.com/aegis-run/aegis/pkg/db/postgres/types"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestConsistencyToken(t *testing.T) {
	t1 := ConsistencyToken{Value: 10}
	t2 := ConsistencyToken{Value: 20}
	t3 := ConsistencyToken{Value: 10}

	res, err := t1.Compare(t2)
	assert.Err(t, err, nil)
	assert.Equal(t, res, -1)

	res, err = t2.Compare(t1)
	assert.Err(t, err, nil)
	assert.Equal(t, res, 1)

	res, err = t1.Compare(t3)
	assert.Err(t, err, nil)
	assert.Equal(t, res, 0)

	assert.Equal(t, t1.String(), "10")
	assert.Equal(t, t2.String(), "20")

	pgToken := postgres.ConsistencyToken{
		Value: types.XID8(pgtype.Uint64{Uint64: 100, Valid: true}),
	}
	_, err = t1.Compare(pgToken)
	assert.Err(t, err)

	// Test Protobuf Codec (Encode)
	pb := consistency.Encode(t1)
	assert.Equal(t, pb.Token, "10")

	// Test Protobuf Codec (Decode)
	decoded, err := DecodeConsistencyToken(pb)
	assert.Err(t, err, nil)
	assert.Equal(t, decoded.Value, uint64(10))

	// Test Decode Error Boundaries
	_, err = DecodeConsistencyToken(nil)
	assert.Err(t, err)

	_, err = DecodeConsistencyToken(&v1.ConsistencyToken{Token: ""})
	assert.Err(t, err)

	_, err = DecodeConsistencyToken(&v1.ConsistencyToken{Token: "not-an-int"})
	assert.Err(t, err)
}
