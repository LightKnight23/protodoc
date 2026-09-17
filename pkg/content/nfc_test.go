package content

import "testing"

// TestCON_002_NFCScopingIsPerRunNotCrossBoundary is T-0073's named test. It
// confirms IsNFCScoped validates a run's text as NFC in isolation, and that
// a two-run adversarial pair whose CONCATENATION would normalise differently
// than each run alone is not renormalised across the boundary: each run is
// judged NFC on its own octets, and the illegal cross-boundary composition
// is never applied.
func TestCON_002_NFCScopingIsPerRunNotCrossBoundary(t *testing.T) {
	rid, err := testMintID()
	if err != nil {
		t.Fatalf("MintID: %v", err)
	}

	// Adversarial pair: run1 ends in a base letter that is NFC alone; run2
	// begins with a combining acute (U+0301) that is NFC alone. Their
	// concatenation "A" + U+0301 would compose to "A-acute" (U+00C1), so the
	// concatenated stream is NOT NFC -- but neither run individually is
	// non-NFC, and per CON-002 no composition is applied across the boundary.
	run1 := Run{RunID: rid, BaseOrdinal: 0, Text: "A"}
	run2 := Run{RunID: rid, BaseOrdinal: 1, Text: "\u0301world"} // leading combining acute

	if !IsNFCScoped(run1) {
		t.Fatalf("run1 %q should be NFC in isolation", run1.Text)
	}
	if !IsNFCScoped(run2) {
		t.Fatalf("run2 %q should be NFC in isolation (a leading combining mark is legal at a segment start)", run2.Text)
	}

	// The concatenation, by contrast, is NOT NFC -- proving the two runs
	// really do form an adversarial boundary that a cross-boundary
	// normalisation would have collapsed. CON-002 forbids applying that
	// normalisation; the runs are stored and validated independently.
	concat := run1.Text + run2.Text
	if IsNFC(concat) {
		t.Fatalf("the concatenation %q is unexpectedly NFC; the test does not exercise the boundary", concat)
	}

	// A single run that is itself non-NFC (a decomposed base+mark that should
	// have composed WITHIN the run) is rejected: scoping does not excuse a
	// run from being NFC on its own.
	nonNFCRun := Run{RunID: rid, Text: "A\u0301"} // A + combining acute, should be U+00C1
	if IsNFCScoped(nonNFCRun) {
		t.Fatalf("a run %q that is decomposed within itself must be rejected as non-NFC", nonNFCRun.Text)
	}

	// A precomposed run is NFC.
	if !IsNFCScoped(Run{RunID: rid, Text: "\u00C1 world"}) { // "A-acute world"
		t.Fatalf("a precomposed (already-NFC) run was wrongly rejected")
	}

	// Hangul boundary case: an L jamo run and a V jamo run are each NFC
	// alone (a lone jamo is NFC), but their concatenation composes to a
	// syllable and is not NFC -- again not applied across the boundary.
	lRun := Run{RunID: rid, Text: "\u1100"}       // L jamo
	vRun := Run{RunID: rid, Text: "\u1161suffix"} // V jamo + text
	if !IsNFCScoped(lRun) || !IsNFCScoped(vRun) {
		t.Fatalf("lone Hangul jamo runs should be NFC in isolation")
	}
	if IsNFC(lRun.Text + vRun.Text) {
		t.Fatalf("L+V jamo concatenation should not be NFC (composes to a syllable)")
	}

	// The generated tables record the Unicode version they were built from.
	if nfcUnicodeVersion == "" {
		t.Fatalf("nfcUnicodeVersion is empty; NFC tables not generated")
	}
}
