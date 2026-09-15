package container

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// fixtureRingRecord builds a CommitRingRecord with every field set to a
// distinct, deterministic non-zero pattern (parent aside), so a field
// transposition bug in Encode/Decode's fixed offsets would fail the
// round-trip comparison instead of passing by coincidence.
func fixtureRingRecord(seq uint64, parent StateID) CommitRingRecord {
	fill := func(seed byte) (b [32]byte) {
		for i := range b {
			b[i] = seed + byte(i)
		}
		return b
	}
	var idx [IndexRouteBuckets]uint16
	for i := range idx {
		idx[i] = uint16(i)*7 + uint16(seq)
	}
	return CommitRingRecord{
		Sequence:             seq,
		LedgerLength:         CommitRingSize + seq,
		LedgerRoot:           fill(0x10),
		SegmentCount:         uint16(seq),
		StateID:              StateID(fill(0x20)),
		FrontmatterDigest:    fill(0x30),
		SegmentTableDigest:   fill(0x40),
		TCRoot:               fill(0x50),
		ParentStateID:        parent,
		RetentionPoint:       uint16(seq),
		CompactionGeneration: uint32(seq),
		IndexRoute:           idx,
		StructureDigest:      fill(0x60),
	}
}

// TestFR_004_ParentStateIdRoundTrip is T-0007's named test.
// Implements: FR-004, FR-117.
func TestFR_004_ParentStateIdRoundTrip(t *testing.T) {
	var derived StateID
	if _, err := rand.Read(derived[:]); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	var ring [CommitRingSlots]CommitRingRecord
	ring[0] = fixtureRingRecord(0, StateID{}) // root state: empty parent-state-id
	for i := 1; i < CommitRingSlots; i++ {
		ring[i] = fixtureRingRecord(uint64(i), derived) // derived state: non-empty parent-state-id
	}

	enc := EncodeCommitRing(&ring, nil)
	if len(enc) != CommitRingSize {
		t.Fatalf("encoded ring length = %d, want %d", len(enc), CommitRingSize)
	}

	dec, err := DecodeCommitRing(enc)
	if err != nil {
		t.Fatalf("DecodeCommitRing: %v", err)
	}
	if *dec != ring {
		t.Fatalf("round-trip mismatch:\n got  %+v\n want %+v", *dec, ring)
	}

	// Root state (slot 0): parent-state-id is 32 zero octets (contracts/
	// container.abnf S3 parent-state-id NORMATIVE comment).
	if dec[0].ParentStateID != (StateID{}) {
		t.Errorf("root state parent-state-id = %x, want all-zero", dec[0].ParentStateID)
	}

	// Derived state (slot 1): parent-state-id names the predecessor and
	// round-trips exactly.
	if dec[1].ParentStateID == (StateID{}) {
		t.Fatal("derived state parent-state-id is all-zero, want non-empty predecessor id")
	}
	if dec[1].ParentStateID != derived {
		t.Errorf("derived state parent-state-id = %x, want %x", dec[1].ParentStateID, derived)
	}

	// Fixed-offset check: parent-state-id occupies slot octets [182,214)
	// within slot 1 (contracts/container.abnf S3 ring-fields layout comment).
	slot1 := enc[RingSlotSize : 2*RingSlotSize]
	gotAtOffset := slot1[roParentStateID : roParentStateID+32]
	if !bytes.Equal(gotAtOffset, derived[:]) {
		t.Errorf("parent-state-id not at fixed offset [182,214): got %x, want %x", gotAtOffset, derived[:])
	}

	// Re-encoding a decoded slot must reproduce the identical octets
	// (byte-exact round-trip, all 7 slots, not just slot 0/1).
	reenc := EncodeCommitRing(dec, nil)
	if !bytes.Equal(reenc, enc) {
		t.Fatal("re-encoded commit ring octets differ from the original encoding")
	}
}

func TestDecodeCommitRingRecord_RejectsTruncatedSlot(t *testing.T) {
	_, err := DecodeCommitRingRecord(make([]byte, RingSlotSize-1))
	if err != ErrRingSlotTruncated {
		t.Fatalf("got %v, want ErrRingSlotTruncated", err)
	}
}

func TestDecodeCommitRing_RejectsTruncatedRing(t *testing.T) {
	_, err := DecodeCommitRing(make([]byte, CommitRingSize-1))
	if err != ErrCommitRingTruncated {
		t.Fatalf("got %v, want ErrCommitRingTruncated", err)
	}
}
