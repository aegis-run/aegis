package schema

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	hashRawLength = sha256.Size
	hashHexLength = hashRawLength * 2
)

type Hash [hashRawLength]byte

// Sum returns the SHA-256 hash of the data.
func Sum(data []byte) Hash {
	return sha256.Sum256(data)
}

// ParseHashHex parses a hex-encoded string into a Hash.
// Required for database compatibility as hashes are stored as hex in Postgres.
func ParseHashHex(value string) (Hash, error) {
	var h Hash
	if len(value) != hashHexLength {
		return h, fmt.Errorf("invalid hash hex length: expected %d, got %d", hashHexLength, len(value))
	}

	decoded, err := hex.DecodeString(value)
	if err != nil {
		return h, fmt.Errorf("failed to decode hash hex: %w", err)
	}

	copy(h[:], decoded)
	return h, nil
}

// ParseHashDigest parses a base64-encoded string (the API "digest") into a Hash.
func ParseHashDigest(value string) (Hash, error) {
	var h Hash
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return h, fmt.Errorf("failed to decode hash digest (base64): %w", err)
	}

	if len(decoded) != hashRawLength {
		return h, fmt.Errorf("invalid hash raw length: expected %d, got %d", hashRawLength, len(decoded))
	}

	copy(h[:], decoded)
	return h, nil
}

// Hex returns the hex-encoded string representation of the hash.
// Used for database storage and telemetry.
func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

// Digest returns the base64-encoded string representation of the hash.
// This is the primary representation used in the API.
func (h Hash) Digest() string {
	return base64.StdEncoding.EncodeToString(h[:])
}

// String returns the base64-encoded string representation.
func (h Hash) String() string {
	return h.Digest()
}

// Scan implements the sql.Scanner interface for database retrieval.
func (h *Hash) Scan(src any) error {
	switch v := src.(type) {
	case string:
		return h.decode(v)
	case []byte:
		return h.decode(string(v))
	default:
		return fmt.Errorf("cannot scan %T into Hash", src)
	}
}

// Value implements the driver.Valuer interface for database storage.
func (h Hash) Value() (driver.Value, error) {
	return h.Hex(), nil
}

func (h *Hash) decode(s string) (err error) {
	// Try parsing as hex first (standard DB storage format)
	if len(s) == hashHexLength {
		*h, err = ParseHashHex(s)
		return
	}
	// Fallback to digest (base64)
	*h, err = ParseHashDigest(s)
	return
}
