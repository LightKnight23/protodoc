package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformanceBIDI000DirectionRoundtrip is T-0269's named conformance test
// (vector id CONFORMANCE-BIDI-000-direction-roundtrip; FR-032). The direction
// field is mirrored into document.abnf's text-block and table productions, and
// a round-trip encode/decode of each direction value is byte-exact.
func TestConformanceBIDI000DirectionRoundtrip(t *testing.T) {
	// The ABNF carries tb-direction and tbl-direction with the closed value set.
	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	text := string(abnf)
	for _, tok := range []string{"tb-direction", "tbl-direction"} {
		if !strings.Contains(text, tok) {
			t.Errorf("document.abnf must define %s", tok)
		}
	}

	// Round-trip both closed-set values byte-exact.
	for _, d := range []Direction{DirectionLTR, DirectionRTL} {
		enc, err := EncodeDirectionField(nil, d)
		if err != nil {
			t.Fatalf("encode %d: %v", d, err)
		}
		got, n, err := DecodeDirectionField(enc)
		if err != nil {
			t.Fatalf("decode %d: %v", d, err)
		}
		if got != d {
			t.Errorf("round-trip: got %d, want %d", got, d)
		}
		if n != len(enc) {
			t.Errorf("decode consumed %d octets, want %d", n, len(enc))
		}
		// Re-encoding the decoded value is byte-identical.
		reenc, _ := EncodeDirectionField(nil, got)
		if string(reenc) != string(enc) {
			t.Errorf("re-encode not byte-exact for %d", d)
		}
	}

	// An out-of-set value is refused on encode and decode.
	if _, err := EncodeDirectionField(nil, Direction(2)); err != ErrDirectionOutOfSet {
		t.Errorf("encode out-of-set: err = %v, want ErrDirectionOutOfSet", err)
	}
	badWire := []byte{DirectionFieldTag, 1, 2} // tag, len=1, value=2
	if _, _, err := DecodeDirectionField(badWire); err != ErrDirectionOutOfSet {
		t.Errorf("decode out-of-set: err = %v, want ErrDirectionOutOfSet", err)
	}
}
