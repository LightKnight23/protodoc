package container

import (
	"testing"
)

// TestTR_006_EnumerateStoredUnitsFromBoundedPrefix is T-0021's named test.
// Implements: TR-006.
func TestTR_006_EnumerateStoredUnitsFromBoundedPrefix(t *testing.T) {
	const n = 5
	slots := make([]SegmentTableSlot, n)
	for i := range slots {
		slots[i] = fixtureSlot(byte(i))
		// Vary segment type across the fixture so the enumeration is
		// checked against more than one constant type value.
		switch i % 4 {
		case 0:
			slots[i].SegmentType = SegmentTypeContent
		case 1:
			slots[i].SegmentType = SegmentTypeResource
		case 2:
			slots[i].SegmentType = SegmentTypeHistory
		case 3:
			slots[i].SegmentType = SegmentTypeAttest
		}
	}

	segTable, err := EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}

	prefix := make([]byte, SegmentTableOffset+SegmentTableRegionSize)
	copy(prefix[SegmentTableOffset:], segTable)

	units, err := EnumerateStoredUnitsFromBoundedPrefix(prefix)
	if err != nil {
		t.Fatalf("EnumerateStoredUnitsFromBoundedPrefix: %v", err)
	}
	if len(units) != n {
		t.Fatalf("got %d units, want %d", len(units), n)
	}
	for i, u := range units {
		want := slots[i]
		if u.Ordinal != i {
			t.Errorf("unit %d: Ordinal = %d, want %d", i, u.Ordinal, i)
		}
		if u.SegmentType != want.SegmentType {
			t.Errorf("unit %d: SegmentType = %d, want %d", i, u.SegmentType, want.SegmentType)
		}
		if u.Length != want.Length {
			t.Errorf("unit %d: Length = %d, want %d", i, u.Length, want.Length)
		}
		if u.Digest != want.Digest {
			t.Errorf("unit %d: Digest = %x, want %x", i, u.Digest, want.Digest)
		}
	}
}

// TestTR_006_EnumerateStoredUnitsSkipsUnusedSlots confirms an all-zero
// unused slot (contracts/container.abnf S5.1) is never reported as a
// stored unit, only the populated ordinals are.
func TestTR_006_EnumerateStoredUnitsSkipsUnusedSlots(t *testing.T) {
	slots := []SegmentTableSlot{fixtureSlot(1)} // ordinal 0 populated, rest unused
	segTable, err := EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	table, err := DecodeSegmentTable(segTable)
	if err != nil {
		t.Fatalf("DecodeSegmentTable: %v", err)
	}
	units := EnumerateStoredUnits(table)
	if len(units) != 1 {
		t.Fatalf("got %d units, want 1", len(units))
	}
	if units[0].Ordinal != 0 {
		t.Fatalf("Ordinal = %d, want 0", units[0].Ordinal)
	}
}

// TestTR_006_EnumerateStoredUnitsFromBoundedPrefixTruncated confirms the
// bounded-prefix reader rejects an input shorter than the fixed
// 1,048,576-octet prefix rather than reading past it.
func TestTR_006_EnumerateStoredUnitsFromBoundedPrefixTruncated(t *testing.T) {
	_, err := EnumerateStoredUnitsFromBoundedPrefix(make([]byte, SegmentTableOffset))
	if err != ErrBoundedPrefixTruncated {
		t.Fatalf("err = %v, want ErrBoundedPrefixTruncated", err)
	}
}
