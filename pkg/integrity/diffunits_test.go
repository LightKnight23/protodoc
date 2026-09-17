package integrity

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_068_EnumeratesUnitsDifferingFromSignedState is T-0162's named unit
// test (FR-068). A reader can enumerate EVERY content unit that differs
// between a signed state and the current state: added, removed, and changed
// units are each named, unchanged units are not, and the enumeration is
// exhaustive and deterministic.
func TestFR_068_EnumeratesUnitsDifferingFromSignedState(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	dg := func(b byte) Digest { return Digest{b} }

	signed := StateUnits{
		u(0x01): dg(0xA1), // unchanged
		u(0x02): dg(0xA2), // will change
		u(0x03): dg(0xA3), // will be removed
	}
	current := StateUnits{
		u(0x01): dg(0xA1), // unchanged
		u(0x02): dg(0xB2), // changed
		u(0x04): dg(0xA4), // added
	}

	diffs := EnumerateUnitDifferences(signed, current)

	// Expect exactly: 0x02 changed, 0x03 removed, 0x04 added. 0x01 unchanged
	// is NOT reported.
	got := map[pdlfmt.UnitID]UnitDiffKind{}
	for _, d := range diffs {
		got[d.Unit] = d.Kind
	}
	if len(diffs) != 3 {
		t.Fatalf("got %d differences, want 3: %+v", len(diffs), diffs)
	}
	if got[u(0x02)] != UnitChanged {
		t.Errorf("unit 0x02: kind %v, want changed", got[u(0x02)])
	}
	if got[u(0x03)] != UnitRemoved {
		t.Errorf("unit 0x03: kind %v, want removed", got[u(0x03)])
	}
	if got[u(0x04)] != UnitAdded {
		t.Errorf("unit 0x04: kind %v, want added", got[u(0x04)])
	}
	if _, reported := got[u(0x01)]; reported {
		t.Errorf("unchanged unit 0x01 was reported as differing")
	}

	// Deterministic order: sorted by unit-id octets.
	for i := 1; i < len(diffs); i++ {
		if compareUnitID(diffs[i-1].Unit, diffs[i].Unit) > 0 {
			t.Errorf("differences not in deterministic unit-id order at %d", i)
		}
	}

	// Identical states yield no differences.
	if d := EnumerateUnitDifferences(signed, signed); len(d) != 0 {
		t.Errorf("identical states reported %d differences, want 0", len(d))
	}
}
