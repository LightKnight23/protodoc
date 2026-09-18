package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformanceA11Y000AltTextRoundtrip is T-0284's named conformance test
// (vector CONFORMANCE-A11Y-000-alttext-roundtrip; FR-040). The embedded-object
// text-alternative field round-trips through encode/decode for both a
// non-decorative object (non-empty alt text) and a decorative one (empty), and
// the wire grammar / entity table declare the field.
func TestConformanceA11Y000AltTextRoundtrip(t *testing.T) {
	cases := []EmbeddedObjectAlt{
		{Decorative: false, AltText: "A bar chart of quarterly revenue rising each quarter."},
		{Decorative: true, AltText: ""},
	}
	for _, want := range cases {
		enc := EncodeAlt(nil, want)
		got, n, err := DecodeAlt(enc)
		if err != nil {
			t.Fatalf("decode %+v: %v", want, err)
		}
		if got != want || n != len(enc) {
			t.Errorf("roundtrip: got %+v (n=%d), want %+v (len=%d)", got, n, want, len(enc))
		}
	}

	// A reserved decorative flag value (0x02) is rejected.
	if _, _, err := DecodeAlt([]byte{0x02, 0x00}); err == nil {
		t.Errorf("reserved decorative flag 0x02 must be rejected")
	}

	// The wire grammar and entity table declare the fields.
	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	for _, tok := range []string{"ri-alttext", "ri-decorative"} {
		if !strings.Contains(string(abnf), tok) {
			t.Errorf("document.abnf must define %s", tok)
		}
	}
	dm, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "data-model.md"))
	if err != nil {
		t.Fatalf("read data-model.md: %v", err)
	}
	if !strings.Contains(string(dm), "alt_text") {
		t.Errorf("data-model.md must document alt_text")
	}
}
