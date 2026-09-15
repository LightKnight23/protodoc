package pdlfmt

import (
	"bytes"
	"reflect"
	"testing"
)

func fixtureUnitID(seed byte) UnitID {
	var id UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

func TestUnitIDEqualExactOctetComparison(t *testing.T) {
	a := fixtureUnitID(1)
	b := fixtureUnitID(1)
	c := fixtureUnitID(2)

	if !a.Equal(b) {
		t.Fatal("Equal: identical unit-ids compared unequal")
	}
	if a.Equal(c) {
		t.Fatal("Equal: distinct unit-ids compared equal")
	}

	// A single-octet difference anywhere in the 16 octets must be enough
	// to make Equal false: no partial/prefix match.
	d := a
	d[15] ^= 0xFF
	if a.Equal(d) {
		t.Fatal("Equal: unit-ids differing only in the last octet compared equal")
	}
}

func TestUnitIDBytesRoundTrip(t *testing.T) {
	id := fixtureUnitID(7)
	got := id.Bytes()
	if len(got) != 16 {
		t.Fatalf("Bytes() length = %d, want 16", len(got))
	}
	if !bytes.Equal(got, id[:]) {
		t.Fatalf("Bytes() = %x, want %x", got, id[:])
	}

	// Bytes() must return a copy: mutating it must not alter id.
	got[0] ^= 0xFF
	if id[0] == got[0] {
		t.Fatal("Bytes() did not return an independent copy")
	}
}

func TestUnitIDAppendDecodeRoundTrip(t *testing.T) {
	id := fixtureUnitID(9)
	enc := AppendUnitID(nil, id)
	if len(enc) != 16 {
		t.Fatalf("AppendUnitID length = %d, want 16", len(enc))
	}
	got, n, err := DecodeUnitID(enc)
	if err != nil {
		t.Fatalf("DecodeUnitID: %v", err)
	}
	if n != 16 || !got.Equal(id) {
		t.Fatalf("DecodeUnitID round trip = (%x,%d), want (%x,16)", got, n, id)
	}
}

func TestUnitIDDecodeRejectsTruncated(t *testing.T) {
	if _, _, err := DecodeUnitID(make([]byte, 15)); err == nil {
		t.Fatal("DecodeUnitID: want error for 15-octet input, got nil")
	}
}

// TestCON_008_UnitIdOpaqueExactEqualityOnly is T-0025's named test.
// Implements: CON-008.
//
// It enumerates UnitID's method set by reflection and fails if anything
// beyond {Equal, Bytes} exists: no ordering (Less, Compare), no prefix
// match, and no case-insensitive comparison method may ever be added to
// the type, since unit-id has no hierarchy, case-folding, or traversal
// semantics (CON-008) — identity is exact-comparison equality alone.
func TestCON_008_UnitIdOpaqueExactEqualityOnly(t *testing.T) {
	allowed := map[string]bool{
		"Equal": true,
		"Bytes": true,
	}

	typ := reflect.TypeOf(UnitID{})
	if typ.NumMethod() == 0 {
		t.Fatal("UnitID exposes no methods at all; expected Equal and Bytes")
	}

	seen := make(map[string]bool, typ.NumMethod())
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		seen[name] = true
		if !allowed[name] {
			t.Fatalf("UnitID exposes method %q beyond {Equal, Bytes}: CON-008 forbids any ordering, prefix, or case-insensitive comparison method on unit-id", name)
		}
	}
	for name := range allowed {
		if !seen[name] {
			t.Fatalf("UnitID is missing expected method %q", name)
		}
	}
}
