package extract_test

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
)

// trackingReaderAt wraps a byte image and records every [offset, offset+len)
// range read, so the test can assert no read touched a non-CONTENT payload.
type trackingReaderAt struct {
	data  []byte
	reads [][2]int64 // [start, end)
}

func (t *trackingReaderAt) ReadAt(p []byte, off int64) (int, error) {
	n := copy(p, t.data[off:])
	t.reads = append(t.reads, [2]int64{off, off + int64(n)})
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (t *trackingReaderAt) touched(start, end int64) bool {
	for _, r := range t.reads {
		if r[0] < end && start < r[1] {
			return true
		}
	}
	return false
}

// buildMixedDoc assembles a document whose SegmentTable declares one segment
// of each type (CONTENT x2, RESOURCE, HISTORY, ATTEST) with distinct payload
// ranges, and returns the image plus the per-type payload ranges.
func buildMixedDoc(t *testing.T) (*trackingReaderAt, map[byte][2]int64, []uint64) {
	t.Helper()
	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLenConst}
	}

	base := uint64(PrefixLenConst)
	// Ordinals 0..4: CONTENT, RESOURCE, CONTENT, HISTORY, ATTEST.
	segLen := uint64(128)
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: base + 0*segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeResource, Offset: base + 1*segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeContent, Offset: base + 2*segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeHistory, Offset: base + 3*segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeAttest, Offset: base + 4*segLen, Length: segLen, FrameCount: 1},
	}

	img := make([]byte, 0, PrefixLenConst+5*int(segLen))
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
	if len(img) != PrefixLenConst {
		t.Fatalf("assembled prefix %d, want %d", len(img), PrefixLenConst)
	}
	// Append the five segment payloads, each filled with a type marker byte.
	markers := []byte{'C', 'R', 'c', 'H', 'A'}
	for _, m := range markers {
		img = append(img, bytes.Repeat([]byte{m}, int(segLen))...)
	}

	payloadRange := map[byte][2]int64{}
	for i, s := range slots {
		payloadRange[markers[i]] = [2]int64{int64(s.Offset), int64(s.Offset + s.Length)}
	}
	contentOrdinals := []uint64{0, 2}
	return &trackingReaderAt{data: img}, payloadRange, contentOrdinals
}

// PrefixLenConst mirrors extract.PrefixLength for use in composite literals.
const PrefixLenConst = container.SegmentTableOffset + container.SegmentTableRegionSize

// TestFR_041_WalksContentSegmentsOnlyNoHeavyDeps is T-0087's named test. It
// walks a mixed-segment-type document and asserts Walk visits exactly the
// CONTENT segments in ascending storage ordinal, reads a CONTENT payload
// when asked, and never touches a RESOURCE/HISTORY/ATTEST payload's bytes.
func TestFR_041_WalksContentSegmentsOnlyNoHeavyDeps(t *testing.T) {
	r, payloadRange, contentOrdinals := buildMixedDoc(t)

	var visited []uint64
	err := extract.Walk(r, func(seg extract.ContentSegment, sr io.Reader) error {
		visited = append(visited, seg.Ordinal)
		// Read the CONTENT payload; it must be the 'C'/'c' marker bytes.
		buf, err := io.ReadAll(sr)
		if err != nil {
			return err
		}
		if len(buf) == 0 || (buf[0] != 'C' && buf[0] != 'c') {
			t.Fatalf("CONTENT segment ordinal %d payload starts with %q, want a content marker", seg.Ordinal, buf[0])
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	// Visited exactly the CONTENT ordinals, in ascending order.
	if len(visited) != len(contentOrdinals) {
		t.Fatalf("visited %v CONTENT segments, want %v", visited, contentOrdinals)
	}
	for i := range visited {
		if visited[i] != contentOrdinals[i] {
			t.Fatalf("visited[%d] = %d, want %d (ascending storage ordinal)", i, visited[i], contentOrdinals[i])
		}
	}

	// No RESOURCE/HISTORY/ATTEST payload byte was ever read.
	for _, marker := range []byte{'R', 'H', 'A'} {
		rng := payloadRange[marker]
		if r.touched(rng[0], rng[1]) {
			t.Fatalf("Walk read a non-CONTENT (%q) payload range [%d,%d)", marker, rng[0], rng[1])
		}
	}
}

// TestFR_041_ExtractHasNoHeavyImportEdges enforces the module-boundary
// property underlying TR-011: the extract package's full transitive
// import set contains no integrity, render or merge package. It shells out
// to `go list -deps` (the CI check the DoD names).
func TestFR_041_ExtractHasNoHeavyImportEdges(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "Protodoc/pkg/extract").Output()
	if err != nil {
		t.Skipf("go list unavailable in this environment: %v", err)
	}
	forbidden := []string{"Protodoc/pkg/integrity", "Protodoc/pkg/render", "Protodoc/pkg/merge"}
	for _, line := range strings.Split(string(out), "\n") {
		dep := strings.TrimSpace(line)
		for _, bad := range forbidden {
			if dep == bad {
				t.Errorf("extract package has a forbidden import edge to %s (TR-011)", bad)
			}
		}
	}
}
