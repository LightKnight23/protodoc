package content

import (
	"math/rand"
	"testing"
	"unicode/utf8"
)

// randomRunText builds a pseudo-random UTF-8 string of the given scalar
// length drawn from a mix of ASCII, BMP and astral scalar values, so split
// points fall on multi-byte boundaries too.
func randomRunText(r *rand.Rand, scalars int) string {
	alphabet := []rune{'a', 'b', 'z', '\u00e9', '\u0416', '\u4e2d', '\U0001D11E', '\U0001F600'}
	out := make([]rune, scalars)
	for i := range out {
		out[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(out)
}

// TestFR_020_SplitPreservesRunID is T-0069's named test. Over random run
// lengths and every valid split point, it asserts SplitRun returns two runs
// whose run_id both equal the original's and whose base_ordinals are
// r.base_ordinal and r.base_ordinal+at, with the two pieces' text
// concatenating back to the original (identity preserved, only base_ordinal
// shifts).
func TestFR_020_SplitPreservesRunID(t *testing.T) {
	r := rand.New(rand.NewSource(7))

	for trial := 0; trial < 500; trial++ {
		rid, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		scalars := r.Intn(40) // 0..39
		base := uint32(r.Intn(1000))
		run := Run{RunID: rid, BaseOrdinal: base, Text: randomRunText(r, scalars)}

		// Every valid split point in [0, ScalarLen()].
		for at := 0; at <= run.ScalarLen(); at++ {
			left, right, ok := SplitRun(run, at)
			if !ok {
				t.Fatalf("trial %d: SplitRun at %d (len %d) not ok", trial, at, run.ScalarLen())
			}
			// Both pieces keep the original run_id.
			if !left.RunID.Equal(rid) || !right.RunID.Equal(rid) {
				t.Fatalf("trial %d at %d: split changed run_id (left %x right %x, orig %x)", trial, at, left.RunID, right.RunID, rid)
			}
			// base_ordinals are base and base+at.
			if left.BaseOrdinal != base {
				t.Fatalf("trial %d at %d: left base_ordinal %d, want %d", trial, at, left.BaseOrdinal, base)
			}
			if right.BaseOrdinal != base+uint32(at) {
				t.Fatalf("trial %d at %d: right base_ordinal %d, want %d", trial, at, right.BaseOrdinal, base+uint32(at))
			}
			// Scalar lengths partition the original.
			if left.ScalarLen() != at || right.ScalarLen() != run.ScalarLen()-at {
				t.Fatalf("trial %d at %d: piece lengths %d/%d, want %d/%d", trial, at, left.ScalarLen(), right.ScalarLen(), at, run.ScalarLen()-at)
			}
			// Text concatenates back to the original (no octet lost or added).
			if left.Text+right.Text != run.Text {
				t.Fatalf("trial %d at %d: concatenation does not reconstruct the original text", trial, at)
			}
			// The split landed on a scalar-value boundary (valid UTF-8 halves).
			if !utf8.ValidString(left.Text) || !utf8.ValidString(right.Text) {
				t.Fatalf("trial %d at %d: split produced invalid UTF-8 (split inside a scalar value)", trial, at)
			}
		}
	}
}

// TestFR_020_SplitRejectsOutOfRange confirms an out-of-range split index is
// rejected rather than silently clamped.
func TestFR_020_SplitRejectsOutOfRange(t *testing.T) {
	rid, _ := MintID()
	run := Run{RunID: rid, BaseOrdinal: 5, Text: "hello"}
	if _, _, ok := SplitRun(run, 6); ok {
		t.Fatalf("split past end unexpectedly ok")
	}
	if _, _, ok := SplitRun(run, -1); ok {
		t.Fatalf("split at negative index unexpectedly ok")
	}
}
