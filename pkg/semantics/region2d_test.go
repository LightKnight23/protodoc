package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformance2D000RegionRoundtrip is T-0290's named conformance test
// (vector CONFORMANCE-2D-000-region-roundtrip; FR-099). A 2D-presentation
// region carrying a linearised-reading-order reference and a text alternative
// round-trips through encode/decode, and document.abnf declares the record.
func TestConformance2D000RegionRoundtrip(t *testing.T) {
	want := Region2D{
		RegionID:  pdUnitSem(0x01),
		LinearRef: pdUnitSem(0x40),
		Alt:       EmbeddedObjectAlt{Decorative: false, AltText: "A flowchart: start -> decide -> end."},
	}
	enc := EncodeRegion2D(nil, want)
	got, n, err := DecodeRegion2D(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.LinearRef != want.LinearRef || got.Alt != want.Alt || n != len(enc) {
		t.Errorf("roundtrip: got %+v (n=%d), want linear-ref %x alt %+v", got, n, want.LinearRef[0], want.Alt)
	}

	// Decorative region round-trips with empty alt.
	dec := Region2D{RegionID: pdUnitSem(0x02), LinearRef: pdUnitSem(0x41), Alt: EmbeddedObjectAlt{Decorative: true}}
	e2 := EncodeRegion2D(nil, dec)
	g2, _, err := DecodeRegion2D(e2)
	if err != nil || g2.Alt.Decorative != true || g2.Alt.AltText != "" {
		t.Errorf("decorative region roundtrip: got %+v err=%v", g2, err)
	}

	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	for _, tok := range []string{"region-2d", "r2-linear-ref", "r2-alttext"} {
		if !strings.Contains(string(abnf), tok) {
			t.Errorf("document.abnf must define %s", tok)
		}
	}
}
