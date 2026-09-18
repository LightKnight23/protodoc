package merge

import (
	"bytes"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// burst builds a contiguous burst: n elements sharing runID with consecutive
// base_ordinals and one authoring state-id, spelling text.
func burst(runID pdlfmt.UnitID, stateID byte, text string) []SeqElement {
	var sid [32]byte
	sid[0] = stateID
	out := make([]SeqElement, len(text))
	for i := 0; i < len(text); i++ {
		out[i] = SeqElement{RunID: runID, BaseOrdinal: uint32(i), StateID: sid, Value: text[i]}
	}
	return out
}

// TestFR_093_ContiguousBurstSurvivesAsSubstring is T-0227's named unit test
// (FR-092/FR-093). Text inserted by one author in one contiguous burst appears
// as an UNBROKEN substring in the merged result, even when a concurrent burst
// from another author is merged at the same position. Two concurrent bursts do
// not interleave: each appears whole.
func TestFR_093_ContiguousBurstSurvivesAsSubstring(t *testing.T) {
	runA := mTarget(0x01)
	runB := mTarget(0x02)
	// Author A types "HELLO" as one burst (state 0x10); author B types "WORLD"
	// concurrently as one burst (state 0x20).
	a := burst(runA, 0x10, "HELLO")
	b := burst(runB, 0x20, "WORLD")

	all := append(append([]SeqElement(nil), a...), b...)
	merged := MergeSequence(all)
	text := MergedText(merged)

	// Both bursts appear as unbroken substrings.
	if !bytes.Contains(text, []byte("HELLO")) {
		t.Errorf("burst A 'HELLO' was split; merged text = %q", text)
	}
	if !bytes.Contains(text, []byte("WORLD")) {
		t.Errorf("burst B 'WORLD' was split; merged text = %q", text)
	}
	// The merged text is exactly the two bursts concatenated in some whole-run
	// order (no interleaving): either HELLOWORLD or WORLDHELLO.
	if !bytes.Equal(text, []byte("HELLOWORLD")) && !bytes.Equal(text, []byte("WORLDHELLO")) {
		t.Errorf("bursts interleaved; merged text = %q, want HELLOWORLD or WORLDHELLO", text)
	}

	// Deterministic: the R2 state-id tiebreak fixes the order (A's state 0x10
	// < B's 0x20), so A comes first.
	merged2 := MergeSequence(append(append([]SeqElement(nil), b...), a...)) // reversed input
	if !bytes.Equal(MergedText(merged), MergedText(merged2)) {
		t.Errorf("merge order is input-order-dependent: %q vs %q", MergedText(merged), MergedText(merged2))
	}
	if !bytes.Equal(MergedText(merged), []byte("HELLOWORLD")) {
		t.Errorf("R2 tiebreak (0x10 < 0x20) should put HELLO first, got %q", MergedText(merged))
	}

	// The burst's own substring is recoverable intact.
	if !bytes.Equal(BurstSubstring(all, runA), []byte("HELLO")) {
		t.Error("burst A substring not intact")
	}
}
