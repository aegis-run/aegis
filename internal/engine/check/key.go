package check

import (
	"strconv"

	"github.com/cespare/xxhash/v2"
)

type Key struct {
	hash uint64
}

func (k Key) String() string {
	return strconv.FormatUint(k.hash, 10)
}

func (req *Request) Key(meta Meta) Key {
	h := xxhash.New()

	write(h, "Check")
	write(h, meta.SchemaHash().Hex())
	if tok := meta.Consistency(); tok != nil {
		write(h, tok.String())
	} else {
		write(h, "")
	}
	write(h, req.Resource.Type)
	write(h, req.Resource.ID)
	write(h, req.Permission)
	write(h, req.Actor.Type)
	write(h, req.Actor.ID)

	return Key{hash: h.Sum64()}
}

//nolint:errcheck
func write(h *xxhash.Digest, s string) {
	// WriteString adds more data to d. It always returns len(s), nil.
	h.WriteString(strconv.Itoa(len(s)))
	h.WriteString(":")
	h.WriteString(s)
}
