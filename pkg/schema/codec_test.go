package schema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

func TestEncodeDecode(t *testing.T) {
	// 1. Create a complete, mock schema
	mockSchema := &v1.Schema{
		Types: []*v1.TypeDefinition{
			{
				Name: "user",
			},
			{
				Name: "document",
				Relations: []*v1.Relation{
					{
						Name: "owner",
						Actors: []*v1.ActorType{
							{
								Actor: &v1.ActorType_Direct{
									Direct: "user",
								},
							},
						},
					},
				},
				Permissions: []*v1.Permission{
					{
						Name: "read",
						Expr: &v1.Expression{
							Kind: &v1.Expression_Term{
								Term: &v1.TermExpr{
									Term: &v1.TermExpr_SelfRef{
										SelfRef: &v1.SelfRef{
											Relation: "owner",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 2. Test Determinism: Encoding multiple times must produce identical bytes
	payload1, err := Encode(mockSchema)
	assert.Err(t, err, nil)

	payload2, err := Encode(mockSchema)
	assert.Err(t, err, nil)

	assert.Equal(t, payload1, payload2)

	// 3. Test Golden Byte Stability
	goldenPath := filepath.Join("testdata", "golden_schema.bin")

	// If running with update flag or golden file doesn't exist, create it
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
		assert.Err(t, err, nil)
		err = os.WriteFile(goldenPath, payload1, 0644)
		assert.Err(t, err, nil)
		t.Logf("Updated golden file: %s", goldenPath)
	}

	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Write initially if missing
			err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
			assert.Err(t, err, nil)
			err = os.WriteFile(goldenPath, payload1, 0644)
			assert.Err(t, err, nil)
			goldenBytes = payload1
		} else {
			assert.Err(t, err, nil)
		}
	}

	assert.Equal(t, payload1, goldenBytes)

	// 4. Test Round-Trip Compatibility (Decode)
	decodedSchema, err := Decode(payload1)
	assert.Err(t, err, nil)
	assert.Equal(t, decodedSchema.Types[1].Name, "document")

	// 5. Test Error Boundaries
	_, err = Encode(nil)
	assert.Err(t, err)

	_, err = Decode(nil)
	assert.Err(t, err)

	_, err = Decode([]byte("corrupted_protobuf_payload"))
	assert.Err(t, err)
}
