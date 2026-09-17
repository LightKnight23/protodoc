package integrity

import (
	"errors"
	"testing"
)

// TestFR_070_AeKindFormatPairingRejectsInvalid is T-0167's named unit test
// (FR-070). The (ae-kind, ae-format) pairing is a closed set:
// credential-chain<->X.509 chain, revocation-evidence<->{OCSP,CRL},
// time-attestation<->TimeStampToken. Every valid pairing is accepted; every
// other pairing, and any reserved ae-kind, is rejected.
func TestFR_070_AeKindFormatPairingRejectsInvalid(t *testing.T) {
	// Valid pairings accepted.
	valid := []struct {
		kind   AeKind
		format AeFormat
	}{
		{AeCredentialChain, AeFormatX509Chain},
		{AeRevocationEvidence, AeFormatOCSP},
		{AeRevocationEvidence, AeFormatCRL},
		{AeTimeAttestation, AeFormatTimestamp},
	}
	for _, v := range valid {
		if err := ValidateKindFormatPairing(v.kind, v.format); err != nil {
			t.Errorf("valid pairing (%v, 0x%02x) rejected: %v", v.kind, uint8(v.format), err)
		}
	}

	// Invalid pairings rejected.
	invalid := []struct {
		name   string
		kind   AeKind
		format AeFormat
	}{
		{"credential-chain with OCSP", AeCredentialChain, AeFormatOCSP},
		{"credential-chain with timestamp", AeCredentialChain, AeFormatTimestamp},
		{"revocation-evidence with X509", AeRevocationEvidence, AeFormatX509Chain},
		{"revocation-evidence with timestamp", AeRevocationEvidence, AeFormatTimestamp},
		{"time-attestation with X509", AeTimeAttestation, AeFormatX509Chain},
		{"time-attestation with OCSP", AeTimeAttestation, AeFormatOCSP},
		{"credential-chain with reserved format", AeCredentialChain, AeFormat(0x7F)},
	}
	for _, c := range invalid {
		if err := ValidateKindFormatPairing(c.kind, c.format); !errors.Is(err, ErrAeKindFormatPairing) {
			t.Errorf("%s: err = %v, want ErrAeKindFormatPairing", c.name, err)
		}
	}

	// A reserved ae-kind is rejected as reserved.
	for _, k := range []AeKind{0x03, 0x7F, 0xFF} {
		if err := ValidateKindFormatPairing(k, AeFormatX509Chain); !errors.Is(err, ErrAeKindReserved) {
			t.Errorf("reserved ae-kind 0x%02x: err = %v, want ErrAeKindReserved", uint8(k), err)
		}
	}
}
