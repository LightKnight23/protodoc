package integrity

import (
	"errors"
	"testing"
)

// TestATTEST_EVID_CEILING_BOUNDARY is T-0184's named conformance test (corpus
// ATTEST-EVID-CEILING-BOUNDARY, FR-070). It ships paired at-limit / one-past
// fixtures for the ATTESTATION_EVIDENCE structural ceilings: the ae-der-octets
// maximum length (MAX_DECODED_UNIT), and the reserved ae-kind / ae-format
// boundary values, each accepted at the limit and rejected one value/octet
// past it, naming the exceeded ceiling (FR-106 abort-before-allocation).
func TestATTEST_EVID_CEILING_BOUNDARY(t *testing.T) {
	// --- ae-der-octets length ceiling (MAX_DECODED_UNIT) ---
	max := MaxAeDEROctets()
	if max == 0 {
		t.Fatal("MAX_DECODED_UNIT ceiling is 0")
	}
	// At the limit: accepted (checked before allocation, so no 256 MiB alloc).
	if err := CheckAeDEROctetsCeiling(max); err != nil {
		t.Errorf("ae-der-octets at the limit (%d) rejected: %v", max, err)
	}
	// One octet over: rejected naming the ceiling and observed value.
	if err := CheckAeDEROctetsCeiling(max + 1); !errors.Is(err, ErrAeDEROctetsOversized) {
		t.Errorf("ae-der-octets one over the limit: err = %v, want ErrAeDEROctetsOversized", err)
	}

	// --- reserved ae-kind boundary ---
	// The last defined kind (time-attestation, 0x02) is accepted in a valid
	// record; the first reserved kind (0x03) is rejected.
	atKindLimit := AttestationEvidence{
		ID: aeFixtureID(1), Kind: AeTimeAttestation, Format: AeFormatTimestamp,
		DEROctets: []byte{0x30, 0x00}, NestedCredChain: aeFixtureID(2), NestedRevocation: aeFixtureID(3),
	}
	if err := atKindLimit.Validate(); err != nil {
		t.Errorf("at-limit ae-kind (0x02) rejected: %v", err)
	}
	overKind := AttestationEvidence{ID: aeFixtureID(4), Kind: AeKind(0x03), Format: AeFormatX509Chain, DEROctets: []byte{0x30, 0x00}}
	if err := overKind.Validate(); !errors.Is(err, ErrAeKindReserved) && !errors.Is(err, ErrAeKindFormatPairing) {
		t.Errorf("one-past ae-kind (0x03): err = %v, want reserved/pairing rejection", err)
	}

	// --- reserved ae-format boundary (per kind) ---
	// credential-chain accepts exactly 0x00; the next value 0x01 is rejected.
	if err := ValidateKindFormatPairing(AeCredentialChain, AeFormatX509Chain); err != nil {
		t.Errorf("at-limit credential-chain format (0x00) rejected: %v", err)
	}
	if err := ValidateKindFormatPairing(AeCredentialChain, AeFormat(0x01)); !errors.Is(err, ErrAeKindFormatPairing) {
		t.Errorf("one-past credential-chain format (0x01): err = %v, want ErrAeKindFormatPairing", err)
	}
	// revocation-evidence accepts 0x01,0x02; 0x03 is one past.
	if err := ValidateKindFormatPairing(AeRevocationEvidence, AeFormatCRL); err != nil {
		t.Errorf("at-limit revocation format (CRL 0x02) rejected: %v", err)
	}
	if err := ValidateKindFormatPairing(AeRevocationEvidence, AeFormat(0x03)); !errors.Is(err, ErrAeKindFormatPairing) {
		t.Errorf("one-past revocation format (0x03): err = %v, want ErrAeKindFormatPairing", err)
	}
}
