package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

func Encode(pb *v1.Schema) ([]byte, error) {
	if pb == nil {
		return nil, fmt.Errorf("schema is required")
	}

	payload, err := proto.MarshalOptions{Deterministic: true}.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	return payload, nil
}

func Decode(payload []byte) (*v1.Schema, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("schema payload is empty")
	}

	var sch v1.Schema
	if err := proto.Unmarshal(payload, &sch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &sch, nil
}
