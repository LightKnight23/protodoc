package cli

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

// writeMergeGuardDoc writes a valid document with the given CONTENT frames (in
// table order, at ordinals 0..len(contentFrames)-1) followed by the given
// HISTORY segment bodies (bare ERASURE_RECORD bytes, no ledger header — the
// same convention historySegmentBody accepts for "prefix-only" fixtures), and
// a commit-ring winner declaring mode and retentionPoint (T-0393).
func writeMergeGuardDoc(t *testing.T, mode container.HistoryMode, retentionPoint uint16, contentFrames [][]byte, historyBodies [][]byte) string {
	t.Helper()
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: mode, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	base := uint64(prefixSize)
	off := base
	var slots []container.SegmentTableSlot
	for _, fr := range contentFrames {
		slots = append(slots, container.SegmentTableSlot{
			SegmentType: container.SegmentTypeContent,
			Offset:      off, Length: uint64(len(fr)), FrameCount: 1,
		})
		off += uint64(len(fr))
	}
	for _, hb := range historyBodies {
		slots = append(slots, container.SegmentTableSlot{
			SegmentType: container.SegmentTypeHistory,
			Offset:      off, Length: uint64(len(hb)), FrameCount: 1,
		})
		off += uint64(len(hb))
	}
	total := off
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total, RetentionPoint: retentionPoint}
	}
	img := make([]byte, 0, total)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, err := (&container.Frontmatter{}).Encode(nil)
	if err != nil {
		t.Fatalf("Frontmatter.Encode: %v", err)
	}
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	if len(img) != prefixSize {
		t.Fatalf("prefix %d, want %d", len(img), prefixSize)
	}
	for _, fr := range contentFrames {
		img = append(img, fr...)
	}
	for _, hb := range historyBodies {
		img = append(img, hb...)
	}
	path := filepath.Join(t.TempDir(), "guardfixture.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

// TestCON_024_MergeRefusesAcrossRetentionPoint is T-0393's named integration
// test for CON-024. The base declares history-retained-from-point with
// retention point 1; an incoming branch carries a change at CONTENT ordinal 0
// (predating the retention point) -> the merge is REFUSED naming CON-024. A
// branch whose only change is at or after the retention point merges normally.
func TestCON_024_MergeRefusesAcrossRetentionPoint(t *testing.T) {
	idX := cliUnit(0x20)
	idY := cliUnit(0x21)
	outPath := func() string { return filepath.Join(t.TempDir(), "merged.pdl") }

	// Base: retained-from-point at 1, two constructs at ordinals 0 and 1.
	base := writeMergeGuardDoc(t, container.HistoryRetainedFromPoint, 1,
		[][]byte{
			realFrameWithPayload(0x01, idX, []byte("base-x")),
			realFrameWithPayload(0x01, idY, []byte("base-y")),
		}, nil)

	// Branch A changes ordinal 0 (idX) -- predates retention point 1 -> refused.
	sideA := writeMergeGuardDoc(t, container.HistoryRetainedFromPoint, 1,
		[][]byte{
			realFrameWithPayload(0x01, idX, []byte("changed-x")),
			realFrameWithPayload(0x01, idY, []byte("base-y")),
		}, nil)

	res := runMerge([]string{base, sideA, base, "--out", outPath()}, nil)
	if res.Status != "REFUSED" || res.Extra["condition"] != "CON-024" {
		t.Errorf("change predating retention point: status=%s condition=%v, want REFUSED/CON-024", res.Status, res.Extra["condition"])
	}

	// Branch B changes only ordinal 1 (idY) -- at/after the retention point,
	// no erasure conflict -> merges cleanly.
	sideB := writeMergeGuardDoc(t, container.HistoryRetainedFromPoint, 1,
		[][]byte{
			realFrameWithPayload(0x01, idX, []byte("base-x")),
			realFrameWithPayload(0x01, idY, []byte("changed-y")),
		}, nil)
	res = runMerge([]string{base, sideB, base, "--out", outPath()}, nil)
	if res.Status != "OK" {
		t.Errorf("change at/after retention point: status=%s findings=%+v, want OK", res.Status, res.Findings)
	}
}

// TestFR_096_MergeRefusesReplayOfErasedUnit is T-0393's named integration test
// for FR-096. The base carries a HISTORY segment recording idZ as erased with
// a specific content digest; an incoming branch introduces idZ with exactly
// that content (a replay of the erased content) -> the merge is REFUSED naming
// FR-096. A branch introducing idZ with DIFFERENT content is a new authoring
// act and merges normally.
func TestFR_096_MergeRefusesReplayOfErasedUnit(t *testing.T) {
	idZ := cliUnit(0x22)
	erasedContent := []byte("erased-secret")
	erasedFrame := realFrameWithPayload(0x01, idZ, erasedContent)
	digest := sha256.Sum256(erasedFrame)

	var salt [history.SaltSize]byte
	for i := range salt {
		salt[i] = 0xAA
	}
	rec := history.NewErasureRecord(idZ, pdlfmt.Digest256(digest), salt).Trimmed()
	recBytes, err := rec.Encode()
	if err != nil {
		t.Fatalf("ErasureRecord.Encode: %v", err)
	}

	outPath := func() string { return filepath.Join(t.TempDir(), "merged.pdl") }

	// Base: complete-history, no content at idZ, but records idZ as erased.
	base := writeMergeGuardDoc(t, container.HistoryComplete, 0, nil, [][]byte{recBytes})

	// Branch A replays the exact erased content at idZ -> refused.
	replay := writeMergeGuardDoc(t, container.HistoryComplete, 0,
		[][]byte{erasedFrame}, [][]byte{recBytes})
	res := runMerge([]string{base, replay, base, "--out", outPath()}, nil)
	if res.Status != "REFUSED" || res.Extra["condition"] != "FR-096" {
		t.Errorf("replay of erased content: status=%s condition=%v, want REFUSED/FR-096", res.Status, res.Extra["condition"])
	}

	// Both branches introduce idZ with the SAME different-from-erased content
	// (an agreed new authoring act, avoiding an unrelated three-way-merge
	// ambiguity when a construct is added on only one side and absent from
	// base) -> the FR-096 digest mismatch permits it, and the agreed value
	// merges cleanly.
	newAuthoring := writeMergeGuardDoc(t, container.HistoryComplete, 0,
		[][]byte{realFrameWithPayload(0x01, idZ, []byte("new-content"))}, [][]byte{recBytes})
	res = runMerge([]string{base, newAuthoring, newAuthoring, "--out", outPath()}, nil)
	if res.Status != "OK" {
		t.Errorf("new content at erased identity: status=%s findings=%+v, want OK", res.Status, res.Findings)
	}
}
