package cli

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
)

// TestTR_012_RedactVerbReadsRealFile is T-0379's named integration test
// (TR-012, DEFECT-2026-09-19 fix). The production redact backend (realRedactRun,
// wired by init) reads a real file: designating a subtree by its
// content-addressed key omits that segment's body from the output (its bytes
// never survive) and declares the omission; an absent file is rejected.
func TestTR_012_RedactVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> INVALID.
	res := runRedact([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--subtree", "deadbeef"}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) Build a doc with one CONTENT segment carrying a distinctive body and a
	// known slot digest; redact it by its content-addressed key.
	var dg [32]byte
	dg[0] = 0xAB
	key := hex.EncodeToString(dg[:4])
	path := writeRedactFixture(t, dg)

	res = runRedact([]string{path, "--subtree", key}, nil)
	if res.Status != "OK" {
		t.Fatalf("redact: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	omissions := res.Extra["declared_omissions"].([]string)
	if len(omissions) != 1 || omissions[0] != key {
		t.Errorf("declared omissions = %v, want [%s]", omissions, key)
	}
	// The redacted segment's body was dropped: output_len is 0 (only segment omitted).
	if res.Extra["output_len"].(int) != 0 {
		t.Errorf("redacted-only output should be empty, got %d octets", res.Extra["output_len"])
	}

	// (3) Not designating anything -> nothing omitted, body retained.
	res = runRedact([]string{path}, nil)
	if len(res.Extra["declared_omissions"].([]string)) != 0 {
		t.Errorf("no subtree designated: nothing should be omitted")
	}
	if res.Extra["output_len"].(int) == 0 {
		t.Errorf("with no redaction the segment body should be retained")
	}
}

// writeRedactFixture writes a valid doc with one CONTENT segment (distinctive
// body) whose slot digest is dg.
func writeRedactFixture(t *testing.T, dg [32]byte) string {
	t.Helper()
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	base := uint64(prefixSize)
	segLen := uint64(128)
	total := base + segLen
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total}
	}
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: base, Length: segLen, FrameCount: 1, Digest: dg},
	}
	img := make([]byte, 0, total)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, _ := (&container.Frontmatter{}).Encode(nil)
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	body := make([]byte, segLen)
	for i := range body {
		body[i] = 0x53 // 'S' — the "secret" plaintext
	}
	img = append(img, body...)
	path := filepath.Join(t.TempDir(), "redact.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}
