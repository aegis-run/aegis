package validator

import (
	"fmt"

	"buf.build/go/protovalidate"
)

type Validator interface {
	protovalidate.Validator
}

func New(opts ...protovalidate.ValidatorOption) (Validator, error) {
	v, err := protovalidate.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to construct validator: %w", err)
	}

	return v, nil
}
