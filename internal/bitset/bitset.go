package bitset

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"math/bits"

	"github.com/kesopeso/cryptography-exercise/internal/assert"
)

// Bitset is a compact data structure that stores boolean values as individual bits
// within a byte slice. Each byte holds up to 8 boolean statuses.
type Bitset struct {
	data []byte
	size int
}

// NewBitset creates and returns an empty Bitset.
func NewBitset() *Bitset {
	return &Bitset{
		data: make([]byte, 0),
		size: 0,
	}
}

// Add appends a boolean value to the Bitset. A new byte is allocated when the
// current byte is full. Panics if the internal data length is inconsistent with
// the expected size.
// Returns the index of newly added value.
func (b *Bitset) Add(value bool) int {
	bitIndex := b.size % 8

	if bitIndex == 0 {
		b.data = append(b.data, 0x00)
	}

	byteIndex := b.size / 8
	assert.True(len(b.data) == byteIndex+1, fmt.Sprintf("inconsistent data, bytes count %d, size %d", len(b.data), b.size))

	b.size++
	err := b.Set(b.size-1, value)
	assert.NoError(err)

	return b.size - 1
}

// Set updates the boolean value at the given index. Returns an error if the
// index is out of bounds.
func (b *Bitset) Set(index int, value bool) error {
	if index < 0 || index > b.size-1 {
		return fmt.Errorf("index %d out of bounds, data size %d", index, b.size)
	}

	byteIndex := index / 8
	bitIndex := index % 8

	if value {
		b.data[byteIndex] |= (1 << bitIndex)
	} else {
		b.data[byteIndex] &= ^(1 << bitIndex)
	}

	return nil
}

// Encode compresses the bitset data with gzip and returns it as a base64-encoded string.
// A sentinel bit is appended after the last data bit to preserve the exact size
// for decoding.
func (b *Bitset) Encode() (string, error) {
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)

	_, err := gz.Write(b.getEncodeData())
	if err != nil {
		return "", err
	}

	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Decode decodes a base64-encoded, gzip-compressed string back into a Bitset.
// It locates the sentinel bit (highest set bit in the last byte), removes it,
// and uses the remaining bits to reconstruct the original data and size.
func Decode(encoded string) (*Bitset, error) {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	raw, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("encoded data is empty")
	}

	lastByte := raw[len(raw)-1]
	if lastByte == 0 {
		return nil, fmt.Errorf("missing sentinel bit in last byte")
	}

	// Find the sentinel: highest set bit in the last byte
	sentinelPos := bits.Len8(lastByte) - 1

	// Clear the sentinel bit
	raw[len(raw)-1] &= ^(1 << sentinelPos)

	// Size = all bits in preceding bytes + bits below the sentinel in the last byte
	size := (len(raw)-1)*8 + sentinelPos

	// If sentinel was at bit 0 of the last byte, the last byte was purely sentinel
	var data []byte
	if sentinelPos == 0 {
		data = raw[:len(raw)-1]
	} else {
		data = raw
	}

	return &Bitset{data: data, size: size}, nil
}

// getEncodeData returns a copy of the bitset data with a sentinel 1 bit appended
// after the last data bit. When decoding, the highest set bit in the last byte
// marks the sentinel — everything below it is actual data.
func (b *Bitset) getEncodeData() []byte {
	result := make([]byte, len(b.data))
	copy(result, b.data)

	remainder := b.size % 8
	if remainder == 0 {
		result = append(result, 0x01)
	} else {
		result[len(result)-1] |= (1 << remainder)
	}

	return result
}
