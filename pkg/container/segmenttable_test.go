package container

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

func fixtureSlot(seed byte) SegmentTableSlot {
	var digest pdlfmt.Digest256
	for i := range digest {
		digest[i] = seed + byte(i)
	}
	return SegmentTableSlot{
		SegmentType: SegmentTypeContent,
		Flags:       SlotFlagCoverageHint,
		Offset:      1048576 + uint64(seed)*4096,
		Length:      4096,
		FrameCount:  3,
		Digest:      digest,
	}
}

// TestFR_091_SegmentTableSlotFieldsExposed is T-0016's named test.
// Implements: FR-091.
func TestFR_091_SegmentTableSlotFieldsExposed(t *testing.T) {
	cases := []struct {
		name  string
		slots []SegmentTableSlot
	}{
		{"zero slots", nil},
		{"one slot", []SegmentTableSlot{fixtureSlot(1)}},
		{"max segments", func() []SegmentTableSlot {
			s := make([]SegmentTableSlot, MaxSegments)
			for i := range s {
				s[i] = fixtureSlot(byte(i))
			}
			return s
		}()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			enc, err := EncodeSegmentTable(c.slots, nil)
			if err != nil {
				t.Fatalf("EncodeSegmentTable: %v", err)
			}
			if len(enc) != SegmentTableRegionSize {
				t.Fatalf("encoded length = %d, want %d", len(enc), SegmentTableRegionSize)
			}

			table, err := DecodeSegmentTable(enc)
			if err != nil {
				t.Fatalf("DecodeSegmentTable: %v", err)
			}

			for i := 0; i < MaxSegments; i++ {
				var want SegmentTableSlot
				if i < len(c.slots) {
					want = c.slots[i]
				}
				got := table[i]
				// FR-091: kind, length and digest must each be readable
				// from the slot alone, with no separate index consulted.
				if got.SegmentType != want.SegmentType {
					t.Fatalf("slot %d SegmentType = %d, want %d", i, got.SegmentType, want.SegmentType)
				}
				if got.Length != want.Length {
					t.Fatalf("slot %d Length = %d, want %d", i, got.Length, want.Length)
				}
				if got.Digest != want.Digest {
					t.Fatalf("slot %d Digest = %x, want %x", i, got.Digest, want.Digest)
				}
				if got.Flags != want.Flags || got.Offset != want.Offset || got.FrameCount != want.FrameCount {
					t.Fatalf("slot %d = %+v, want %+v", i, got, want)
				}
			}

			// Re-encoding the fully decoded table must reproduce the
			// identical octets: a byte-exact round trip.
			reenc, err := EncodeSegmentTable(table[:], nil)
			if err != nil {
				t.Fatalf("re-EncodeSegmentTable: %v", err)
			}
			if string(reenc) != string(enc) {
				t.Fatal("re-encoded segment table octets differ from the original encoding")
			}
		})
	}
}

func TestDecodeSegmentTable_RejectsTruncatedRegion(t *testing.T) {
	_, err := DecodeSegmentTable(make([]byte, SegmentTableRegionSize-1))
	if err != ErrSegmentTableTruncated {
		t.Fatalf("got %v, want ErrSegmentTableTruncated", err)
	}
}

func TestEncodeSegmentTable_RejectsTooManySlots(t *testing.T) {
	slots := make([]SegmentTableSlot, MaxSegments+1)
	if _, err := EncodeSegmentTable(slots, nil); err == nil {
		t.Fatal("EncodeSegmentTable: want error for MaxSegments+1 slots, got nil")
	}
}

func TestDecodeSegmentTable_UnusedSlotIsAllZero(t *testing.T) {
	enc, err := EncodeSegmentTable(nil, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	for i, b := range enc {
		if b != 0 {
			t.Fatalf("all-unused segment table octet %d = %d, want 0", i, b)
		}
	}
}
