package integrity

import "testing"

// TestFR_063_CoverageDescriptorDigestDeterministic is T-0151's named unit test
// (FR-063). coverage-descriptor-digest is SHA-256 over the descriptor's
// canonical encoded octets (integrity.abnf S3.2). It must be deterministic --
// equal descriptors yield equal digests -- and sensitive: any change to a
// range entry or the bitmask changes the digest. This is the value that feeds
// the signed_object preimage.
func TestFR_063_CoverageDescriptorDigestDeterministic(t *testing.T) {
	d := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 4}},
		Uncovered: []SegmentRange{{Start: 4, End: 10}},
		Bitmask:   CoverageBitHeader | CoverageBitIntegrity,
	}

	// Deterministic: repeated digests of the same descriptor are identical.
	d1, err := d.Digest()
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	d2, err := d.Digest()
	if err != nil {
		t.Fatalf("Digest (again): %v", err)
	}
	if d1 != d2 {
		t.Fatal("coverage-descriptor-digest is not deterministic for the same descriptor")
	}

	// An independently-constructed equal descriptor yields the same digest.
	same := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 4}},
		Uncovered: []SegmentRange{{Start: 4, End: 10}},
		Bitmask:   CoverageBitHeader | CoverageBitIntegrity,
	}
	sd, _ := same.Digest()
	if sd != d1 {
		t.Fatal("equal descriptors produced different digests")
	}

	// Sensitivity: each of a range change, a bitmask change, and a mode
	// change alters the digest.
	variants := []CoverageDescriptor{
		{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 5}}, Uncovered: []SegmentRange{{Start: 5, End: 10}}, Bitmask: CoverageBitHeader | CoverageBitIntegrity}, // range change
		{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 4}}, Uncovered: []SegmentRange{{Start: 4, End: 10}}, Bitmask: CoverageBitHeader},                        // bitmask change
		{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 4}}, Bitmask: CoverageBitHeader | CoverageBitIntegrity},                                                  // mode change
	}
	for i, v := range variants {
		vd, err := v.Digest()
		if err != nil {
			t.Fatalf("variant %d Digest: %v", i, err)
		}
		if vd == d1 {
			t.Errorf("variant %d has the same digest as the base descriptor (digest not sensitive)", i)
		}
	}
}
