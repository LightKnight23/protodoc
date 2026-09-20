package main

import (
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
)

// TestSampleGeneratesValidBottomState confirms the regenerated PRONOM/IANA
// sample is a genuine, storage-integrity-correct minimal BOTTOM-state document:
// the DROID magic signature is intact at offset 0, the prefix is exactly the
// fixed size, the ring winner is unique and its ledger_root equals T_S over the
// (empty) full segment table (so `protodoc validate`'s storage_integrity_tree
// check passes), and there are zero segments.
func TestSampleGeneratesValidBottomState(t *testing.T) {
	img, err := buildSample()
	if err != nil {
		t.Fatalf("buildSample: %v", err)
	}

	prefixLen := container.SegmentTableOffset + container.SegmentTableRegionSize
	if len(img) != prefixLen {
		t.Fatalf("sample is %d octets, want %d", len(img), prefixLen)
	}

	// DROID PRONOM signature: "PDL1" + four zero octets at offset 0.
	wantMagic := []byte{0x50, 0x44, 0x4C, 0x31, 0x00, 0x00, 0x00, 0x00}
	for i, b := range wantMagic {
		if img[i] != b {
			t.Fatalf("magic byte %d = 0x%02X, want 0x%02X (breaks the PRONOM signature)", i, img[i], b)
		}
	}

	// The winning ring record is unique and carries the correct ledger_root.
	winner, _, err := container.SelectWinner(img[container.HeaderSize:container.HeaderSize+container.CommitRingSize], uint64(len(img)))
	if err != nil {
		t.Fatalf("SelectWinner (no unique winner => PD-RING-001): %v", err)
	}
	fullSlots := make([]container.SegmentTableSlot, container.MaxSegments)
	wantRoot, err := integrity.TSRoot(fullSlots)
	if err != nil {
		t.Fatalf("TSRoot: %v", err)
	}
	var got integrity.Digest
	copy(got[:], winner.LedgerRoot[:])
	if got != wantRoot {
		t.Errorf("ledger_root %x != T_S over empty table %x", winner.LedgerRoot[:4], wantRoot[:4])
	}
	if winner.SegmentCount != 0 {
		t.Errorf("segment_count = %d, want 0 (BOTTOM state)", winner.SegmentCount)
	}

	// The segment table decodes with zero live segments.
	table, err := container.DecodeSegmentTable(img[container.SegmentTableOffset:])
	if err != nil {
		t.Fatalf("DecodeSegmentTable: %v", err)
	}
	for i, slot := range table {
		if slot.SegmentType != container.SegmentTypeUnused {
			t.Errorf("slot %d is not unused (BOTTOM state should have no live segments)", i)
		}
	}
}
