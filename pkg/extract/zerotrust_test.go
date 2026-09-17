package extract_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
)

// TestFR_049_ExtractionSucceedsWithZeroTrustMaterial is T-0096's named
// integration test. Extraction of a document whose ATTEST segment is zeroed
// out (corrupted / no trust material) still returns full, correct CONTENT
// text with no error and no verification: the extract package has no import
// edge to integrity (CI-enforced), so no verification path exists to call.
func TestFR_049_ExtractionSucceedsWithZeroTrustMaterial(t *testing.T) {
	// Compile-time / build-graph guarantee: no integrity import edge.
	out, err := exec.Command("go", "list", "-deps", "Protodoc/pkg/extract").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(line) == "Protodoc/pkg/integrity" {
				t.Errorf("extract has a forbidden import edge to integrity (FR-049)")
			}
		}
	}

	// Build a document with two CONTENT segments and one ATTEST segment,
	// then zero out the ATTEST segment's payload (as if trust material were
	// corrupt or absent).
	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLenConst}
	}
	base := uint64(PrefixLenConst)
	segLen := uint64(64)
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: base, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeAttest, Offset: base + segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeContent, Offset: base + 2*segLen, Length: segLen, FrameCount: 1},
	}
	img := make([]byte, 0, PrefixLenConst+3*int(segLen))
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, _ := (&container.Frontmatter{}).Encode(nil)
	img = append(img, fmEnc...)
	stEnc, _ := container.EncodeSegmentTable(slots, nil)
	img = append(img, stEnc...)
	// CONTENT #1 payload 'A', ATTEST payload (will be zeroed), CONTENT #2 'B'.
	img = append(img, bytes.Repeat([]byte{'A'}, int(segLen))...)
	img = append(img, bytes.Repeat([]byte{0x00}, int(segLen))...) // zeroed ATTEST
	img = append(img, bytes.Repeat([]byte{'B'}, int(segLen))...)

	r := bytes.NewReader(img)
	decode := func(seg extract.ContentSegment, octets []byte) (string, error) {
		return string(octets[:1]), nil
	}
	ex, err := extract.NewExtractor(r, decode)
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}

	var texts []string
	for {
		u, ok := ex.Next()
		if !ok {
			break
		}
		texts = append(texts, u.Text)
	}
	if ex.Err() != nil {
		t.Fatalf("extraction errored despite zeroed ATTEST: %v", ex.Err())
	}
	// Both CONTENT units extracted correctly; ATTEST never touched.
	if len(texts) != 2 || texts[0] != "A" || texts[1] != "B" {
		t.Fatalf("extracted %v, want [A B] (full correct CONTENT text, ATTEST ignored)", texts)
	}
}
