package schema

import (
	"time"

	"github.com/aegis-run/aegis/pkg/consistency"
)

type Version struct {
	Hash      Hash
	Data      []byte
	WrittenAt consistency.Token
	CreatedAt time.Time
}
