package cli

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
)

// fixtureSeg is one live segment for the shared assembler: its type and its
// complete octet body.
type fixtureSeg struct {
	segType byte
	body    []byte
}

// assembleDoc builds a complete, storage-integrity-CORRECT document: it lays
// the segment bodies out contiguously after the fixed prefix, sets each slot's
// Digest to the real SHA-256 of its body (the slot-digest convention), computes
// the winning commit-ring record's LedgerRoot as integrity.TSRoot over those
// slots, and reissues every ring slot with the correct SegmentCount and
// LedgerLength. `ringMutator`, if non-nil, may set extra ring fields (e.g.
// TCRoot, RetentionPoint, HistoryMode-independent state) BEFORE the LedgerRoot
// is computed — so the sealed LedgerRoot always matches the final slots. The
// returned bytes are a document `protodoc validate` accepts.
func assembleDoc(t *testing.T, h *container.Header, segs []fixtureSeg, ringMutator func(*container.CommitRingRecord)) []byte {
	t.Helper()

	base := uint64(prefixSize)
	off := base
	slots := make([]container.SegmentTableSlot, len(segs))
	var bodyTotal uint64
	for i, s := range segs {
		dg := sha256.Sum256(s.body)
		slots[i] = container.SegmentTableSlot{
			SegmentType: s.segType,
			Offset:      off,
			Length:      uint64(len(s.body)),
			FrameCount:  1,
			Digest:      dg,
		}
		off += uint64(len(s.body))
		bodyTotal += uint64(len(s.body))
	}
	total := base + bodyTotal

	// The validator recomputes T_S over the FULL decoded segment table (all
	// MaxSegments slots; unused ones carry a zero digest, distinct from an
	// absent leaf). Compute the authentic ledger_root over that same full-width
	// slot set so the sealed ledger_root matches what validate recomputes.
	fullSlots := make([]container.SegmentTableSlot, container.MaxSegments)
	copy(fullSlots, slots)
	ledgerRoot, err := integrity.TSRoot(fullSlots)
	if err != nil {
		t.Fatalf("assembleDoc: TSRoot: %v", err)
	}

	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		rec := container.CommitRingRecord{
			Sequence:     uint64(i + 1),
			LedgerLength: total,
			SegmentCount: uint16(len(slots)),
		}
		if ringMutator != nil {
			ringMutator(&rec)
		}
		copy(rec.LedgerRoot[:], ledgerRoot[:])
		ring[i] = rec
	}

	img := make([]byte, 0, total)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, err := (&container.Frontmatter{}).Encode(nil)
	if err != nil {
		t.Fatalf("assembleDoc: Frontmatter.Encode: %v", err)
	}
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("assembleDoc: EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	if len(img) != prefixSize {
		t.Fatalf("assembleDoc: prefix %d, want %d", len(img), prefixSize)
	}
	for _, s := range segs {
		img = append(img, s.body...)
	}
	return img
}

// writeDoc writes assembleDoc's output to a temp file and returns the path.
func writeDoc(t *testing.T, h *container.Header, segs []fixtureSeg, ringMutator func(*container.CommitRingRecord)) string {
	t.Helper()
	img := assembleDoc(t, h, segs, ringMutator)
	path := filepath.Join(t.TempDir(), "doc.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("writeDoc: %v", err)
	}
	return path
}

// defaultHeader is the standard valid header the fixtures use.
func defaultHeader() *container.Header {
	return &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
}
