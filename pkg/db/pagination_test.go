package db

import (
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

func TestCursorCodec(t *testing.T) {
	// Test normal encode
	c := Cursor("12345").Encode()
	assert.True(t, c != nil)
	assert.Equal(t, c.Value, "MTIzNDU=")

	// Test normal decode
	val, err := DecodeCursor(c)
	assert.Err(t, err, nil)
	assert.Equal(t, val, Cursor("12345"))

	// Test nil encode
	cNil := Cursor("").Encode()
	assert.True(t, cNil == nil)

	// Test nil decode
	valNil, err := DecodeCursor(nil)
	assert.Err(t, err, nil)
	assert.Equal(t, valNil, Cursor(""))

	// Test empty value decode
	valEmpty, err := DecodeCursor(&v1.Cursor{Value: ""})
	assert.Err(t, err, nil)
	assert.Equal(t, valEmpty, Cursor(""))

	// Test invalid base64 decode
	_, err = DecodeCursor(&v1.Cursor{Value: "invalid-base64-!!!"})
	assert.Err(t, err)
}

func TestDecodePagination(t *testing.T) {
	u32 := func(v uint32) *uint32 { return &v }

	// Test nil pagination (now returns defaults)
	pNil, err := DecodePagination(nil)
	assert.Err(t, err, nil)
	assert.Equal(t, pNil.Limit, int32(50))
	assert.Equal(t, pNil.Cursor, int64(0))

	// Test default pagination (nil cursor)
	p, err := DecodePagination(&v1.Pagination{})
	assert.Err(t, err, nil)
	assert.Equal(t, p.Limit, int32(50))
	assert.Equal(t, p.Cursor, int64(0))

	// Test custom limit and nil cursor
	p, err = DecodePagination(&v1.Pagination{Limit: u32(100)})
	assert.Err(t, err, nil)
	assert.Equal(t, p.Limit, int32(100))
	assert.Equal(t, p.Cursor, int64(0))

	// Test custom limit and valid cursor
	c := Cursor("12345").Encode()
	p, err = DecodePagination(&v1.Pagination{Limit: u32(25), Cursor: c})
	assert.Err(t, err, nil)
	assert.Equal(t, p.Limit, int32(25))
	assert.Equal(t, p.Cursor, int64(12345))

	// Test invalid cursor
	p, err = DecodePagination(&v1.Pagination{
		Cursor: &v1.Cursor{Value: "invalid-base64"},
	})
	assert.Err(t, err)
}
