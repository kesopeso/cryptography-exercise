package bitset

import (
	"fmt"

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
func (s *Bitset) Add(value bool) {
	byteIndex := s.size / 8
	bitIndex := s.size % 8

	if bitIndex == 0 {
		s.data = append(s.data, 0)
	}

	assert.True(len(s.data) == byteIndex+1, fmt.Sprintf("data length missmatch, required: %d, actual: %d", byteIndex+1, len(s.data)))

	s.size++

	if !value {
		return
	}

	s.data[byteIndex] |= (1 << bitIndex)
}
