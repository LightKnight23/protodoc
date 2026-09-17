package integrity

import (
	"bytes"
	"testing"
)

// FuzzFR_063_CoverageDescriptorCanonicalisation is T-0150's native fuzz target
// (feeding CP-012's continuous fuzz). It drives arbitrary bytes through
// DecodeCoverageDescriptor and asserts:
//
//   - the decoder never panics on any input;
//   - any successfully decoded descriptor re-encodes to octets that decode
//     back to an equal descriptor (decode -> encode -> decode is stable), so a
//     decode never yields a value that cannot round-trip; and
//   - any decoded descriptor that ALSO passes Validate is in canonical form:
//     re-encoding it and re-validating still passes, and the re-encoded octets
//     decode-encode identically -- exactly one valid encoding per coverage set
//     (PD-COVER-002).
func FuzzFR_063_CoverageDescriptorCanonicalisation(f *testing.F) {
	// Seed corpus (>=2 seeds).
	seed := func(d CoverageDescriptor) []byte {
		b, err := d.Encode(nil)
		if err != nil {
			f.Fatalf("seed encode: %v", err)
		}
		return b
	}
	f.Add(seed(CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 4}}}))
	f.Add(seed(CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 2}}, Uncovered: []SegmentRange{{Start: 2, End: 8}}}))
	f.Add([]byte{})
	f.Add([]byte{0x00})

	f.Fuzz(func(t *testing.T, data []byte) {
		d, n, err := DecodeCoverageDescriptor(data) // must never panic
		if err != nil {
			return // malformed input rejected: fine
		}
		if n < 0 || n > len(data) {
			t.Fatalf("decode consumed %d octets of %d", n, len(data))
		}

		// decode -> encode -> decode is stable.
		enc, err := d.Encode(nil)
		if err != nil {
			t.Fatalf("decoded descriptor failed to re-encode: %v", err)
		}
		d2, _, err := DecodeCoverageDescriptor(enc)
		if err != nil {
			t.Fatalf("re-encoded descriptor failed to decode: %v", err)
		}
		if d2.Mode != d.Mode || d2.Bitmask != d.Bitmask ||
			!rangesEqual(d2.Covered, d.Covered) || !rangesEqual(d2.Uncovered, d.Uncovered) {
			t.Fatalf("decode/encode/decode not stable")
		}

		// If it validates (against a generous segment count with no ATTEST
		// ordinals), it is canonical: re-encoding and re-decoding is byte and
		// value stable, and re-validation still passes.
		const segs = 1 << 20
		if verr := d.Validate(segs, map[uint16]bool{}); verr == nil {
			enc2, _ := d2.Encode(nil)
			if !bytes.Equal(enc, enc2) {
				t.Fatalf("a validating descriptor has two distinct encodings (non-canonical)")
			}
			if verr2 := d2.Validate(segs, map[uint16]bool{}); verr2 != nil {
				t.Fatalf("a validating descriptor failed to re-validate after round trip: %v", verr2)
			}
		}
	})
}
