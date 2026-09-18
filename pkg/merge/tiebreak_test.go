package merge

import "testing"

// TestFR_092_R2TotalOrderTiebreakDeterministic is T-0229's named unit test
// (FR-092). The R2 TOTAL-ORDER-TIEBREAK resolves two concurrent value-claims by
// their authoring state-ids under unsigned big-endian comparison. It must be
// TOTAL (a strict order for any two distinct ids), DETERMINISTIC (a pure
// function of the 32 octets, independent of argument order beyond the sign),
// and CONSISTENT (antisymmetric: R2Winner(a,b) == -R2Winner(b,a)).
func TestFR_092_R2TotalOrderTiebreakDeterministic(t *testing.T) {
	sid := func(bs ...byte) [32]byte {
		var s [32]byte
		copy(s[:], bs)
		return s
	}

	// A larger state-id wins (last-writer by state id).
	small := sid(0x01)
	large := sid(0x02)
	if R2Winner(large, small) != 1 {
		t.Error("larger state-id should win (+1)")
	}
	if R2Winner(small, large) != -1 {
		t.Error("smaller state-id should lose (-1)")
	}

	// Big-endian: a difference in the FIRST octet dominates a difference in a
	// later octet.
	a := sid(0x02, 0x00)
	b := sid(0x01, 0xFF)
	if R2Winner(a, b) != 1 {
		t.Error("big-endian: leading octet must dominate")
	}

	// Equal ids compare 0 (same authoring state, not a contention).
	if R2Winner(small, small) != 0 {
		t.Error("identical state-ids must compare equal")
	}

	// Antisymmetric and deterministic over many pairs.
	cases := [][2][32]byte{
		{sid(0x00), sid(0xFF)},
		{sid(0x10, 0x20), sid(0x10, 0x21)},
		{sid(0xFF, 0xFF, 0xFF), sid(0xFF, 0xFF, 0xFE)},
	}
	for _, c := range cases {
		w := R2Winner(c[0], c[1])
		rev := R2Winner(c[1], c[0])
		if w != -rev {
			t.Errorf("R2Winner not antisymmetric: %d vs %d", w, rev)
		}
		if w == 0 {
			t.Errorf("distinct ids compared equal")
		}
		// Deterministic: repeated calls agree.
		if R2Winner(c[0], c[1]) != w {
			t.Errorf("R2Winner not deterministic")
		}
	}
}
