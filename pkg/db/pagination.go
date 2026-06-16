package db

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

// Cursor is a Go newtype representing an opaque, database-agnostic keyset pagination cursor.
type Cursor string

// Encode takes a database-specific identifier Cursor, base64-encodes it
// to ensure opaqueness, and wraps it in a protobuf Cursor message.
func (c Cursor) Encode() *v1.Cursor {
	if c == "" {
		return nil
	}

	opaqueValue := base64.StdEncoding.EncodeToString([]byte(c))
	return &v1.Cursor{
		Value: opaqueValue,
	}
}

func (c Cursor) Value() (int64, error) {
	if c == "" {
		return 0, errors.New("cursor is empty")
	}

	value, err := strconv.ParseInt(string(c), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid page cursor value: %w", err)
	}

	return value, nil
}

// DecodeCursor unwraps a protobuf Cursor message, base64-decodes its opaque value,
// and returns the raw database-specific identifier Cursor newtype.
func DecodeCursor(c *v1.Cursor) (Cursor, error) {
	if c == nil || c.Value == "" {
		return "", nil
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		return "", fmt.Errorf("failed to decode opaque cursor: %w", err)
	}

	return Cursor(decodedBytes), nil
}

type Pagination struct {
	Cursor int64
	Limit  int32
}

func DecodePagination(pb *v1.Pagination) (Pagination, error) {
	if pb == nil {
		return Pagination{Limit: 50}, nil
	}

	limit := int32(50)
	if pb.GetLimit() > 0 {
		limit = int32(pb.GetLimit())
	}

	c, err := DecodeCursor(pb.Cursor)
	if err != nil {
		return Pagination{}, err
	}

	var cursor int64
	if c != "" {
		cursor, err = c.Value()
		if err != nil {
			return Pagination{}, err
		}
	}

	return Pagination{
		Cursor: cursor,
		Limit:  limit,
	}, nil
}

// Page represents the metadata for a single page of results.
type Page struct {
	Count      uint32
	NextCursor int64
	Total      uint32
}

// Next constructs a Page metadata object from the actual results and the current pagination config.
func (p Pagination) Next(itemsCount int, nextCursor int64) Page {
	res := Page{
		Count: uint32(itemsCount),
	}

	// Only provide a next cursor if we hit the limit and have a valid next identifier.
	if itemsCount > 0 && itemsCount == int(p.Limit) && nextCursor > 0 {
		res.NextCursor = nextCursor
	}

	return res
}

// Encode converts the domain Page metadata into a protobuf Page message.
func (p Page) Encode() *v1.Page {
	res := &v1.Page{
		Count: p.Count,
	}

	if p.NextCursor > 0 {
		res.NextCursor = Cursor(strconv.FormatInt(p.NextCursor, 10)).Encode()
	}

	if p.Total > 0 {
		total := p.Total
		res.Total = &total
	}

	return res
}
