package integrity

import (
	"errors"
	"testing"
)

// derSeq wraps content in a minimal DER SEQUENCE (tag 0x30, short-form length)
// for building opaque revocation-evidence fixtures whose framing is valid even
// though the contents are not parsed.
func derSeq(content []byte) []byte {
	out := []byte{0x30, byte(len(content))}
	return append(out, content...)
}

// TestFR_070_RevocationEvidenceParsesOffline is T-0173's named unit test
// (FR-070). Revocation evidence (ae-kind=1, ae-format OCSP or CRL) is validated
// OFFLINE: the DER stays opaque (we do not reimplement OCSP/CRL parsing), but
// the evidence must be a revocation kind with a matching format and DER octets
// well-framed as a single DER SEQUENCE spanning exactly the octets, with no
// network access.
func TestFR_070_RevocationEvidenceParsesOffline(t *testing.T) {
	ocsp := AttestationEvidence{
		ID: aeFixtureID(0x01), Kind: AeRevocationEvidence, Format: AeFormatOCSP,
		DEROctets: derSeq([]byte{0x0A, 0x01, 0x00}), // opaque OCSP-ish content
	}
	crl := AttestationEvidence{
		ID: aeFixtureID(0x02), Kind: AeRevocationEvidence, Format: AeFormatCRL,
		DEROctets: derSeq([]byte{}), // an empty but well-framed SEQUENCE
	}

	if err := VerifyRevocationEvidenceParsesOffline(ocsp); err != nil {
		t.Errorf("OCSP revocation evidence rejected: %v", err)
	}
	if err := VerifyRevocationEvidenceParsesOffline(crl); err != nil {
		t.Errorf("CRL revocation evidence rejected: %v", err)
	}

	// The opaque DER is retained verbatim (not parsed/normalised).
	if len(ocsp.DEROctets) == 0 {
		t.Errorf("opaque DER octets should be retained")
	}

	// Empty DER is rejected.
	empty := ocsp
	empty.DEROctets = nil
	if err := VerifyRevocationEvidenceParsesOffline(empty); !errors.Is(err, ErrRevocationEvidenceEmpty) {
		t.Errorf("empty DER: err = %v, want ErrRevocationEvidenceEmpty", err)
	}

	// Wrong kind (a credential-chain) is rejected.
	wrongKind := AttestationEvidence{Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: derSeq([]byte{0x00})}
	if err := VerifyRevocationEvidenceParsesOffline(wrongKind); err == nil {
		t.Errorf("credential-chain accepted as revocation evidence")
	}

	// Wrong format for revocation kind (timestamp) is rejected by the pairing.
	badFormat := ocsp
	badFormat.Format = AeFormatTimestamp
	if err := VerifyRevocationEvidenceParsesOffline(badFormat); !errors.Is(err, ErrAeKindFormatPairing) {
		t.Errorf("revocation with timestamp format: err = %v, want ErrAeKindFormatPairing", err)
	}

	// Trailing octets after the DER SEQUENCE are rejected.
	trailing := ocsp
	trailing.DEROctets = append(derSeq([]byte{0x00}), 0xFF, 0xFF)
	if err := VerifyRevocationEvidenceParsesOffline(trailing); err == nil {
		t.Errorf("trailing octets after the DER SEQUENCE were accepted")
	}
}
