package merge

import (
	"bytes"
	"testing"
)

// FuzzFR_093_BurstNonInterleaving is T-0228's fuzz target (test_kind fuzz,
// named for FR-093's non-interleaving property). It builds several contiguous
// bursts from fuzz-controlled run-ids, state-ids, and texts, merges them via
// the Fugue-order sequence merge, and asserts every burst's text appears as an
// UNBROKEN substring in the merged result -- no burst is ever split by another.
// It also asserts the merge is deterministic (input-order-independent).
func FuzzFR_093_BurstNonInterleaving(f *testing.F) {
	f.Add(byte(1), byte(0x10), "HELLO", byte(2), byte(0x20), "WORLD")
	f.Add(byte(3), byte(0x05), "a", byte(3), byte(0x05), "b") // same run id/state
	f.Add(byte(9), byte(0xFF), "", byte(8), byte(0x00), "xyz")

	f.Fuzz(func(t *testing.T, ra, sa byte, ta string, rb, sb byte, tb string) {
		// Bound payloads so the fuzz stays fast; content is arbitrary.
		if len(ta) > 64 {
			ta = ta[:64]
		}
		if len(tb) > 64 {
			tb = tb[:64]
		}
		runA := mTarget(ra)
		runB := mTarget(rb)
		a := burst(runA, sa, ta)
		b := burst(runB, sb, tb)

		all := append(append([]SeqElement(nil), a...), b...)
		merged := MergeSequence(all)
		text := MergedText(merged)

		// Each NON-EMPTY burst appears as an unbroken substring -- unless the
		// two bursts share the same run_id (then they are one logical run,
		// merged by base_ordinal, and the "burst" is their union).
		if runA != runB {
			if len(ta) > 0 && !bytes.Contains(text, []byte(ta)) {
				t.Fatalf("burst A %q split in merged %q", ta, text)
			}
			if len(tb) > 0 && !bytes.Contains(text, []byte(tb)) {
				t.Fatalf("burst B %q split in merged %q", tb, text)
			}
		}

		// Merge length equals the total element count (no drops, no dups).
		if len(merged) != len(all) {
			t.Fatalf("merge changed element count: %d -> %d", len(all), len(merged))
		}

		// Deterministic: reversing the input yields the same merged text.
		merged2 := MergeSequence(append(append([]SeqElement(nil), b...), a...))
		if !bytes.Equal(text, MergedText(merged2)) {
			t.Fatalf("merge is input-order-dependent: %q vs %q", text, MergedText(merged2))
		}
	})
}
