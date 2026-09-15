package container

import (
	"errors"
	"testing"
)

// rawSlotBytes builds a SegmentTableSlotSize-octet raw slot with the given
// slot-segment-type octet and otherwise-fixed field bytes, for exercising
// decodeSegmentTableSlot's enum-closure check in isolation from encode.
func rawSlotBytes(segmentType byte) []byte {
	b := make([]byte, SegmentTableSlotSize)
	b[0] = segmentType
	b[1] = SlotFlagCoverageHint
	for i := 2; i < SegmentTableSlotSize; i++ {
		b[i] = byte(i)
	}
	return b
}

// TestTR_008_SegmentTypeEnumClosedNoExecutingKind is T-0023's named test.
// Implements: TR-008.
//
// It enumerates every possible slot-segment-type octet (0-255) and asserts
// the enum is closed to exactly 5 values: the 0 "unused" sentinel plus the
// 4 real segment kinds {CONTENT, RESOURCE, HISTORY, ATTEST}. Every other
// octet is a structural rejection at decode, before any field is
// dereferenced or interpreted.
func TestTR_008_SegmentTypeEnumClosedNoExecutingKind(t *testing.T) {
	closed := map[byte]bool{
		SegmentTypeUnused:   true,
		SegmentTypeContent:  true,
		SegmentTypeResource: true,
		SegmentTypeHistory:  true,
		SegmentTypeAttest:   true,
	}

	for v := 0; v <= 255; v++ {
		segType := byte(v)
		t.Run("", func(t *testing.T) {
			slot, err := decodeSegmentTableSlot(rawSlotBytes(segType))
			if closed[segType] {
				if err != nil {
					t.Fatalf("segment-type %d is in the closed enum, decode failed: %v", segType, err)
				}
				if slot.SegmentType != segType {
					t.Fatalf("segment-type %d: decoded SegmentType = %d", segType, slot.SegmentType)
				}
				return
			}
			if err == nil {
				t.Fatalf("segment-type %d is outside the closed enum, decode succeeded, want rejection", segType)
			}
			if !errors.Is(err, ErrInvalidSegmentType) {
				t.Fatalf("segment-type %d: err = %v, want ErrInvalidSegmentType", segType, err)
			}
		})
	}
}

// TestTR_008_DecodePathIdenticalAcrossKinds asserts decode takes no
// per-kind branch: given identical non-type octets, every one of the 4
// real segment kinds (and the unused sentinel) decodes through the exact
// same field-copy path, differing in the result only by SegmentType.
// This is the "no executing kind" half of TR-008: no kind triggers a
// distinct exec/interpret/dereference code path at decode time.
func TestTR_008_DecodePathIdenticalAcrossKinds(t *testing.T) {
	kinds := []byte{
		SegmentTypeUnused,
		SegmentTypeContent,
		SegmentTypeResource,
		SegmentTypeHistory,
		SegmentTypeAttest,
	}

	var reference SegmentTableSlot
	for i, k := range kinds {
		slot, err := decodeSegmentTableSlot(rawSlotBytes(k))
		if err != nil {
			t.Fatalf("segment-type %d: decode failed: %v", k, err)
		}
		slot.SegmentType = 0 // neutralise the one field allowed to differ
		if i == 0 {
			reference = slot
			continue
		}
		if slot != reference {
			t.Fatalf("segment-type %d decoded slot (type zeroed) = %+v, want %+v (identical decode path)", k, slot, reference)
		}
	}
}
