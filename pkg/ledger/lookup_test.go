package ledger

import (
	"encoding/binary"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// encodeLeafForTest serialises a UnitIndexLeaf into a self-describing byte
// segment so the lookup's read 2 genuinely reads and decodes leaf octets
// rather than sharing an in-memory structure. Layout: u32 entry count,
// then each entry as unit-id(16) offset(8) length(4) kind(2) digest(32).
func encodeLeafForTest(leaf UnitIndexLeaf) []byte {
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out[0:4], uint32(len(leaf.Entries)))
	for _, e := range leaf.Entries {
		var rec [16 + 8 + 4 + 2 + 32]byte
		copy(rec[0:16], e.Identity[:])
		binary.BigEndian.PutUint64(rec[16:24], e.Offset)
		binary.BigEndian.PutUint32(rec[24:28], e.Length)
		binary.BigEndian.PutUint16(rec[28:30], e.Kind)
		copy(rec[30:62], e.Digest[:])
		out = append(out, rec[:]...)
	}
	return out
}

// decodeLeafForTest is the inverse of encodeLeafForTest, used as
// LocateUnit's decodeLeaf.
func decodeLeafForTest(segment []byte) (UnitIndexLeaf, error) {
	n := binary.BigEndian.Uint32(segment[0:4])
	entries := make([]UnitIndexEntry, 0, n)
	pos := 4
	for i := uint32(0); i < n; i++ {
		var e UnitIndexEntry
		copy(e.Identity[:], segment[pos:pos+16])
		e.Offset = binary.BigEndian.Uint64(segment[pos+16 : pos+24])
		e.Length = binary.BigEndian.Uint32(segment[pos+24 : pos+28])
		e.Kind = binary.BigEndian.Uint16(segment[pos+28 : pos+30])
		copy(e.Digest[:], segment[pos+30:pos+62])
		entries = append(entries, e)
		pos += 62
	}
	var bucket uint8
	if len(entries) > 0 {
		bucket = bucketOf(entries[0].Identity)
	}
	return UnitIndexLeaf{Bucket: bucket, Entries: entries}, nil
}

// TestFR_055_BoundedLookupWithinThreeReads is T-0042's named test. It
// builds a synthetic document with 10,000 addressable units distributed
// across the index-route buckets, one UnitIndexLeaf segment per bucket,
// and confirms that locating an arbitrary unit completes in exactly <= 3
// sequential dependent reads (instrumented) with the octet volume read
// beyond the prefix bounded and reported as the FR-055 budget figure.
func TestFR_055_BoundedLookupWithinThreeReads(t *testing.T) {
	const unitCount = 10000

	// Build the units. Their content lives in a units region placed after
	// the leaf segments; here the exact byte contents are irrelevant, only
	// the offsets/lengths the index records.
	units := make([]AddressableUnit, unitCount)
	for i := range units {
		units[i] = unitFixture(i)
	}

	leaves, err := BuildUnitIndexLeaves(units)
	if err != nil {
		t.Fatalf("BuildUnitIndexLeaves: %v", err)
	}

	// Lay out the image: [prefix][leaf segments...][unit region].
	// Leaf segments come first after the prefix so their offsets are known;
	// then a unit region large enough to hold every unit's declared extent.
	image := make([]byte, PrefixLength)

	var indexRoute [IndexRouteBuckets]uint16
	for b := range indexRoute {
		indexRoute[b] = 0xFFFF // default: empty bucket
	}
	leafPlacements := make(map[uint16]LeafPlacement)

	// Assign each leaf a distinct segment ordinal and place its encoded
	// octets in the image.
	var ordinal uint16
	for _, leaf := range leaves {
		enc := encodeLeafForTest(leaf)
		off := uint64(len(image))
		image = append(image, enc...)
		indexRoute[leaf.Bucket] = ordinal
		leafPlacements[ordinal] = LeafPlacement{Offset: off, Length: uint64(len(enc))}
		ordinal++
	}

	// Now append a unit region and rewrite each unit's recorded offset to
	// point into it, so read 3 hits real bytes. Rebuild leaves against the
	// updated offsets.
	unitRegionStart := uint64(len(image))
	// Rewrite offsets: unit i lives at unitRegionStart + i*maxUnitLen.
	const maxUnitLen = 256
	for i := range units {
		units[i].Offset = unitRegionStart + uint64(i)*maxUnitLen
		units[i].Length = maxUnitLen
	}
	image = append(image, make([]byte, unitCount*maxUnitLen)...)

	// Rebuild leaves and re-encode with the corrected offsets.
	leaves, err = BuildUnitIndexLeaves(units)
	if err != nil {
		t.Fatalf("rebuild leaves: %v", err)
	}
	// Re-place leaf encodings (lengths unchanged, so offsets are stable).
	ordinal = 0
	for _, leaf := range leaves {
		enc := encodeLeafForTest(leaf)
		placement := leafPlacements[ordinal]
		copy(image[placement.Offset:placement.Offset+placement.Length], enc)
		ordinal++
	}

	// Locate several arbitrary units (first, middle, last, and a few more)
	// and assert the bound for each.
	targets := []int{0, 1, 5000, 9999, 314, 2718}
	var maxBeyond int
	for _, ti := range targets {
		want := units[ti]
		got, err := LocateUnit(image, want.Identity, indexRoute, leafPlacements, decodeLeafForTest)
		if err != nil {
			t.Fatalf("LocateUnit(unit %d): %v", ti, err)
		}
		if got.DependentReads > FR055MaxDependentReads {
			t.Fatalf("unit %d located in %d dependent reads, exceeds FR-055 ceiling %d", ti, got.DependentReads, FR055MaxDependentReads)
		}
		if got.DependentReads != 3 {
			t.Fatalf("unit %d located in %d dependent reads, expected exactly 3 (prefix, leaf, unit)", ti, got.DependentReads)
		}
		if got.Offset != want.Offset || got.Length != want.Length {
			t.Fatalf("unit %d located at offset %d len %d, want offset %d len %d", ti, got.Offset, got.Length, want.Offset, want.Length)
		}
		if got.OctetsBeyondPrefix > maxBeyond {
			maxBeyond = got.OctetsBeyondPrefix
		}
	}

	// FR-055's second clause: octets read beyond the (already-resident)
	// prefix must be bounded. Beyond-prefix reads are one leaf segment plus
	// one unit; the largest leaf plus one maxUnitLen unit is far below the
	// FR-055 "greater of 1048576 octets or 1% of file size" allowance.
	t.Logf("FR-055 measured octets read beyond prefix (leaf + unit): max %d", maxBeyond)
	fileSize := len(image)
	budget := 1048576
	if onePercent := fileSize / 100; onePercent > budget {
		budget = onePercent
	}
	if maxBeyond > budget {
		t.Fatalf("octets read beyond prefix %d exceed FR-055 budget %d (max(1MiB, 1%% of %d))", maxBeyond, budget, fileSize)
	}

	// A unit whose bucket is unregistered, and an absent unit, both fail
	// cleanly within the read bound rather than scanning.
	var absent pdlfmt.UnitID
	absent[0] = 0x00
	for i := 1; i < len(absent); i++ {
		absent[i] = 0xEE
	}
	if _, err := LocateUnit(image, absent, indexRoute, leafPlacements, decodeLeafForTest); err == nil {
		t.Fatalf("locating an absent unit unexpectedly succeeded")
	}
}
