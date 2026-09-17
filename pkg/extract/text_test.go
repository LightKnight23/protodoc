package extract_test

import (
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_035_ExtractionReproducesAuthoredScalarSequenceExactly is T-0089's
// named conformance test. Extracted text is byte-for-byte identical to the
// golden text for each fixture, including one with pre-composed scalars and
// one with adjacent-but-distinct valid scalar sequences (which a wrongful
// renormalisation would collapse or reorder).
func TestFR_035_ExtractionReproducesAuthoredScalarSequenceExactly(t *testing.T) {
	mkRun := func(seed byte, text string) content.Run {
		var id pdlfmt.UnitID
		id[0] = seed
		return content.Run{RunID: id, Text: text, LangRef: content.LangRef(1)}
	}

	cases := []struct {
		name   string
		runs   []content.Run
		golden string
	}{
		{
			name:   "plain multi-run reading order",
			runs:   []content.Run{mkRun(1, "The "), mkRun(2, "quick "), mkRun(3, "brown fox")},
			golden: "The quick brown fox",
		},
		{
			name:   "pre-composed scalars preserved",
			runs:   []content.Run{mkRun(4, "caf\u00e9"), mkRun(5, " r\u00e9sum\u00e9")},
			golden: "caf\u00e9 r\u00e9sum\u00e9",
		},
		{
			// Adjacent-but-distinct valid scalar sequences: a precomposed
			// e-acute in one run and a base+combining across a boundary in
			// the next. A wrongful cross-unit renormalisation would alter
			// these; extraction must reproduce them exactly as stored.
			name:   "adjacent distinct valid sequences unchanged",
			runs:   []content.Run{mkRun(6, "\u00e9"), mkRun(7, "e\u0301"), mkRun(8, "\U0001D11E")},
			golden: "\u00e9" + "e\u0301" + "\U0001D11E",
		},
	}

	for _, c := range cases {
		got := extract.ExtractText(c.runs)
		if got != c.golden {
			t.Fatalf("%s: extracted %q, want golden %q (byte-for-byte)", c.name, got, c.golden)
		}
		// Byte-for-byte equality (not just string ==) for total certainty.
		if len(got) != len(c.golden) {
			t.Fatalf("%s: extracted length %d, want %d", c.name, len(got), len(c.golden))
		}
		for i := 0; i < len(got); i++ {
			if got[i] != c.golden[i] {
				t.Fatalf("%s: octet %d differs: got 0x%02x, want 0x%02x", c.name, i, got[i], c.golden[i])
			}
		}
	}
}
