package integrity

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

func sigFixtureUnitID(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

// TestFR_063_SignatureRecordRoundTrip is T-0153's named unit test
// (FR-063/FR-064). The SIGNATURE record (integrity.abnf S5) encodes and
// decodes byte-exact with every field surviving, rejects a wrong discriminant
// and a missing field, and enforces the non-zero16 rule for
// sig-cred-chain-ref and sig-time-attestation-ref.
func TestFR_063_SignatureRecordRoundTrip(t *testing.T) {
	var val [SigValueLen]byte
	for i := range val {
		val[i] = byte(i)
	}
	orig := SignatureRecord{
		ParamSet:     0x0001,
		SignedObject: Digest{0xAA},
		Value:        val,
		Coverage: CoverageDescriptor{
			Mode:      CoverageModeSubset,
			Covered:   []SegmentRange{{Start: 0, End: 3}},
			Uncovered: []SegmentRange{{Start: 3, End: 8}},
			Bitmask:   CoverageBitHeader,
		},
		CredChainRef:       sigFixtureUnitID(0x10),
		RevocationRef:      pdlfmt.UnitID{}, // zero16 allowed here
		TimeAttestationRef: sigFixtureUnitID(0x20),
		Intent:             0x02,
		PresentationRef:    sigFixtureUnitID(0x30),
	}

	enc, err := orig.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodeSignatureRecord(enc)
	if err != nil {
		t.Fatalf("DecodeSignatureRecord: %v", err)
	}
	if dec.ParamSet != orig.ParamSet || dec.SignedObject != orig.SignedObject ||
		dec.Value != orig.Value || dec.CredChainRef != orig.CredChainRef ||
		dec.RevocationRef != orig.RevocationRef || dec.TimeAttestationRef != orig.TimeAttestationRef ||
		dec.Intent != orig.Intent || dec.PresentationRef != orig.PresentationRef {
		t.Errorf("scalar/ref fields did not survive round trip")
	}
	if dec.Coverage.Mode != orig.Coverage.Mode || dec.Coverage.Bitmask != orig.Coverage.Bitmask ||
		!rangesEqual(dec.Coverage.Covered, orig.Coverage.Covered) || !rangesEqual(dec.Coverage.Uncovered, orig.Coverage.Uncovered) {
		t.Errorf("coverage descriptor did not survive round trip")
	}
	reEnc, err := dec.Encode()
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(enc, reEnc) {
		t.Fatalf("round trip not byte-exact")
	}

	// Wrong discriminant rejected.
	bad := append([]byte(nil), enc...)
	if len(bad) >= 3 && bad[2] == signatureDiscriminant {
		bad[2] = 0x41
		if _, err := DecodeSignatureRecord(bad); !errors.Is(err, ErrSignatureDiscriminant) {
			t.Errorf("wrong discriminant: err = %v, want ErrSignatureDiscriminant", err)
		}
	}

	// A zero16 sig-cred-chain-ref is rejected (encode and decode).
	badRef := orig
	badRef.CredChainRef = pdlfmt.UnitID{}
	if _, err := badRef.Encode(); !errors.Is(err, ErrSignatureZeroRef) {
		t.Errorf("encode zero cred-chain-ref: err = %v, want ErrSignatureZeroRef", err)
	}

	// A missing required field (omit sig-intent, tag 8) is rejected.
	var ps [2]byte
	ps[1] = 1
	cov, _ := orig.Coverage.Encode(nil)
	missing, _ := pdlfmt.EncodeRecord([]pdlfmt.Field{
		{Tag: sigTagDiscriminant, Value: []byte{signatureDiscriminant}},
		{Tag: sigTagParamSet, Value: ps[:]},
		{Tag: sigTagSignedObject, Value: pdlfmt.AppendDigest256(nil, pdlfmt.Digest256(orig.SignedObject))},
		{Tag: sigTagValue, Value: val[:]},
		{Tag: sigTagCoverage, Value: cov},
		{Tag: sigTagCredChainRef, Value: pdlfmt.AppendUnitID(nil, orig.CredChainRef)},
		{Tag: sigTagRevocationRef, Value: pdlfmt.AppendUnitID(nil, orig.RevocationRef)},
		{Tag: sigTagTimeAttestRef, Value: pdlfmt.AppendUnitID(nil, orig.TimeAttestationRef)},
		// tag 8 (sig-intent) omitted
		{Tag: sigTagPresentationRef, Value: pdlfmt.AppendUnitID(nil, orig.PresentationRef)},
	})
	if _, err := DecodeSignatureRecord(missing); !errors.Is(err, ErrSignatureMissingField) {
		t.Errorf("missing sig-intent: err = %v, want ErrSignatureMissingField", err)
	}
}
