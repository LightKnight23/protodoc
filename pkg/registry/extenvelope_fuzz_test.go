package registry

import (
	"testing"

	"Protodoc/pkg/content"
)

// FuzzExtEnvelopeDecode is T-0066's native fuzz target (feeding CP-012's
// continuous fuzz). It drives arbitrary bytes through DecodeExtEnvelope,
// asserting the decoder never panics on any input, and that any envelope it
// accepts round-trips byte-exact (Encode of the decoded value reproduces a
// record that decodes back to the same value) -- so a successful decode always
// yields a canonical, stable encoding. Malformed inputs must be rejected with
// an error, never a panic and never a partial/garbage envelope.
func FuzzExtEnvelopeDecode(f *testing.F) {
	// Seed corpus (>=2 seeds so fuzz-maturity treats this as a real target).
	// (1) a well-formed envelope, (2) empty, (3) a truncated-looking record.
	valid, err := ExtEnvelope{
		Tok:           NewExtToken(0x00000001, 0x00000002),
		PayloadLength: 42,
		PayloadRef:    fixtureUnitID(0x10),
		PayloadDigest: fixtureDigest(0x20),
		Disposition:   DispositionDegrade,
		FallbackRef:   fixtureUnitID(0x30),
		PositionKey:   content.AnchorPoint{RunID: fixtureUnitID(0x40), BirthOrdinal: 1, Side: content.SideBefore, Boundary: content.BoundaryInside},
	}.Encode()
	if err != nil {
		f.Fatalf("seed encode: %v", err)
	}
	f.Add(valid)
	f.Add([]byte{})
	f.Add([]byte{extEnvelopeDiscriminant, 0x01})

	f.Fuzz(func(t *testing.T, data []byte) {
		e, err := DecodeExtEnvelope(data) // must never panic
		if err != nil {
			return // malformed input rejected: fine
		}
		// Accepted: re-encoding must succeed and decode back to an equal value.
		reEnc, err := e.Encode()
		if err != nil {
			t.Fatalf("accepted envelope failed to re-encode: %v", err)
		}
		e2, err := DecodeExtEnvelope(reEnc)
		if err != nil {
			t.Fatalf("re-encoded envelope failed to decode: %v", err)
		}
		if e2 != e {
			t.Fatalf("decode(encode(decode(x))) != decode(x): not a stable canonical form")
		}
	})
}
