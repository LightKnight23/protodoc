package ledger

import (
	"math/rand"
	"reflect"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// unitFixture builds a deterministic AddressableUnit from a seed. The first
// octet is set so units spread across several index-route buckets.
func unitFixture(seed int) AddressableUnit {
	var id pdlfmt.UnitID
	for i := range id {
		id[i] = byte(seed*7 + i*13 + 1)
	}
	var dig pdlfmt.Digest256
	for i := range dig {
		dig[i] = byte(seed*3 + i)
	}
	return AddressableUnit{
		Identity: id,
		Offset:   uint64(PrefixLength + seed*512),
		Length:   uint32(64 + seed),
		Kind:     uint16(seed % 5),
		Digest:   dig,
	}
}

// TestFR_055_UnitIndexLeafConstructionRebuildsFromContent is T-0041's named
// test. It builds the UnitIndexLeaf set from a set of addressable units,
// rebuilds it from the identical content presented in a different input
// order, and asserts the two builds are identical (the derived index is a
// pure function of content, carrying no persisted-only or input-order-
// dependent field), that every addressable unit appears exactly once, and
// that each entry is looked up correctly through its routing bucket.
func TestFR_055_UnitIndexLeafConstructionRebuildsFromContent(t *testing.T) {
	const unitCount = 200
	units := make([]AddressableUnit, unitCount)
	for i := range units {
		units[i] = unitFixture(i)
	}

	leaves, err := BuildUnitIndexLeaves(units)
	if err != nil {
		t.Fatalf("BuildUnitIndexLeaves: %v", err)
	}

	// Rebuild from the SAME content in a shuffled order: a derived index
	// must not depend on input order.
	shuffled := append([]AddressableUnit(nil), units...)
	r := rand.New(rand.NewSource(99))
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	rebuilt, err := BuildUnitIndexLeaves(shuffled)
	if err != nil {
		t.Fatalf("BuildUnitIndexLeaves (rebuild): %v", err)
	}

	if !reflect.DeepEqual(leaves, rebuilt) {
		t.Fatalf("rebuilt UnitIndex differs from the original build: the derived index is not a pure function of content")
	}

	// Every unit appears exactly once across all leaves, and each entry
	// sits in the bucket its unit-id routes to.
	total := 0
	for _, leaf := range leaves {
		for _, e := range leaf.Entries {
			total++
			if got := bucketOf(e.Identity); got != leaf.Bucket {
				t.Fatalf("entry %x is in leaf bucket %d but routes to bucket %d", e.Identity, leaf.Bucket, got)
			}
		}
	}
	if total != unitCount {
		t.Fatalf("index holds %d entries, want one per unit = %d", total, unitCount)
	}

	// Entries within each leaf are sorted ascending by unit-id octets.
	for _, leaf := range leaves {
		for i := 1; i < len(leaf.Entries); i++ {
			prev := leaf.Entries[i-1].Identity
			cur := leaf.Entries[i].Identity
			if !(lessID(prev, cur)) {
				t.Fatalf("leaf bucket %d entries not strictly ascending at index %d", leaf.Bucket, i)
			}
		}
	}

	// Every unit is findable via its routing bucket's leaf, and the found
	// entry's offset/length/kind/digest match the source unit.
	leafByBucket := make(map[uint8]UnitIndexLeaf, len(leaves))
	for _, leaf := range leaves {
		leafByBucket[leaf.Bucket] = leaf
	}
	for _, u := range units {
		leaf, ok := leafByBucket[bucketOf(u.Identity)]
		if !ok {
			t.Fatalf("no leaf for bucket %d holding unit %x", bucketOf(u.Identity), u.Identity)
		}
		e, ok := leaf.Lookup(u.Identity)
		if !ok {
			t.Fatalf("unit %x not found in its bucket leaf", u.Identity)
		}
		if e.Offset != u.Offset || e.Length != u.Length || e.Kind != u.Kind || e.Digest != u.Digest {
			t.Fatalf("looked-up entry for %x does not match source unit", u.Identity)
		}
	}
}

// lessID reports whether a's octets sort before b's.
func lessID(a, b pdlfmt.UnitID) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// TestFR_055_RejectsDuplicateIdentityAndEmptyInput confirms the derived
// index rejects a duplicate content-unit identity (the document-wide
// identity-uniqueness rule) and an empty input, rather than building a
// malformed index.
func TestFR_055_RejectsDuplicateIdentityAndEmptyInput(t *testing.T) {
	if _, err := BuildUnitIndexLeaves(nil); err == nil {
		t.Fatalf("empty input was not rejected")
	}
	u := unitFixture(1)
	if _, err := BuildUnitIndexLeaves([]AddressableUnit{u, u}); err == nil {
		t.Fatalf("duplicate identity was not rejected")
	}
}
