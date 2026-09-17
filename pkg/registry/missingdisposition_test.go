package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_014_MissingDispositionRejectedNamingToken is T-0063's named
// conformance test (FR-014). A corpus of extension-envelope records each
// missing the ext-disposition field (with the ext-tok present) must be
// rejected as a structural error that NAMES the ext-tok -- never defaulted to
// a reader-chosen disposition.
func TestFR_014_MissingDispositionRejectedNamingToken(t *testing.T) {
	// Build a record with every field EXCEPT ext-disposition (tag 5), for a
	// given tok.
	buildMissingDisposition := func(tok ExtToken) []byte {
		fields := []pdlfmt.Field{
			{Tag: extTagDiscriminant, Value: []byte{extEnvelopeDiscriminant}},
			{Tag: extTagTok, Value: tok[:]},
			{Tag: extTagPayloadLength, Value: pdlfmt.AppendVarint(nil, 7)},
			{Tag: extTagPayloadRef, Value: pdlfmt.AppendUnitID(nil, fixtureUnitID(0x10))},
			{Tag: extTagPayloadDigest, Value: pdlfmt.AppendDigest256(nil, fixtureDigest(0x20))},
			// tag 5 (ext-disposition) omitted
			{Tag: extTagFallbackRef, Value: pdlfmt.AppendUnitID(nil, Zero16)},
			{Tag: extTagPositionKey, Value: encodeZeroAnchor()},
		}
		rec, err := pdlfmt.EncodeRecord(fields)
		if err != nil {
			t.Fatalf("EncodeRecord: %v", err)
		}
		return rec
	}

	corpusToks := []ExtToken{
		NewExtToken(0x00000001, 0x00000001),
		NewExtToken(0x40000000, 0x00009999),
		NewExtToken(0x80000000, 0xFFFFFFFE),
	}
	for _, tok := range corpusToks {
		rec := buildMissingDisposition(tok)
		_, err := DecodeExtEnvelope(rec)
		if !errors.Is(err, ErrExtEnvelopeMissingField) {
			t.Fatalf("tok %x: err = %v, want ErrExtEnvelopeMissingField", tok, err)
		}
		var me *MissingFieldError
		if !errors.As(err, &me) {
			t.Fatalf("tok %x: expected *MissingFieldError, got %T", tok, err)
		}
		if me.Tag != extTagDisposition {
			t.Errorf("tok %x: missing tag = %d, want %d (ext-disposition)", tok, me.Tag, extTagDisposition)
		}
		if !me.TokKnown || me.Tok != tok {
			t.Errorf("tok %x: reject must name the ext-tok, got TokKnown=%v Tok=%x", tok, me.TokKnown, me.Tok)
		}
	}
}
