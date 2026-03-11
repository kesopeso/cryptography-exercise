package bitset

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"testing"
)

func TestNewBitset(t *testing.T) {
	bs := NewBitset()

	if bs.size != 0 {
		t.Errorf("size = %d, want 0", bs.size)
	}
	if len(bs.data) != 0 {
		t.Errorf("data length = %d, want 0", len(bs.data))
	}
}

func TestAdd_ReturnsCorrectIndex(t *testing.T) {
	bs := NewBitset()

	for i := range 20 {
		got := bs.Add(i%2 == 0)
		if got != i {
			t.Errorf("Add() = %d, want %d", got, i)
		}
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

func TestSet_TrueToFalse(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)
	bs.Add(true)
	bs.Add(true)

	err := bs.Set(1, false)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Bits: 1,0,1 = 0x05
	if bs.data[0] != 0x05 {
		t.Errorf("data[0] = %08b, want 00000101", bs.data[0])
	}
}

func TestSet_FalseToTrue(t *testing.T) {
	bs := NewBitset()
	bs.Add(false)
	bs.Add(false)
	bs.Add(false)

	err := bs.Set(1, true)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Bits: 0,1,0 = 0x02
	if bs.data[0] != 0x02 {
		t.Errorf("data[0] = %08b, want 00000010", bs.data[0])
	}
}

func TestSet_SecondByte(t *testing.T) {
	bs := NewBitset()
	for range 16 {
		bs.Add(false)
	}

	err := bs.Set(10, true)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if bs.data[0] != 0x00 {
		t.Errorf("data[0] = %08b, want 00000000", bs.data[0])
	}
	// Bit 10 is index 2 in the second byte: 00000100 = 0x04
	if bs.data[1] != 0x04 {
		t.Errorf("data[1] = %08b, want 00000100", bs.data[1])
	}
}

func TestSet_NegativeIndex(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)

	err := bs.Set(-1, true)
	if err == nil {
		t.Fatal("expected error for negative index, got nil")
	}
}

func TestSet_IndexOutOfBounds(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)
	bs.Add(true)

	err := bs.Set(2, true)
	if err == nil {
		t.Fatal("expected error for out of bounds index, got nil")
	}
}

func TestSet_EmptyBitset(t *testing.T) {
	bs := NewBitset()

	err := bs.Set(0, true)
	if err == nil {
		t.Fatal("expected error for empty bitset, got nil")
	}
}

// decodeEncoded is a test helper that base64-decodes and gunzips an encoded string.
func decodeEncoded(t *testing.T, encoded string) []byte {
	t.Helper()
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	defer gz.Close()
	data, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("gzip read error: %v", err)
	}
	return data
}

func TestEncode_UsesGzipCompression(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)

	encoded, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}

	if len(compressed) < 2 {
		t.Fatal("compressed data too short to contain gzip header")
	}
	if compressed[0] != 0x1f || compressed[1] != 0x8b {
		t.Errorf("gzip magic header = [%#x, %#x], want [0x1f, 0x8b]", compressed[0], compressed[1])
	}
}

func TestEncode_ReturnsValidBase64(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)

	encoded, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	_, err = base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("result is not valid base64: %v", err)
	}
}

func TestEncode_EmptyBitset(t *testing.T) {
	bs := NewBitset()

	encoded, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	data := decodeEncoded(t, encoded)
	// Empty bitset: no data bytes, just sentinel byte 0x01
	if len(data) != 1 || data[0] != 0x01 {
		t.Errorf("data = %v, want [00000001]", data)
	}
}

func TestEncode_SentinelBitPartialByte(t *testing.T) {
	bs := NewBitset()
	// 3 bits: true, false, true -> data = 00000101
	// With sentinel at bit 3: 00001101
	bs.Add(true)
	bs.Add(false)
	bs.Add(true)

	encoded, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	data := decodeEncoded(t, encoded)
	if len(data) != 1 {
		t.Fatalf("data length = %d, want 1", len(data))
	}
	if data[0] != 0x0D {
		t.Errorf("data[0] = %08b, want 00001101", data[0])
	}
}

func TestEncode_SentinelBitFullByte(t *testing.T) {
	bs := NewBitset()
	// 8 bits all true -> data = 0xFF
	// Sentinel in new byte: 0x01
	for range 8 {
		bs.Add(true)
	}

	encoded, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	data := decodeEncoded(t, encoded)
	if len(data) != 2 {
		t.Fatalf("data length = %d, want 2", len(data))
	}
	if data[0] != 0xFF {
		t.Errorf("data[0] = %08b, want 11111111", data[0])
	}
	if data[1] != 0x01 {
		t.Errorf("data[1] = %08b, want 00000001", data[1])
	}
}

func TestEncode_DoesNotMutateOriginalData(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)
	bs.Add(false)
	bs.Add(true)

	originalByte := bs.data[0]

	_, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if bs.data[0] != originalByte {
		t.Errorf("data[0] mutated: got %08b, want %08b", bs.data[0], originalByte)
	}
	if bs.size != 3 {
		t.Errorf("size mutated: got %d, want 3", bs.size)
	}
}

func TestEncode_DifferentDataProducesDifferentOutput(t *testing.T) {
	bs1 := NewBitset()
	bs1.Add(true)
	bs1.Add(false)

	bs2 := NewBitset()
	bs2.Add(false)
	bs2.Add(true)

	enc1, err := bs1.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	enc2, err := bs2.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if enc1 == enc2 {
		t.Error("different bitsets produced identical encoded output")
	}
}

func encodeHelper(t *testing.T, bs *Bitset) string {
	t.Helper()
	encoded, err := bs.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	return encoded
}

func assertBitsetEqual(t *testing.T, got *Bitset, wantData []byte, wantSize int) {
	t.Helper()
	if got.size != wantSize {
		t.Errorf("size = %d, want %d", got.size, wantSize)
	}
	if len(got.data) != len(wantData) {
		t.Errorf("data length = %d, want %d", len(got.data), len(wantData))
		return
	}
	for i := range wantData {
		if got.data[i] != wantData[i] {
			t.Errorf("data[%d] = %08b, want %08b", i, got.data[i], wantData[i])
		}
	}
}

func TestDecode_EmptyBitset(t *testing.T) {
	bs := NewBitset()
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{}, 0)
}

func TestDecode_SingleTrue(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{0x01}, 1)
}

func TestDecode_SingleFalse(t *testing.T) {
	bs := NewBitset()
	bs.Add(false)
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{0x00}, 1)
}

func TestDecode_PartialByte(t *testing.T) {
	bs := NewBitset()
	// 3 bits: true, false, true
	bs.Add(true)
	bs.Add(false)
	bs.Add(true)
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{0x05}, 3)
}

func TestDecode_FullByte(t *testing.T) {
	bs := NewBitset()
	for range 8 {
		bs.Add(true)
	}
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{0xFF}, 8)
}

func TestDecode_MultipleBytes(t *testing.T) {
	bs := NewBitset()
	for range 20 {
		bs.Add(true)
	}
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{0xFF, 0xFF, 0x0F}, 20)
}

func TestDecode_AllFalse(t *testing.T) {
	bs := NewBitset()
	for range 16 {
		bs.Add(false)
	}
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, []byte{0x00, 0x00}, 16)
}

func TestDecode_RoundTripPreservesData(t *testing.T) {
	bs := NewBitset()
	values := []bool{true, false, true, true, false, false, true, false, true}
	for _, v := range values {
		bs.Add(v)
	}
	encoded := encodeHelper(t, bs)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertBitsetEqual(t, decoded, bs.data, bs.size)
}

func TestDecode_InvalidBase64(t *testing.T) {
	_, err := Decode("not-valid-base64!@#")
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestDecode_InvalidGzip(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("not gzip data"))
	_, err := Decode(encoded)
	if err == nil {
		t.Fatal("expected error for invalid gzip, got nil")
	}
}

func TestGet_ReturnsCorrectValues(t *testing.T) {
	bs := NewBitset()
	values := []bool{true, false, true, false, false, true, false, true}
	for _, v := range values {
		bs.Add(v)
	}

	for i, want := range values {
		got, err := bs.Get(i)
		if err != nil {
			t.Fatalf("Get(%d) unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("Get(%d) = %v, want %v", i, got, want)
		}
	}
}

func TestGet_AcrossMultipleBytes(t *testing.T) {
	bs := NewBitset()
	values := []bool{true, false, true, false, true, false, true, false, false, true}
	for _, v := range values {
		bs.Add(v)
	}

	for i, want := range values {
		got, err := bs.Get(i)
		if err != nil {
			t.Fatalf("Get(%d) unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("Get(%d) = %v, want %v", i, got, want)
		}
	}
}

func TestGet_NegativeIndex(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)

	_, err := bs.Get(-1)
	if err == nil {
		t.Fatal("expected error for negative index, got nil")
	}
}

func TestGet_IndexOutOfBounds(t *testing.T) {
	bs := NewBitset()
	bs.Add(true)

	_, err := bs.Get(1)
	if err == nil {
		t.Fatal("expected error for out of bounds index, got nil")
	}
}

func TestGet_EmptyBitset(t *testing.T) {
	bs := NewBitset()

	_, err := bs.Get(0)
	if err == nil {
		t.Fatal("expected error for empty bitset, got nil")
	}
}
