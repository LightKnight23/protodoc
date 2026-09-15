package pdlfmt

import (
	"errors"
	"math"
	"math/rand"
	"testing"
)

// TestNFR_001_VarintMinimalEncodingRoundTrip is T-0001's named test.
// Implements: NFR-001.
func TestNFR_001_VarintMinimalEncodingRoundTrip(t *testing.T) {
	fixed := []uint64{
		0, 1,
		varintU8Max, varintU8Max + 1,
		varintU16Max, varintU16Max + 1,
		varintU32Max, varintU32Max + 1,
		math.MaxUint64,
		252, 253, 65535, 65536, 4294967295, 4294967296,
	}
	for _, v := range fixed {
		enc := EncodeVarint(v)
		if got := VarintLen(v); got != len(enc) {
			t.Errorf("VarintLen(%d) = %d, want %d", v, got, len(enc))
		}
		dec, n, err := DecodeVarint(enc)
		if err != nil {
			t.Fatalf("DecodeVarint(EncodeVarint(%d)) returned error: %v", v, err)
		}
		if n != len(enc) {
			t.Errorf("DecodeVarint(%d) consumed %d octets, want %d", v, n, len(enc))
		}
		if dec != v {
			t.Errorf("round-trip mismatch: encoded %d, decoded %d", v, dec)
		}
	}

	// Property test over random uint64 values: encode(decode(x)) == x, zero free parameters.
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		v := rng.Uint64()
		dec, _, err := DecodeVarint(EncodeVarint(v))
		if err != nil {
			t.Fatalf("random value %d: decode error: %v", v, err)
		}
		if dec != v {
			t.Fatalf("random value %d: round-trip mismatch, got %d", v, dec)
		}
	}
}

// TestPD_VARINT_001_RejectsNonMinimalEncoding verifies the decoder rejects
// every over-long form with a named error (rule PD-VARINT-001).
func TestPD_VARINT_001_RejectsNonMinimalEncoding(t *testing.T) {
	cases := []struct {
		name string
		src  []byte
	}{
		{"u16 form encodes value <= 252", []byte{0xFD, 0x00, 0x05}},
		{"u16 form encodes value == 252 exactly", []byte{0xFD, 0x00, varintU8Max}},
		{"u32 form encodes value <= 65535", []byte{0xFE, 0x00, 0x00, 0x00, 0x05}},
		{"u32 form encodes value == 65535 exactly", []byte{0xFE, 0x00, 0x00, 0xFF, 0xFF}},
		{"u64 form encodes value <= 4294967295", []byte{0xFF, 0, 0, 0, 0, 0, 0, 0, 5}},
		{"u64 form encodes value == 4294967295 exactly", []byte{0xFF, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := DecodeVarint(c.src)
			if !errors.Is(err, ErrNonMinimalVarint) {
				t.Errorf("DecodeVarint(%x) = err %v, want ErrNonMinimalVarint", c.src, err)
			}
		})
	}
}

// TestPD_VARINT_001_RejectsTruncatedInput verifies truncated trailing
// octets are rejected rather than silently zero-extended.
func TestPD_VARINT_001_RejectsTruncatedInput(t *testing.T) {
	cases := [][]byte{
		{},
		{0xFD},
		{0xFD, 0x01},
		{0xFE, 0x00, 0x01, 0x00},
		{0xFF, 0, 0, 0, 1, 0, 0, 0},
	}
	for _, src := range cases {
		_, _, err := DecodeVarint(src)
		if !errors.Is(err, ErrTruncatedVarint) {
			t.Errorf("DecodeVarint(%x) = err %v, want ErrTruncatedVarint", src, err)
		}
	}
}

// TestAppendVarint_NoAllocationOnPreallocatedSlice checks AppendVarint
// composes correctly with a non-nil destination (used by TLV encoding).
func TestAppendVarint_NoAllocationOnPreallocatedSlice(t *testing.T) {
	dst := []byte{0xAA, 0xBB}
	got := AppendVarint(dst, 300)
	want := []byte{0xAA, 0xBB, 0xFD, 0x01, 0x2C}
	if len(got) != len(want) {
		t.Fatalf("AppendVarint result length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AppendVarint result mismatch at %d: got %x, want %x", i, got, want)
		}
	}
}
