package ledger

import (
	"crypto/sha256"
	"testing"

	"Protodoc/pkg/container"
)

// corpusDocument assembles one complete, valid Protodoc file image for the
// no-op-save round-trip test: the 1,048,576-octet fixed prefix (Header,
// CommitRing, Frontmatter, SegmentTable, via the container package's own
// exported canonical encoders) optionally followed by trailing ledger
// segment octets. The images stand in for the M01 conformance corpus at
// this layer; they are assembled from the same encoders M01's own exit
// suite round-trips, so a no-op save that is byte-identical over these is
// byte-identical over any well-formed prefix.
func corpusDocument(t *testing.T, ring *[container.CommitRingSlots]container.CommitRingRecord, fm *container.Frontmatter, slots []container.SegmentTableSlot, ledgerTail []byte) []byte {
	t.Helper()

	h := &container.Header{
		FormatMajor:        1,
		FormatMinor:        0,
		DocumentClass:      1,
		CapabilityWritten:  1,
		CapabilityRequired: 1,
		HistoryMode:        container.HistoryComplete,
		UnicodeVersionID:   1,
		PrefixLayoutID:     1,
	}

	img := make([]byte, 0, PrefixLength+len(ledgerTail))
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(ring, nil)...)

	fmEnc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("corpusDocument: Frontmatter.Encode: %v", err)
	}
	img = append(img, fmEnc...)

	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("corpusDocument: EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)

	if len(img) != PrefixLength {
		t.Fatalf("corpusDocument: assembled prefix is %d octets, want %d", len(img), PrefixLength)
	}
	img = append(img, ledgerTail...)
	return img
}

// noOpCorpus returns a small representative corpus of assembled documents:
// a minimal all-defaults document, a document carrying a populated
// Frontmatter and several segment-table slots, and a document with a
// non-empty ledger tail (already-sealed segment octets). Each carries a
// segmentCount matching its populated slots so a no-op save can report the
// same high-water mark it opened with.
func noOpCorpus(t *testing.T) []struct {
	name         string
	image        []byte
	segmentCount uint64
} {
	t.Helper()

	var minimalRing [container.CommitRingSlots]container.CommitRingRecord
	for i := range minimalRing {
		minimalRing[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLength}
	}
	minimal := corpusDocument(t, &minimalRing, &container.Frontmatter{}, nil, nil)

	var typicalRing [container.CommitRingSlots]container.CommitRingRecord
	for i := range typicalRing {
		typicalRing[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLength, SegmentCount: 2}
	}
	typicalFM := &container.Frontmatter{
		PreviewKind:   container.PreviewKindPLP1,
		PreviewRaster: []byte{1, 2, 3, 4},
		Title:         "no-op save round-trip corpus document",
		PageCount:     3,
	}
	typicalFM.PreviewDigest = container.ComputeFrontmatterPreviewDigest(typicalFM)
	typicalSlots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: PrefixLength, Length: 128, FrameCount: 1},
		{SegmentType: container.SegmentTypeResource, Offset: PrefixLength + 128, Length: 96, FrameCount: 1},
	}
	typical := corpusDocument(t, &typicalRing, typicalFM, typicalSlots, nil)

	// A document with an actual ledger tail: two sealed segments' octets
	// already present past the prefix.
	tail := make([]byte, 0, 224)
	for i := 0; i < 224; i++ {
		tail = append(tail, byte(i*5+1))
	}
	withTail := corpusDocument(t, &typicalRing, typicalFM, typicalSlots, tail)

	return []struct {
		name         string
		image        []byte
		segmentCount uint64
	}{
		{"minimal", minimal, 0},
		{"typical with slots", typical, 2},
		{"with ledger tail", withTail, 2},
	}
}

// TestNFR_003_NoOpSaveByteIdentical is T-0035's named test. It opens each
// corpus document and immediately "saves" it with zero user edits -- a
// no-op save is place() with an empty EditDelta -- then asserts the saved
// image's SHA-256 equals the input image's SHA-256. Any inequality is a
// failure (NFR-003: "emit a file octet-identical to the input for 100% of
// the conformance corpus"). The no-op save must issue zero writes: no
// re-serialisation, no re-ordinal-assignment, no incidental digest
// recomputation touching stored octets.
func TestNFR_003_NoOpSaveByteIdentical(t *testing.T) {
	for _, doc := range noOpCorpus(t) {
		doc := doc
		t.Run(doc.name, func(t *testing.T) {
			inputSum := sha256.Sum256(doc.image)

			res, err := place(doc.image, doc.segmentCount, EditDelta{})
			if err != nil {
				t.Fatalf("no-op save via place: %v", err)
			}

			// The no-op save appended nothing: no placements issued.
			if len(res.Placements) != 0 {
				t.Fatalf("no-op save issued %d placements, want 0 (a zero-edit save must allocate no segment)", len(res.Placements))
			}

			savedSum := sha256.Sum256(res.Image)
			if savedSum != inputSum {
				t.Fatalf("no-op save is not byte-identical: input SHA-256 %x, saved SHA-256 %x", inputSum, savedSum)
			}
			if len(res.Image) != len(doc.image) {
				t.Fatalf("no-op save changed length: input %d octets, saved %d", len(doc.image), len(res.Image))
			}
		})
	}
}

// TestNFR_003_NoOpSaveIsIdempotent confirms repeated no-op saves stay
// byte-identical: opening and zero-edit-saving a document any number of
// times never drifts the octets, which is what makes the round trip safe
// to run in a version-control or sync loop.
func TestNFR_003_NoOpSaveIsIdempotent(t *testing.T) {
	docs := noOpCorpus(t)
	doc := docs[1] // typical with slots
	want := sha256.Sum256(doc.image)

	image := doc.image
	for i := 0; i < 5; i++ {
		res, err := place(image, doc.segmentCount, EditDelta{})
		if err != nil {
			t.Fatalf("iteration %d: place: %v", i, err)
		}
		if got := sha256.Sum256(res.Image); got != want {
			t.Fatalf("iteration %d: no-op save drifted from original SHA-256 %x to %x", i, want, got)
		}
		image = res.Image
	}
}
