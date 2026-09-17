package validate

import (
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
)

// TestFR_104_ValidateReportsStorageIntegrityTreeCheck is T-0366's named
// integration test. cli.md's validate stdout schema names a required check
// key storage_integrity_tree (data-model.md S7 step 6): T_S is recomputed
// fresh and compared against the winning CommitRingRecord.ledger_root. This
// test wires integrity.TSRoot (M08) into the check and asserts FR-104/FR-105:
//
//   - an untampered document reports storage_integrity_tree PASSED;
//   - a document with a tampered SegmentTableSlot digest reports it FAILED and
//     names the offending storage ordinal;
//   - the stored ledger_root is never trusted as the answer -- the fresh root
//     is always recomputed and is what decides the verdict.
func TestFR_104_ValidateReportsStorageIntegrityTreeCheck(t *testing.T) {
	// Build a small set of populated slots with distinct digests.
	slots := make([]container.SegmentTableSlot, 4)
	for i := range slots {
		slots[i] = container.SegmentTableSlot{
			SegmentType: container.SegmentTypeContent,
			Offset:      uint64(1 << 20),
			Length:      256,
			FrameCount:  1,
		}
		for b := range slots[i].Digest {
			slots[i].Digest[b] = byte(i*7 + b) // deterministic, distinct per slot
		}
	}

	// The authentic ledger_root is T_S over the untampered slots.
	authentic, err := integrity.TSRoot(slots)
	if err != nil {
		t.Fatalf("TSRoot: %v", err)
	}

	// A digester that returns each slot's own stored digest as "recomputed"
	// for the authentic slots (i.e. no slot is tampered relative to its
	// segment content). We capture the authentic digests up front so that
	// after we tamper the stored slot the digester still reports the true
	// segment digest, exposing the mismatch.
	authenticDigests := make([][32]byte, len(slots))
	for i := range slots {
		authenticDigests[i] = slots[i].Digest
	}
	digester := func(ordinal int, _ container.SegmentTableSlot) (digest [32]byte, ok bool) {
		if ordinal < 0 || ordinal >= len(authenticDigests) {
			return [32]byte{}, false
		}
		return authenticDigests[ordinal], true
	}

	// (1) Untampered: check passes.
	res, err := CheckStorageIntegrityTree(slots, authentic, digester)
	if err != nil {
		t.Fatalf("check (untampered): %v", err)
	}
	if !res.Passed {
		t.Fatalf("untampered document: storage_integrity_tree reported FAILED, want PASSED")
	}
	if len(res.OffendingSlots) != 0 {
		t.Errorf("untampered document: unexpected offending slots %v", res.OffendingSlots)
	}
	if StorageIntegrityFinding(res) != nil {
		t.Error("untampered document: expected no pipeline finding")
	}

	// (2) Tamper slot ordinal 2's stored digest. T_S now diverges from the
	// authentic ledger_root, and the tampered slot is the offending one.
	const tamperedOrdinal = 2
	tampered := make([]container.SegmentTableSlot, len(slots))
	copy(tampered, slots)
	tampered[tamperedOrdinal].Digest[0] ^= 0xFF

	res2, err := CheckStorageIntegrityTree(tampered, authentic, digester)
	if err != nil {
		t.Fatalf("check (tampered): %v", err)
	}
	if res2.Passed {
		t.Fatalf("tampered document: storage_integrity_tree reported PASSED, want FAILED")
	}
	if len(res2.OffendingSlots) != 1 || res2.OffendingSlots[0] != tamperedOrdinal {
		t.Errorf("tampered document: offending slots = %v, want [%d]", res2.OffendingSlots, tamperedOrdinal)
	}
	f := StorageIntegrityFinding(res2)
	if f == nil || f.Step != StepTSRecompute || f.RuleID != StorageIntegrityCheckKey {
		t.Fatalf("tampered document: expected a step-6 %s finding, got %+v", StorageIntegrityCheckKey, f)
	}

	// The fresh root, not the stored one, decides: even though we passed the
	// authentic ledger_root as "stored", the recomputed T_S over the tampered
	// slots differs from it, which is exactly why the check fails.
	if res2.FreshRoot == res2.StoredRoot {
		t.Error("fresh T_S root should differ from stored ledger_root after tampering")
	}
}
