package registry

import (
	"bytes"
	"testing"

	"Protodoc/pkg/content"
)

// TestFR_012_013_DispositionRoundTripCorpus is T-0062's named conformance test
// (FR-012/FR-013). A corpus of extension envelopes covering all three
// dispositions must round-trip byte-exact (encode -> decode -> re-encode
// yields identical octets) and preserve the declared disposition across the
// round trip -- so a document's dispositions survive being read and rewritten,
// with no reader-chosen normalisation.
func TestFR_012_013_DispositionRoundTripCorpus(t *testing.T) {
	corpus := []struct {
		name string
		env  ExtEnvelope
	}{
		{
			name: "ignore-with-fallback",
			env: ExtEnvelope{
				Tok: NewExtToken(0x00000001, 0x00000001), PayloadLength: 0,
				PayloadRef: fixtureUnitID(0x11), PayloadDigest: fixtureDigest(0x21),
				Disposition: DispositionIgnore, FallbackRef: fixtureUnitID(0x31),
				PositionKey: content.AnchorPoint{RunID: fixtureUnitID(0x41), BirthOrdinal: 1, Side: content.SideBefore, Boundary: content.BoundaryInside},
			},
		},
		{
			name: "degrade-with-fallback",
			env: ExtEnvelope{
				Tok: NewExtToken(0x40000000, 0x00001234), PayloadLength: 65535,
				PayloadRef: fixtureUnitID(0x12), PayloadDigest: fixtureDigest(0x22),
				Disposition: DispositionDegrade, FallbackRef: fixtureUnitID(0x32),
				PositionKey: content.AnchorPoint{RunID: fixtureUnitID(0x42), BirthOrdinal: 4242, Side: content.SideAfter, Boundary: content.BoundaryOutside},
			},
		},
		{
			name: "refuse-no-fallback",
			env: ExtEnvelope{
				Tok: NewExtToken(0x80000000, 0xFFFFFFFE), PayloadLength: 1 << 30,
				PayloadRef: fixtureUnitID(0x13), PayloadDigest: fixtureDigest(0x23),
				Disposition: DispositionRefuse, FallbackRef: Zero16,
				PositionKey: content.AnchorPoint{RunID: fixtureUnitID(0x43), BirthOrdinal: 0, Side: content.SideBefore, Boundary: content.BoundaryInsideIfInsertedAfter},
			},
		},
	}

	seenDisp := map[ExtDisposition]bool{}
	for _, c := range corpus {
		enc, err := c.env.Encode()
		if err != nil {
			t.Fatalf("%s: Encode: %v", c.name, err)
		}
		dec, err := DecodeExtEnvelope(enc)
		if err != nil {
			t.Fatalf("%s: Decode: %v", c.name, err)
		}
		// Disposition survives.
		if dec.Disposition != c.env.Disposition {
			t.Errorf("%s: disposition %v did not survive (got %v)", c.name, c.env.Disposition, dec.Disposition)
		}
		seenDisp[dec.Disposition] = true

		// Byte-exact round trip.
		reEnc, err := dec.Encode()
		if err != nil {
			t.Fatalf("%s: re-Encode: %v", c.name, err)
		}
		if !bytes.Equal(enc, reEnc) {
			t.Errorf("%s: round trip not byte-exact", c.name)
		}
		// Full-struct equality too (every field survives, not just disposition).
		if dec != c.env {
			t.Errorf("%s: decoded envelope differs from original", c.name)
		}
	}

	// The corpus must cover all three dispositions.
	for _, d := range []ExtDisposition{DispositionIgnore, DispositionDegrade, DispositionRefuse} {
		if !seenDisp[d] {
			t.Errorf("corpus does not cover disposition %v", d)
		}
	}
}
