package integrity

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_071_NestedChainMandatoryForTimeAttestation is T-0170's named unit
// test (FR-071). A time-attestation MUST carry both the attesting authority's
// own nested credential chain AND its nested revocation evidence (both
// non-zero16), so the TSA's signature is checkable offline as strictly as the
// primary signatory's. A credential-chain or revocation-evidence record MUST
// carry neither nested ref (no second level of nesting).
func TestFR_071_NestedChainMandatoryForTimeAttestation(t *testing.T) {
	cred := aeFixtureID(0x40)
	rev := aeFixtureID(0x50)
	zero := pdlfmt.UnitID{}

	// A time-attestation with both nested refs is valid.
	okTime := AttestationEvidence{
		ID: aeFixtureID(0x01), Kind: AeTimeAttestation, Format: AeFormatTimestamp,
		NestedCredChain: cred, NestedRevocation: rev,
	}
	if err := okTime.ValidateNestedRefs(); err != nil {
		t.Errorf("time-attestation with both nested refs rejected: %v", err)
	}

	// Missing either nested ref is rejected.
	missCred := okTime
	missCred.NestedCredChain = zero
	if err := missCred.ValidateNestedRefs(); !errors.Is(err, ErrAeNestedRule) {
		t.Errorf("time-attestation missing nested cred chain: err = %v, want ErrAeNestedRule", err)
	}
	missRev := okTime
	missRev.NestedRevocation = zero
	if err := missRev.ValidateNestedRefs(); !errors.Is(err, ErrAeNestedRule) {
		t.Errorf("time-attestation missing nested revocation: err = %v, want ErrAeNestedRule", err)
	}
	missBoth := okTime
	missBoth.NestedCredChain = zero
	missBoth.NestedRevocation = zero
	if err := missBoth.ValidateNestedRefs(); !errors.Is(err, ErrAeNestedRule) {
		t.Errorf("time-attestation missing both nested refs: err = %v, want ErrAeNestedRule", err)
	}

	// A credential-chain / revocation-evidence record MUST carry no nested ref.
	credRec := AttestationEvidence{ID: aeFixtureID(0x02), Kind: AeCredentialChain, Format: AeFormatX509Chain}
	if err := credRec.ValidateNestedRefs(); err != nil {
		t.Errorf("credential-chain with no nested refs rejected: %v", err)
	}
	credWithNested := credRec
	credWithNested.NestedCredChain = cred
	if err := credWithNested.ValidateNestedRefs(); !errors.Is(err, ErrAeNestedRule) {
		t.Errorf("credential-chain carrying a nested ref: err = %v, want ErrAeNestedRule", err)
	}
	revRec := AttestationEvidence{ID: aeFixtureID(0x03), Kind: AeRevocationEvidence, Format: AeFormatCRL}
	revWithNested := revRec
	revWithNested.NestedRevocation = rev
	if err := revWithNested.ValidateNestedRefs(); !errors.Is(err, ErrAeNestedRule) {
		t.Errorf("revocation-evidence carrying a nested ref: err = %v, want ErrAeNestedRule", err)
	}

	// A full decode of a time-attestation missing a nested ref is rejected
	// (the decoder runs Validate). Hand-build the wire record.
	fields := []pdlfmt.Field{
		{Tag: aeTagDiscriminant, Value: []byte{attestationEvidenceDiscriminant}},
		{Tag: aeTagID, Value: pdlfmt.AppendUnitID(nil, aeFixtureID(0x05))},
		{Tag: aeTagKind, Value: []byte{byte(AeTimeAttestation)}},
		{Tag: aeTagFormat, Value: []byte{byte(AeFormatTimestamp)}},
		{Tag: aeTagDEROctets, Value: []byte{0x30, 0x00}},
		{Tag: aeTagNestedCredChain, Value: pdlfmt.AppendUnitID(nil, zero)},
		{Tag: aeTagNestedRevocation, Value: pdlfmt.AppendUnitID(nil, zero)},
	}
	rec, _ := pdlfmt.EncodeRecord(fields)
	if _, err := DecodeAttestationEvidence(rec); !errors.Is(err, ErrAeNestedRule) {
		t.Errorf("decode of time-attestation missing nested refs: err = %v, want ErrAeNestedRule", err)
	}
}
