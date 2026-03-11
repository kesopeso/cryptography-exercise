package bitset

import "testing"

func TestNewBitset(t *testing.T) {
	bs := NewBitset()

	if bs.size != 0 {
		t.Errorf("size = %d, want 0", bs.size)
	}
	if len(bs.data) != 0 {
		t.Errorf("data length = %d, want 0", len(bs.data))
	}
}

func TestAdd_SingleTrue(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)

	if bs.size != 1 {
		t.Errorf("size = %d, want 1", bs.size)
	}
	if len(bs.data) != 1 {
		t.Errorf("data length = %d, want 1", len(bs.data))
	}
	if bs.data[0] != 0x01 {
		t.Errorf("data[0] = %08b, want 00000001", bs.data[0])
	}
}

func TestAdd_SingleFalse(t *testing.T) {
	bs := NewBitset()
	bs.Add(false)

	if bs.size != 1 {
		t.Errorf("size = %d, want 1", bs.size)
	}
	if len(bs.data) != 1 {
		t.Errorf("data length = %d, want 1", len(bs.data))
	}
	if bs.data[0] != 0x00 {
		t.Errorf("data[0] = %08b, want 00000000", bs.data[0])
	}
}

func TestAdd_FillOneByte(t *testing.T) {
	bs := NewBitset()
	// Add 8 true values — should fill exactly one byte (0xFF)
	for range 8 {
		bs.Add(true)
	}

	if bs.size != 8 {
		t.Errorf("size = %d, want 8", bs.size)
	}
	if len(bs.data) != 1 {
		t.Errorf("data length = %d, want 1", len(bs.data))
	}
	if bs.data[0] != 0xFF {
		t.Errorf("data[0] = %08b, want 11111111", bs.data[0])
	}
}

func TestAdd_AllocatesNewByteAfterEight(t *testing.T) {
	bs := NewBitset()
	for range 8 {
		bs.Add(false)
	}
	bs.Add(true)

	if bs.size != 9 {
		t.Errorf("size = %d, want 9", bs.size)
	}
	if len(bs.data) != 2 {
		t.Errorf("data length = %d, want 2", len(bs.data))
	}
	if bs.data[0] != 0x00 {
		t.Errorf("data[0] = %08b, want 00000000", bs.data[0])
	}
	if bs.data[1] != 0x01 {
		t.Errorf("data[1] = %08b, want 00000001", bs.data[1])
	}
}

func TestAdd_BitOrdering(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)
	bs.Add(false)
	bs.Add(true)
	bs.Add(false)
	bs.Add(false)
	bs.Add(false)
	bs.Add(false)
	bs.Add(false)

	if bs.data[0] != 0x05 {
		t.Errorf("data[0] = %08b, want 00000101", bs.data[0])
	}
}

func TestAdd_MultipleBytes(t *testing.T) {
	bs := NewBitset()
	for range 20 {
		bs.Add(true)
	}

	if bs.size != 20 {
		t.Errorf("size = %d, want 20", bs.size)
	}
	if len(bs.data) != 3 {
		t.Errorf("data length = %d, want 3", len(bs.data))
	}
	if bs.data[0] != 0xFF {
		t.Errorf("data[0] = %08b, want 11111111", bs.data[0])
	}
	if bs.data[1] != 0xFF {
		t.Errorf("data[1] = %08b, want 11111111", bs.data[1])
	}
	// 4 bits set in third byte: 00001111 = 0x0F
	if bs.data[2] != 0x0F {
		t.Errorf("data[2] = %08b, want 00001111", bs.data[2])
	}
}

func TestAdd_AllFalse(t *testing.T) {
	bs := NewBitset()
	for range 16 {
		bs.Add(false)
	}

	if bs.size != 16 {
		t.Errorf("size = %d, want 16", bs.size)
	}
	for i, b := range bs.data {
		if b != 0x00 {
			t.Errorf("data[%d] = %08b, want 00000000", i, b)
		}
	}
}
