package schema

import (
	"crypto/rand"
	"strings"
	"testing"
	"testing/quick"

	"github.com/aegis-run/aegis/pkg/assert"
)

func TestHash_ParseHexAndDigest(t *testing.T) {
	t.Run("Valid Hex Parsing", func(t *testing.T) {
		// SHA-256 of empty bytes
		raw := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		h, err := ParseHashHex(raw)
		assert.Err(t, err, nil)
		assert.Equal(t, h.Hex(), raw)
	})

	t.Run("Invalid Hex Length", func(t *testing.T) {
		_, err := ParseHashHex("short_hex")
		assert.Err(t, err)
	})

	t.Run("Invalid Hex Characters", func(t *testing.T) {
		// 'g' is non-hex
		invalidHex := "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		_, err := ParseHashHex(invalidHex)
		assert.Err(t, err)
	})

	t.Run("Valid Digest Parsing", func(t *testing.T) {
		// base64 of empty bytes SHA-256
		rawDigest := "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
		h, err := ParseHashDigest(rawDigest)
		assert.Err(t, err, nil)
		assert.Equal(t, h.Digest(), rawDigest)
	})

	t.Run("Invalid Digest Characters", func(t *testing.T) {
		_, err := ParseHashDigest("invalid_base64_!@#")
		assert.Err(t, err)
	})

	t.Run("Invalid Digest Length", func(t *testing.T) {
		// base64 of "too_short"
		_, err := ParseHashDigest("dG9vX3Nob3J0")
		assert.Err(t, err)
	})
}

// Property-based testing proving encoding bijections and database drivers.
func TestHash_Properties(t *testing.T) {
	t.Run("Hex Round-trip Property", func(t *testing.T) {
		fHexRoundTrip := func(entropy [32]byte) bool {
			h := Hash(entropy)
			parsed, err := ParseHashHex(h.Hex())
			if err != nil {
				return false
			}
			return parsed == h
		}
		err := quick.Check(fHexRoundTrip, nil)
		assert.Err(t, err, nil)
	})

	t.Run("Digest Round-trip Property", func(t *testing.T) {
		fDigestRoundTrip := func(entropy [32]byte) bool {
			h := Hash(entropy)
			parsed, err := ParseHashDigest(h.Digest())
			if err != nil {
				return false
			}
			return parsed == h
		}
		err := quick.Check(fDigestRoundTrip, nil)
		assert.Err(t, err, nil)
	})

	t.Run("Invalid Hex String Parsing Failures", func(t *testing.T) {
		fHexInvalid := func(s string) bool {
			// Only check lengths that aren't the expected 64-char hex
			if len(s) == 64 {
				return true
			}
			_, err := ParseHashHex(s)
			return err != nil
		}
		err := quick.Check(fHexInvalid, nil)
		assert.Err(t, err, nil)
	})

	t.Run("Database Scanner/Valuer Round-trip Property", func(t *testing.T) {
		fSQLRoundTrip := func(entropy [32]byte) bool {
			h := Hash(entropy)
			val, err := h.Value()
			if err != nil {
				return false
			}
			strVal, ok := val.(string)
			if !ok || strVal != h.Hex() {
				return false
			}

			// Scan from string
			var scannedStr Hash
			if err := scannedStr.Scan(strVal); err != nil || scannedStr != h {
				return false
			}

			// Scan from []byte
			var scannedBytes Hash
			if err := scannedBytes.Scan([]byte(strVal)); err != nil || scannedBytes != h {
				return false
			}

			// Scan from digest (base64)
			var scannedDigest Hash
			if err := scannedDigest.Scan(h.Digest()); err != nil || scannedDigest != h {
				return false
			}

			return true
		}
		err := quick.Check(fSQLRoundTrip, nil)
		assert.Err(t, err, nil)
	})
}

func TestHash_SQLScannerValuer(t *testing.T) {
	// Seed a valid random hash
	entropy := make([]byte, 32)
	rand.Read(entropy)
	originalHash := Sum(entropy)

	t.Run("Scan String", func(t *testing.T) {
		var h Hash
		err := h.Scan(originalHash.Hex())
		assert.Err(t, err, nil)
		assert.Equal(t, h, originalHash)
	})

	t.Run("Scan Byte Slice", func(t *testing.T) {
		var h Hash
		err := h.Scan([]byte(originalHash.Hex()))
		assert.Err(t, err, nil)
		assert.Equal(t, h, originalHash)
	})

	t.Run("Scan Invalid Type", func(t *testing.T) {
		var h Hash
		type customStruct struct{ name string }
		err := h.Scan(customStruct{name: "some_data"})
		assert.Err(t, err)
	})

	t.Run("Valuer", func(t *testing.T) {
		val, err := originalHash.Value()
		assert.Err(t, err, nil)
		strVal, ok := val.(string)
		assert.True(t, ok)
		assert.Equal(t, strVal, originalHash.Hex())
	})
}

// Fuzz testing covering string parsers and db scanners over arbitrary fuzz corpus.

func FuzzHash_ParseHex(f *testing.F) {
	// Seed corpus with known valid and invalid hashes
	f.Add("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	f.Add("short_hex")
	f.Add("g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	f.Fuzz(func(t *testing.T, s string) {
		h, err := ParseHashHex(s)
		if err != nil {
			return
		}
		// Hex parsing is case-insensitive, but h.Hex() is always lowercase
		assert.Equal(t, h.Hex(), strings.ToLower(s))
	})
}

func FuzzHash_ParseDigest(f *testing.F) {
	f.Add("47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=")
	f.Add("invalid_base64_!@#")
	f.Add("dG9vX3Nob3J0")

	f.Fuzz(func(t *testing.T, s string) {
		h, err := ParseHashDigest(s)
		if err != nil {
			return
		}
		// Canonical round-trip must yield the exact parsed hash
		parsed, err := ParseHashDigest(h.Digest())
		assert.Err(t, err, nil)
		assert.Equal(t, parsed, h)
	})
}

func FuzzHash_Scan(f *testing.F) {
	f.Add("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	f.Add("47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=")
	f.Add("short_hex")
	f.Add("invalid_base64_!@#")

	f.Fuzz(func(t *testing.T, s string) {
		var h Hash
		err := h.Scan(s)
		if err != nil {
			return
		}

		// Scanning standard string must match scanning the byte slice
		var h2 Hash
		err = h2.Scan([]byte(s))
		assert.Err(t, err, nil)
		assert.Equal(t, h2, h)

		// Scanning back the canonical hex representation must succeed and match
		var h3 Hash
		err = h3.Scan(h.Hex())
		assert.Err(t, err, nil)
		assert.Equal(t, h3, h)
	})
}
