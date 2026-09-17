package registry

import (
	"bytes"
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/pdlfmt"
)

func fixtureUnitID(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

func fixtureDigest(seed byte) pdlfmt.Digest256 {
	var d pdlfmt.Digest256
	for i := range d {
		d[i] = seed ^ byte(i)
	}
	return d
}

// TestFR_012_ExtEnvelopeRoundTrip is T-0052's named unit test for the core
// ExtensionEnvelope wire struct (FR-012): Encode then Decode must reproduce
// the same envelope, and re-encoding must be byte-identical (deterministic
// canonical form). It exercises all eight fields including the anchor-point
// position key and a non-zero fallback ref.
func TestFR_012_ExtEnvelopeRoundTrip(t *testing.T) {
	orig := ExtEnvelope{
		Tok:           NewExtToken(0x00010002, 0x00030004),
		PayloadLength: 123456,
		PayloadRef:    fixtureUnitID(0x10),
		PayloadDigest: fixtureDigest(0x20),
		Disposition:   DispositionDegrade,
		FallbackRef:   fixtureUnitID(0x30),
		PositionKey: content.AnchorPoint{
			RunID:        fixtureUnitID(0x40),
			BirthOrdinal: 7,
			Side:         content.SideAfter,
			Boundary:     content.AnchorBoundary(0),
		},
	}

	enc, err := orig.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodeExtEnvelope(enc)
	if err != nil {
		t.Fatalf("DecodeExtEnvelope: %v", err)
	}
	if dec.Tok != orig.Tok {
		t.Errorf("Tok: got %x want %x", dec.Tok, orig.Tok)
	}
	if dec.PayloadLength != orig.PayloadLength {
		t.Errorf("PayloadLength: got %d want %d", dec.PayloadLength, orig.PayloadLength)
	}
	if dec.PayloadRef != orig.PayloadRef {
		t.Errorf("PayloadRef mismatch")
	}
	if dec.PayloadDigest != orig.PayloadDigest {
		t.Errorf("PayloadDigest mismatch")
	}
	if dec.Disposition != orig.Disposition {
		t.Errorf("Disposition: got %v want %v", dec.Disposition, orig.Disposition)
	}
	if dec.FallbackRef != orig.FallbackRef {
		t.Errorf("FallbackRef mismatch")
	}
	if dec.PositionKey != orig.PositionKey {
		t.Errorf("PositionKey: got %+v want %+v", dec.PositionKey, orig.PositionKey)
	}

	// Byte-exact deterministic canonical form: re-encoding the decoded value
	// reproduces the identical octets.
	reEnc, err := dec.Encode()
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(enc, reEnc) {
		t.Fatalf("round trip not byte-exact:\n first=%x\nsecond=%x", enc, reEnc)
	}

	// A zero16 fallback (ignore/degrade rules aside) round-trips too.
	orig2 := orig
	orig2.Disposition = DispositionRefuse
	orig2.FallbackRef = Zero16
	enc2, err := orig2.Encode()
	if err != nil {
		t.Fatalf("Encode zero-fallback: %v", err)
	}
	dec2, err := DecodeExtEnvelope(enc2)
	if err != nil {
		t.Fatalf("Decode zero-fallback: %v", err)
	}
	if dec2.FallbackRef != Zero16 {
		t.Errorf("zero16 fallback did not round-trip: %x", dec2.FallbackRef)
	}

	// A wrong discriminant is rejected.
	bad := append([]byte(nil), enc...)
	if len(bad) >= 3 && bad[2] == extEnvelopeDiscriminant {
		bad[2] = 0x05
		if _, err := DecodeExtEnvelope(bad); err == nil {
			t.Error("decode accepted a record with a non-0x06 discriminant")
		}
	}

	// A missing required field is rejected.
	// Re-encode with only the first two fields to simulate a truncated record.
	partial, _ := pdlfmt.EncodeRecord([]pdlfmt.Field{
		{Tag: extTagDiscriminant, Value: []byte{extEnvelopeDiscriminant}},
		{Tag: extTagTok, Value: orig.Tok[:]},
	})
	if _, err := DecodeExtEnvelope(partial); err == nil {
		t.Error("decode accepted a record missing required fields")
	}
}
