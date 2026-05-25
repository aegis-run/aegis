package schema

import "time"

type Version struct {
	Hash      Hash
	Data      []byte
	CreatedAt time.Time
}
