package extract_test

import (
	"bytes"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
)

// assembleDoc builds a complete fixed-prefix document image from a
// Frontmatter, returning it as a *bytes.Reader (an io.ReaderAt).
func assembleDoc(t *testing.T, fm *container.Frontmatter) *bytes.Reader {
	t.Helper()
	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLenConst}
	}
	img := make([]byte, 0, PrefixLenConst)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Frontmatter.Encode: %v", err)
	}
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(nil, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	if len(img) != PrefixLenConst {
		t.Fatalf("assembled prefix %d, want %d", len(img), PrefixLenConst)
	}
	return bytes.NewReader(img)
}

// TestFR_044_EmitsDescriptiveMetadataWithoutHeavyDecoders is T-0091's named
// test. Metadata returns non-empty title/page-count/language for a populated
// Frontmatter and the documented zero value for absent fields, decoding only
// the Frontmatter region (the extraction import boundary, verified
// elsewhere, guarantees no render/integrity call).
func TestFR_044_EmitsDescriptiveMetadataWithoutHeavyDecoders(t *testing.T) {
	// Populated Frontmatter.
	populated := &container.Frontmatter{
		Title:     "The Extraction View",
		PageCount: 42,
		Language:  "en-GB",
	}
	r := assembleDoc(t, populated)
	md, err := extract.Metadata(r)
	if err != nil {
		t.Fatalf("Metadata (populated): %v", err)
	}
	if md.Title != "The Extraction View" {
		t.Fatalf("title = %q, want populated", md.Title)
	}
	if md.PageCount != 42 {
		t.Fatalf("page count = %d, want 42", md.PageCount)
	}
	if md.Language != "en-GB" {
		t.Fatalf("language = %q, want en-GB", md.Language)
	}

	// Absent fields: documented zero values.
	empty := assembleDoc(t, &container.Frontmatter{})
	md, err = extract.Metadata(empty)
	if err != nil {
		t.Fatalf("Metadata (empty): %v", err)
	}
	if md.Title != "" || md.PageCount != 0 || md.Language != "" {
		t.Fatalf("absent fields did not return zero values: %+v", md)
	}
}
