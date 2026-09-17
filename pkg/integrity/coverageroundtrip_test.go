package integrity

import (
	"bytes"
	"testing"
)

// TestFR_063_CoverageDescriptorRoundTrip is T-0145's named unit test (FR-063).
// It exercises the single canonical CoverageDescriptor wire implementation
// (T-0139) that M09's signature/coverage call sites reuse: encode then decode
// must reproduce the descriptor, and re-encoding must be byte-identical, for
// TOTAL and SUBSET descriptors including the prefix-region bitmask.
func TestFR_063_CoverageDescriptorRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		d    CoverageDescriptor
	}{
		{
			name: "total mode, all bits, single covered range",
			d: CoverageDescriptor{
				Mode:    CoverageModeTotal,
				Covered: []SegmentRange{{Start: 0, End: 5}},
				Bitmask: CoverageBitHeader | CoverageBitRingWinner | CoverageBitFrontmatter | CoverageBitSegmentTable | CoverageBitIntegrity,
			},
		},
		{
			name: "subset mode, covered + uncovered, no bits",
			d: CoverageDescriptor{
				Mode:      CoverageModeSubset,
				Covered:   []SegmentRange{{Start: 0, End: 2}, {Start: 5, End: 9}},
				Uncovered: []SegmentRange{{Start: 2, End: 5}, {Start: 9, End: 20}},
				Bitmask:   0,
			},
		},
		{
			name: "subset mode, empty covered, one uncovered, header bit only",
			d: CoverageDescriptor{
				Mode:      CoverageModeSubset,
				Uncovered: []SegmentRange{{Start: 0, End: 100}},
				Bitmask:   CoverageBitHeader,
			},
		},
		{
			name: "total mode, empty ranges (degenerate)",
			d:    CoverageDescriptor{Mode: CoverageModeTotal},
		},
	}

	for _, c := range cases {
		enc, err := c.d.Encode(nil)
		if err != nil {
			t.Fatalf("%s: Encode: %v", c.name, err)
		}
		dec, n, err := DecodeCoverageDescriptor(enc)
		if err != nil {
			t.Fatalf("%s: Decode: %v", c.name, err)
		}
		if n != len(enc) {
			t.Errorf("%s: Decode consumed %d of %d octets", c.name, n, len(enc))
		}
		reEnc, err := dec.Encode(nil)
		if err != nil {
			t.Fatalf("%s: re-Encode: %v", c.name, err)
		}
		if !bytes.Equal(enc, reEnc) {
			t.Errorf("%s: round trip not byte-exact:\n first=%x\nsecond=%x", c.name, enc, reEnc)
		}
		if dec.Mode != c.d.Mode || dec.Bitmask != c.d.Bitmask {
			t.Errorf("%s: mode/bitmask mismatch: got mode=%d bits=%08b", c.name, dec.Mode, dec.Bitmask)
		}
		if !rangesEqual(dec.Covered, c.d.Covered) || !rangesEqual(dec.Uncovered, c.d.Uncovered) {
			t.Errorf("%s: range lists did not survive round trip", c.name)
		}
	}
}

func rangesEqual(a, b []SegmentRange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
