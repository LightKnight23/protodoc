package content

import (
	"errors"
	"testing"
)

// TestCON_003_WriterRejectsNonNFCRatherThanConverting is T-0074's named
// test. Over a corpus of decomposed / non-NFC strings, WriteRun returns a
// named rejection error (never a normalised/mutated value); over already-NFC
// strings it returns the run byte-identical.
func TestCON_003_WriterRejectsNonNFCRatherThanConverting(t *testing.T) {
	rid, err := testMintID()
	if err != nil {
		t.Fatalf("MintID: %v", err)
	}

	// Non-NFC (decomposed or reorderable) inputs that a writer must refuse.
	nonNFC := []string{
		"A\u0301",            // A + combining acute -> should be U+00C1
		"e\u0301",            // e + combining acute -> should be U+00E9
		"\u0041\u030A\u0301", // A + ring + acute -> should be U+01FA
		"cafe\u0301",         // "café" decomposed
		"\u0301\u0323",       // marks out of canonical order (ccc 230 then 220)
		"\u1100\u1161",       // Hangul L + V -> should compose to a syllable
		"\uAC00\u11A8",       // Hangul LV syllable + T -> should compose to LVT
		"\u1E0B\u0323",       // d-with-dot-above + dot-below, reorderable/composable
	}
	for _, s := range nonNFC {
		in := Run{RunID: rid, Text: s}
		out, err := WriteRun(in)
		if err == nil {
			t.Fatalf("WriteRun accepted non-NFC text %q (should refuse)", s)
		}
		if !errors.Is(err, ErrNonNFCText) {
			t.Fatalf("WriteRun(%q) error %v, want ErrNonNFCText", s, err)
		}
		var nfcErr *NonNFCError
		if !errors.As(err, &nfcErr) {
			t.Fatalf("WriteRun(%q) error is %T, want *NonNFCError", s, err)
		}
		if nfcErr.Text != s {
			t.Fatalf("NonNFCError names text %q, want the offending input %q", nfcErr.Text, s)
		}
		// The returned run must be the zero value, NOT a converted value:
		// no normalisation was performed.
		if out.Text != "" {
			t.Fatalf("WriteRun returned a non-empty run %q for rejected input; it must not convert", out.Text)
		}
	}

	// Already-NFC inputs are accepted byte-identically (no conversion).
	nfc := []string{
		"",
		"hello world",
		"caf\u00e9",      // "café" precomposed
		"\u00C1",         // A-acute precomposed
		"\u01FA",         // A-ring-acute precomposed
		"\uAC00",         // a Hangul syllable (composed)
		"\u0301at start", // a run legally starting with a combining mark (segment boundary)
		"\U0001F600 emoji",
	}
	for _, s := range nfc {
		in := Run{RunID: rid, BaseOrdinal: 3, Text: s}
		out, err := WriteRun(in)
		if err != nil {
			t.Fatalf("WriteRun rejected already-NFC text %q: %v", s, err)
		}
		if out.Text != s || out.RunID != rid || out.BaseOrdinal != 3 {
			t.Fatalf("WriteRun mutated an accepted run: got %+v, want text %q unchanged", out, s)
		}
	}
}
