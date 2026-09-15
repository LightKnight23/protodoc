package container

import (
	"bytes"
	"testing"
)

func fixtureFrontmatter() *Frontmatter {
	return &Frontmatter{
		Title:      "Quarterly Report",
		PageCount:  42,
		PageWidth:  8 * 914400,
		PageHeight: 11 * 914400,
		Language:   "en-US",
	}
}

// TestFR_054_MetadataFromBoundedPrefix is T-0011's named test.
// Implements: FR-054.
func TestFR_054_MetadataFromBoundedPrefix(t *testing.T) {
	fm := fixtureFrontmatter()

	enc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(enc) != FrontmatterRegionSize {
		t.Fatalf("encoded length = %d, want %d", len(enc), FrontmatterRegionSize)
	}

	// DecodeFrontmatter is handed EXACTLY the bounded region and nothing
	// more: a slice of precisely FrontmatterRegionSize octets. Any attempt
	// by the decoder to read past that boundary indexes out of range and
	// panics this test, proving the four fields are determinable from the
	// leading 262,144 octets alone (FR-054) with no further I/O — no file
	// open, no socket, nothing beyond the bytes already in hand.
	bounded := enc[:FrontmatterRegionSize:FrontmatterRegionSize]
	dec, err := DecodeFrontmatter(bounded)
	if err != nil {
		t.Fatalf("DecodeFrontmatter: %v", err)
	}

	if dec.Title != fm.Title {
		t.Errorf("Title = %q, want %q", dec.Title, fm.Title)
	}
	if dec.PageCount != fm.PageCount {
		t.Errorf("PageCount = %d, want %d", dec.PageCount, fm.PageCount)
	}
	if dec.PageWidth != fm.PageWidth || dec.PageHeight != fm.PageHeight {
		t.Errorf("PageDimensions = (%d,%d), want (%d,%d)", dec.PageWidth, dec.PageHeight, fm.PageWidth, fm.PageHeight)
	}
	if dec.Language != fm.Language {
		t.Errorf("Language = %q, want %q", dec.Language, fm.Language)
	}

	// Re-encoding a decoded value must reproduce the identical octets.
	reenc, err := dec.Encode(nil)
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(reenc, enc) {
		t.Fatal("re-encoded frontmatter octets differ from the original encoding")
	}
}

func TestDecodeFrontmatter_AllFieldsAbsentRoundTrips(t *testing.T) {
	fm := &Frontmatter{}
	enc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(enc) != FrontmatterRegionSize {
		t.Fatalf("encoded length = %d, want %d", len(enc), FrontmatterRegionSize)
	}
	for i, b := range enc {
		if b != 0 {
			t.Fatalf("all-absent frontmatter octet %d = %d, want 0 (empty record, all padding)", i, b)
		}
	}
	dec, err := DecodeFrontmatter(enc)
	if err != nil {
		t.Fatalf("DecodeFrontmatter: %v", err)
	}
	if dec.PreviewKind != 0 || len(dec.PreviewRaster) != 0 || dec.Title != "" ||
		dec.PageCount != 0 || dec.PageWidth != 0 || dec.PageHeight != 0 || dec.Language != "" {
		t.Fatalf("decoded all-absent frontmatter = %+v, want every field at its zero value", *dec)
	}
}

func TestDecodeFrontmatter_RejectsTruncatedRegion(t *testing.T) {
	_, err := DecodeFrontmatter(make([]byte, FrontmatterRegionSize-1))
	if err != ErrFrontmatterTruncated {
		t.Fatalf("got %v, want ErrFrontmatterTruncated", err)
	}
}

// TestFR_051_PreviewPayloadWithinBoundedPrefix is T-0012's named test.
// Implements: FR-051.
func TestFR_051_PreviewPayloadWithinBoundedPrefix(t *testing.T) {
	fm := fixtureFrontmatter()
	fm.PreviewKind = PreviewKindPLP1
	fm.PreviewRaster = bytes.Repeat([]byte{0xAB}, FMPreviewRasterMaxSize) // maximum allowed size

	enc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(enc) != FrontmatterRegionSize {
		t.Fatalf("encoded length = %d, want %d", len(enc), FrontmatterRegionSize)
	}

	dec, err := DecodeFrontmatter(enc)
	if err != nil {
		t.Fatalf("DecodeFrontmatter: %v", err)
	}
	if dec.PreviewKind != fm.PreviewKind {
		t.Errorf("PreviewKind = %d, want %d", dec.PreviewKind, fm.PreviewKind)
	}
	if !bytes.Equal(dec.PreviewRaster, fm.PreviewRaster) {
		t.Errorf("PreviewRaster round-trip mismatch, len got %d want %d", len(dec.PreviewRaster), len(fm.PreviewRaster))
	}
	// The other, already-covered fields must still round-trip unchanged.
	if dec.Title != fm.Title || dec.PageCount != fm.PageCount || dec.Language != fm.Language {
		t.Errorf("unrelated metadata fields changed: got %+v", dec)
	}

	reenc, err := dec.Encode(nil)
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(reenc, enc) {
		t.Fatal("re-encoded frontmatter octets differ from the original encoding")
	}
}

func TestEncodeFrontmatter_RejectsOversizePreviewRaster(t *testing.T) {
	fm := &Frontmatter{PreviewKind: PreviewKindPLP1, PreviewRaster: make([]byte, FMPreviewRasterMaxSize+1)}
	if _, err := fm.Encode(nil); err == nil {
		t.Fatal("Encode: want error for oversize fm-preview-raster, got nil")
	}
}

func TestEncodeFrontmatter_RejectsInvalidPreviewKind(t *testing.T) {
	fm := &Frontmatter{PreviewKind: 3}
	if _, err := fm.Encode(nil); err != ErrFrontmatterInvalidPreviewKind {
		t.Fatalf("Encode: got %v, want ErrFrontmatterInvalidPreviewKind", err)
	}
}

// TestFR_052_PreviewDigestCoversRenderInputs is T-0013's named test.
// Implements: FR-052.
func TestFR_052_PreviewDigestCoversRenderInputs(t *testing.T) {
	base := fixtureFrontmatter()
	base.PreviewKind = PreviewKindPLP1
	base.PreviewRaster = []byte{1, 2, 3}
	baseDigest := ComputeFrontmatterPreviewDigest(base)

	mutations := []struct {
		name string
		mut  func(*Frontmatter)
	}{
		{"title", func(fm *Frontmatter) { fm.Title = fm.Title + "!" }},
		{"page count", func(fm *Frontmatter) { fm.PageCount++ }},
		{"page width", func(fm *Frontmatter) { fm.PageWidth++ }},
		{"page height", func(fm *Frontmatter) { fm.PageHeight++ }},
		{"language", func(fm *Frontmatter) { fm.Language = "fr-FR" }},
	}
	for _, m := range mutations {
		mutated := *base
		m.mut(&mutated)
		got := ComputeFrontmatterPreviewDigest(&mutated)
		if got == baseDigest {
			t.Errorf("mutating %s did not change fm-preview-digest", m.name)
		}
	}

	// Mutating a NOT-in-window input (the preview raster/kind themselves)
	// must NOT change the digest: FR-052's digest is documented as scoped
	// to the in-window document_metadata subset only (plan.md Section 6's
	// disclosed gap), not the preview payload itself.
	outOfScope := *base
	outOfScope.PreviewRaster = []byte{9, 9, 9}
	outOfScope.PreviewKind = PreviewKindRestrictedPNG
	if got := ComputeFrontmatterPreviewDigest(&outOfScope); got != baseDigest {
		t.Errorf("mutating preview kind/raster changed fm-preview-digest; digest must cover only the in-window metadata fields")
	}

	base.PreviewDigest = baseDigest
	enc, err := base.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodeFrontmatter(enc)
	if err != nil {
		t.Fatalf("DecodeFrontmatter: %v", err)
	}
	if dec.PreviewDigest != baseDigest {
		t.Errorf("PreviewDigest round-trip = %x, want %x", dec.PreviewDigest, baseDigest)
	}
}

// TestFR_053_PreviewDigestMismatchReportsStale is T-0014's named test.
// Implements: FR-053.
func TestFR_053_PreviewDigestMismatchReportsStale(t *testing.T) {
	fm := fixtureFrontmatter()
	fm.PreviewKind = PreviewKindPLP1
	fm.PreviewRaster = []byte{1, 2, 3}
	fm.PreviewDigest = ComputeFrontmatterPreviewDigest(fm)

	// Sanity: an unmutated fixture reports OK and returns the raster.
	raster, status := FrontmatterPreview(fm)
	if status != PreviewOK {
		t.Fatalf("status = %v, want PreviewOK", status)
	}
	if string(raster) != string(fm.PreviewRaster) {
		t.Fatalf("raster = %v, want %v", raster, fm.PreviewRaster)
	}

	// Deliberately mismatch fm-preview-digest against its in-window
	// inputs (mutate title after the digest was recorded, as a stale
	// cached preview would look on disk).
	stale := *fm
	stale.Title = stale.Title + " (revised)"

	raster, status = FrontmatterPreview(&stale)
	if status != PreviewStale {
		t.Fatalf("status = %v, want PreviewStale", status)
	}
	if raster != nil {
		t.Fatalf("raster = %v, want nil: FR-053 forbids returning the raw preview bytes on mismatch", raster)
	}

	// The stale record must still round-trip through the wire encoding
	// unchanged (staleness is a consumer-side verdict, not a decode
	// rejection: a bounded-prefix consumer reads no further and renders
	// nothing, but the record itself is well-formed).
	enc, err := stale.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodeFrontmatter(enc)
	if err != nil {
		t.Fatalf("DecodeFrontmatter: %v", err)
	}
	raster, status = FrontmatterPreview(dec)
	if status != PreviewStale || raster != nil {
		t.Fatalf("decoded stale record: status=%v raster=%v, want PreviewStale/nil", status, raster)
	}
}

func TestFrontmatterPreview_AbsentWhenNoPayload(t *testing.T) {
	fm := fixtureFrontmatter() // no PreviewKind/PreviewRaster set
	raster, status := FrontmatterPreview(fm)
	if status != PreviewAbsent {
		t.Fatalf("status = %v, want PreviewAbsent", status)
	}
	if raster != nil {
		t.Fatalf("raster = %v, want nil", raster)
	}
}

func TestDecodeFrontmatter_RejectsNonzeroPadding(t *testing.T) {
	fm := fixtureFrontmatter()
	enc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc[FrontmatterRegionSize-1] = 0xFF // corrupt one padding octet
	_, err = DecodeFrontmatter(enc)
	if err != ErrFrontmatterPaddingNonzero {
		t.Fatalf("got %v, want ErrFrontmatterPaddingNonzero", err)
	}
}
