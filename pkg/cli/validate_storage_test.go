package cli

import (
	"os"
	"testing"

	"Protodoc/pkg/container"
)

// TestTR_012_ValidateDetectsRealDigestMismatch is T-0396's named integration
// test (TR-012, FR-104/FR-105). It proves validate's real storage_integrity_
// tree check is wired: a document whose segment bytes match their declared
// slot-digest (and whose ledger_root is T_S over those slots) passes; the same
// document with one segment's stored digest deliberately corrupted is INVALID
// with the storage_integrity_tree finding, naming that step. The stored
// ledger_root is never trusted — only the freshly recomputed T_S decides.
func TestTR_012_ValidateDetectsRealDigestMismatch(t *testing.T) {
	// (1) A correct document (assembleDoc seals real digests + ledger_root).
	segs := []fixtureSeg{
		{segType: container.SegmentTypeContent, body: realFrame(0x01, cliUnit(0x11))},
		{segType: container.SegmentTypeContent, body: realFrame(0x01, cliUnit(0x22))},
	}
	good := assembleDoc(t, defaultHeader(), segs, nil)
	goodPath := tempWrite(t, good)
	if res := runValidate([]string{goodPath}, nil); res.Status != "OK" {
		t.Fatalf("correct document: status=%s (%+v), want OK", res.Status, res.Findings)
	}

	// (2) Corrupt one SEGMENT's actual bytes AFTER sealing, so the recomputed
	// T_S over the (unchanged) slot-digests still equals the stored ledger_root
	// — but the digester's re-hash of the tampered segment octets no longer
	// matches that slot's declared digest, localising the tampered ordinal.
	// To make T_S itself diverge (the primary FR-104 signal), corrupt the
	// STORED slot-digest in the segment table instead: T_S recomputes over the
	// corrupted leaf and no longer equals the stored ledger_root.
	tampered := make([]byte, len(good))
	copy(tampered, good)
	// The segment table lives at [SegmentTableOffset, SegmentTableOffset+RegionSize).
	// Flip a byte inside the first slot's digest field. A slot is
	// SegmentTableSlotSize octets; the digest is its trailing 32 octets.
	slotBase := container.SegmentTableOffset
	// Corrupt the first slot's digest region (last 32 octets of the slot).
	digestByte := slotBase + container.SegmentTableSlotSize - 1
	tampered[digestByte] ^= 0xFF
	tamperedPath := tempWrite(t, tampered)

	res := runValidate([]string{tamperedPath}, nil)
	if res.Status != "INVALID" {
		t.Fatalf("tampered slot-digest: status=%s, want INVALID", res.Status)
	}
	foundStorageFinding := false
	for _, f := range res.Findings {
		if f.RuleID == "storage_integrity_tree" {
			foundStorageFinding = true
		}
	}
	if !foundStorageFinding {
		t.Errorf("tampered document must report the storage_integrity_tree finding, got %+v", res.Findings)
	}
}

func tempWrite(t *testing.T, data []byte) string {
	t.Helper()
	p := t.TempDir() + "/doc.pdl"
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatalf("tempWrite: %v", err)
	}
	return p
}
