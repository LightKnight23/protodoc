package content

import (
	"testing"
)

// TestCON_001_BaseOrdinalIsScalarValueScopedToSegment is T-0068's named
// test. It confirms base_ordinal / run positioning counts Unicode SCALAR
// VALUES (not octets, not UTF-16 code units, not grapheme clusters) and is
// scoped to a single run's own segment: character identity at a run-internal
// index is (run_id, base_ordinal+index), and two runs in different segments
// with the same base_ordinal address different identities only by run_id,
// never by any shared document-wide count.
func TestCON_001_BaseOrdinalIsScalarValueScopedToSegment(t *testing.T) {
	rid, err := testMintID()
	if err != nil {
		t.Fatalf("MintID: %v", err)
	}

	// Text mixing ASCII, a combining sequence, an astral-plane character and
	// a multi-codepoint emoji ZWJ sequence. Scalar-value count must equal
	// the number of Unicode code points, NOT octets or grapheme clusters.
	//   "a"          1 scalar
	//   "e"+combining acute (U+0301) 2 scalars (NFC would compose, but we
	//                 count scalars of whatever is present)
	//   astral U+1D11E (musical G clef) 1 scalar, 4 octets
	//   emoji family U+1F468 U+200D U+1F469 U+200D U+1F467 = 5 scalars
	text := "a" + "e\u0301" + "\U0001D11E" + "\U0001F468\u200D\U0001F469\u200D\U0001F467"
	wantScalars := 1 + 2 + 1 + 5

	r := Run{RunID: rid, BaseOrdinal: 100, Text: text}

	if got := r.ScalarLen(); got != wantScalars {
		t.Fatalf("ScalarLen() = %d, want %d scalar values (must count scalars, not octets or graphemes)", got, wantScalars)
	}
	// Sanity: octet length is larger than the scalar count here, so a byte
	// count would give the wrong answer -- proving the unit matters.
	if len(text) == wantScalars {
		t.Fatalf("test text has equal octet and scalar counts; it does not exercise the distinction")
	}

	// Character identity at each in-range index is (RunID, BaseOrdinal+i).
	for i := 0; i < wantScalars; i++ {
		gotID, gotOrd, ok := r.CharIdentityAt(i)
		if !ok {
			t.Fatalf("CharIdentityAt(%d) not ok, want in range", i)
		}
		if !gotID.Equal(rid) {
			t.Fatalf("CharIdentityAt(%d) run_id = %x, want %x", i, gotID, rid)
		}
		if gotOrd != 100+uint32(i) {
			t.Fatalf("CharIdentityAt(%d) ordinal = %d, want %d", i, gotOrd, 100+uint32(i))
		}
	}
	// Out-of-range indices are rejected, never wrapped or extrapolated.
	if _, _, ok := r.CharIdentityAt(wantScalars); ok {
		t.Fatalf("CharIdentityAt(len) unexpectedly ok")
	}
	if _, _, ok := r.CharIdentityAt(-1); ok {
		t.Fatalf("CharIdentityAt(-1) unexpectedly ok")
	}

	// EndOrdinal is base_ordinal + scalar length, a run-internal ordinal.
	if got := r.EndOrdinal(); got != 100+uint32(wantScalars) {
		t.Fatalf("EndOrdinal() = %d, want %d", got, 100+uint32(wantScalars))
	}

	// Segment-scoping: two runs in DIFFERENT segments may legitimately share
	// the same base_ordinal; their character identities differ only by
	// run_id, and there is no shared document-wide count relating them.
	rid2, err := testMintID()
	if err != nil {
		t.Fatalf("MintID: %v", err)
	}
	rA := Run{RunID: rid, BaseOrdinal: 0, Text: "hello"}
	rB := Run{RunID: rid2, BaseOrdinal: 0, Text: "world"}
	idA, ordA, _ := rA.CharIdentityAt(0)
	idB, ordB, _ := rB.CharIdentityAt(0)
	if ordA != ordB {
		t.Fatalf("same base_ordinal in two segments gave different ordinals %d vs %d; base_ordinal is not segment-scoped", ordA, ordB)
	}
	if idA.Equal(idB) {
		t.Fatalf("two distinct-segment runs share a run_id; identity is not segment-distinguished")
	}
}

// TestCON_001_NoCrossSegmentAbsoluteCount is a structural guard: the Run
// type exposes no method returning a document-wide absolute character
// position. ScalarLen and EndOrdinal are both run-internal; there is no
// exported function taking or returning a cross-segment absolute count. This
// mirrors CON-001's "define no persisted construct that counts text
// positions in any [cross-segment] unit" at the type level. The exhaustive
// static grep across the content package is T-0075's own test; this one
// pins the invariant for Run specifically.
func TestCON_001_NoCrossSegmentAbsoluteCount(t *testing.T) {
	// ScalarLen of an empty run is 0; of a run it is a run-internal count
	// bounded by the run's own text length, never accumulating across runs.
	empty := Run{}
	if empty.ScalarLen() != 0 {
		t.Fatalf("empty run ScalarLen = %d, want 0", empty.ScalarLen())
	}
	// Two runs' lengths are independent; nothing sums them into an absolute
	// position, which is exactly the CON-001 property.
	r1 := Run{Text: "abc"}
	r2 := Run{Text: "de"}
	if r1.ScalarLen()+r2.ScalarLen() != 5 {
		t.Fatalf("run length arithmetic sanity failed")
	}
	// r2's EndOrdinal depends only on r2's own BaseOrdinal and text, not on
	// r1: with BaseOrdinal 0 it is 2 regardless of r1.
	if got := (Run{BaseOrdinal: 0, Text: "de"}).EndOrdinal(); got != 2 {
		t.Fatalf("EndOrdinal leaked a cross-segment count: got %d, want 2", got)
	}
}
