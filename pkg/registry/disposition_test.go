package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_013_014_ExtDispositionStructuralReject is T-0054's named unit test
// (FR-013/FR-014). ext-disposition must be exactly one of 0x00/0x01/0x02; a
// missing or out-of-range value is a structural reject that NAMES the ext-tok,
// never a reader-chosen default.
func TestFR_013_014_ExtDispositionStructuralReject(t *testing.T) {
	tok := NewExtToken(0x00000abc, 0x0000def0)
	base := ExtEnvelope{
		Tok:           tok,
		PayloadLength: 1,
		PayloadRef:    fixtureUnitID(0x10),
		PayloadDigest: fixtureDigest(0x20),
		FallbackRef:   Zero16,
	}

	// (1) Each of the exactly-three legal values validates.
	for _, disp := range []ExtDisposition{DispositionIgnore, DispositionDegrade, DispositionRefuse} {
		e := base
		e.Disposition = disp
		if err := ValidateDisposition(e); err != nil {
			t.Errorf("legal disposition %v rejected: %v", disp, err)
		}
	}

	// (2) Every out-of-range value 0x03..0xFF is a structural reject naming
	// the ext-tok.
	for v := 3; v <= 0xFF; v++ {
		e := base
		e.Disposition = ExtDisposition(v)
		err := ValidateDisposition(e)
		if err == nil {
			t.Fatalf("out-of-range disposition 0x%02x was accepted", v)
		}
		if !errors.Is(err, ErrExtDispositionOutOfRange) {
			t.Fatalf("disposition 0x%02x: err = %v, want ErrExtDispositionOutOfRange", v, err)
		}
		var de *DispositionError
		if !errors.As(err, &de) {
			t.Fatalf("disposition 0x%02x: expected a *DispositionError, got %T", v, err)
		}
		if de.Tok != tok {
			t.Errorf("disposition 0x%02x: error names tok %x, want %x", v, de.Tok, tok)
		}
		if de.Value != uint8(v) {
			t.Errorf("disposition 0x%02x: error carries value 0x%02x", v, de.Value)
		}
	}

	// (3) A missing ext-disposition field is a structural reject at decode,
	// as a missing required field -- never defaulted.
	fields := []pdlfmt.Field{
		{Tag: extTagDiscriminant, Value: []byte{extEnvelopeDiscriminant}},
		{Tag: extTagTok, Value: tok[:]},
		{Tag: extTagPayloadLength, Value: pdlfmt.AppendVarint(nil, 1)},
		{Tag: extTagPayloadRef, Value: pdlfmt.AppendUnitID(nil, base.PayloadRef)},
		{Tag: extTagPayloadDigest, Value: pdlfmt.AppendDigest256(nil, base.PayloadDigest)},
		// tag 5 (ext-disposition) intentionally omitted
		{Tag: extTagFallbackRef, Value: pdlfmt.AppendUnitID(nil, Zero16)},
		{Tag: extTagPositionKey, Value: encodeZeroAnchor()},
	}
	rec, err := pdlfmt.EncodeRecord(fields)
	if err != nil {
		t.Fatalf("EncodeRecord: %v", err)
	}
	if _, err := DecodeExtEnvelope(rec); !errors.Is(err, ErrExtEnvelopeMissingField) {
		t.Fatalf("missing ext-disposition: decode err = %v, want ErrExtEnvelopeMissingField", err)
	}
}

// encodeZeroAnchor builds a valid 22-octet anchor-point value (all-zero
// run-id, ordinal 0, side/boundary 0) for use as an ext-position-key in
// fixtures.
func encodeZeroAnchor() []byte {
	b := make([]byte, 16+4)   // run-id + ordinal
	b = append(b, 0x00, 0x00) // side=0, boundary=0 (both valid)
	return b
}
